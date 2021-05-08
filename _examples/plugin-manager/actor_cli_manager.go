package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
	"github.com/AsynkronIT/protoactor-go/eventstream"
)

// --------------------------------------------------------------------------------
type CliSessionActor struct {
	sub       *eventstream.Subscription
	binds     map[string]bool
	c         net.Conn
	curPlugin string
}

func (csa *CliSessionActor) sendToClient(msg string) {
	if csa.c == nil {
		log.Panicln("cli conn not exist")
	}
	csa.c.Write([]byte(msg + "\n"))
}

func (csa *CliSessionActor) bind(context actor.Context, pluginName string) {
	val, ok := actorRegistery.Load("plugin:" + pluginName)
	if !ok {
		csa.sendToClient("error plugin:" + pluginName + " not exist")
		return
	}

	csa.binds[pluginName] = true
	csa.sendToClient("plugin:" + pluginName + " bound")

	pid := val.(*actor.PID)
	context.Request(pid, &evtPluginBind{false})
	csa.curPlugin = pluginName
}

func (csa *CliSessionActor) unbind(context actor.Context, pluginName string) {
	val, ok := actorRegistery.Load("plugin:" + pluginName)
	if !ok {
		csa.sendToClient("error plugin:" + pluginName + " not exist")
		return
	}

	if _, ok := csa.binds[pluginName]; !ok {
		csa.sendToClient("error plugin:" + pluginName + " not bound")
		return
	}

	delete(csa.binds, pluginName)
	csa.sendToClient("plugin:" + pluginName + " unbound")

	pid := val.(*actor.PID)
	context.Request(pid, &evtPluginBind{true})
	if csa.curPlugin == pluginName {
		csa.curPlugin = ""
	}
}

func (csa *CliSessionActor) unbindAll(context actor.Context) {
	for pluginName := range csa.binds {
		csa.unbind(context, pluginName)
	}
	csa.curPlugin = ""
}

func (csa *CliSessionActor) sendToPlugin(context actor.Context, pluginName, msg string) {
	val, ok := actorRegistery.Load("plugin:" + pluginName)
	if !ok {
		csa.sendToClient("error plugin:" + pluginName + " not exist")
		return
	}
	if _, ok := csa.binds[pluginName]; !ok {
		csa.sendToClient("error plugin:" + pluginName + " not bound")
		return
	}
	pid := val.(*actor.PID)
	context.Request(pid, &msgClientCommand{cmd: msg})
}

// 处理客户端消息
func (csa *CliSessionActor) dealCliMessage(context actor.Context, msg string) {
	args := strings.SplitN(msg, " ", 2)
	log.Println(args)
	for _, v := range args {
		log.Println([]byte(v), len(v))
	}
	cmd := args[0]
	switch cmd {
	case "b", "bind":
		pluginName := csa.curPlugin
		if len(args) > 1 {
			pluginName = args[1]
		}
		csa.bind(context, pluginName)
	case "u", "unbind":
		pluginName := csa.curPlugin
		if len(args) > 1 {
			pluginName = args[1]
		}
		csa.unbind(context, pluginName)
	case "ch", "change":
		if _, ok := csa.binds[args[1]]; !ok {
			csa.sendToClient("plugin " + args[1] + " not bound")
			return
		}
		csa.curPlugin = args[1]
		csa.sendToClient("change current plugin to " + args[1])
	case "c", "current":
		if csa.curPlugin == "" {
			csa.sendToClient("current plugin not set")
			return
		}
		csa.sendToPlugin(context, csa.curPlugin, args[1])
	default:
		csa.sendToPlugin(context, cmd, args[1])
	}
}

func (csa *CliSessionActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		csa.curPlugin = ""
		// 订阅消息
		csa.sub = eventStream.Subscribe(func(evt interface{}) {
			log.Println("got evt:", evt)
			switch msg := evt.(type) {
			case *msgPlugin: // 只处理插件相关消息
				if _, ok := csa.binds[msg.pluginName]; !ok {
					return
				}
				csa.sendToClient(msg.pluginName + " " + string(msg.msgType) + " " + msg.msg)
			}
		})

		// 读取客户端消息
		go func() {
			defer context.Poison(context.Self()) // 把自己关闭

			br := bufio.NewReader(csa.c)

			for line, _, err := br.ReadLine(); err == nil; line, _, err = br.ReadLine() {
				str := string(line)
				log.Println("recv str from cli: ", str)
				context.Send(context.Self(), str)
			}
		}()

	case *actor.Stopping:
		csa.c.Close()
	case *actor.Restarting:
		csa.c.Close()
	case *actor.Stopped:
		csa.unbindAll(context)
		eventStream.Unsubscribe(csa.sub)
	case string:
		csa.dealCliMessage(context, msg)
	}
}

// --------------------------------------------------------------------------------
type CliManagerActor struct {
	l net.Listener
	w sync.WaitGroup
}

func (cma *CliManagerActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Started:
		addr, err := findAvailableAddr(8000, 8999)
		panicOnErr(err)
		log.Println("found available addr for cma", addr)
		l, e := Listen("cli", addr)
		panicOnErr(e)
		cma.w.Add(1)
		go func() {
			defer func() {
				context.Poison(context.Self())
				cma.w.Done()
			}()
			for {
				c, e := cma.l.Accept()

				if e != nil {
					return
				}
				cma.w.Add(1)
				props := actor.PropsFromProducer(func() actor.Actor { return &CliSessionActor{c: c, binds: map[string]bool{}} }).WithReceiverMiddleware(middleware.Logger)
				context.SpawnPrefix(props, "CliSessionActor")
			}
		}()
		cma.l = l
	case *actor.Terminated:
		cma.w.Done()
	case *actor.Stopping:
		cma.l.Close()
	case *actor.Stopped:
		cma.w.Wait()
	case *actor.Restarting:
		cma.l.Close()
	}
}
