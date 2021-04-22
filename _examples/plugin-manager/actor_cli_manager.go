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

/*
协议
c->s
	init##PLUGINNAME: 初始化插件
	close: 啥也不做
	exit: 销毁插件
	cmd##CMD|ARG1...: 发给插件的命令
	quit: 断开连接
s->c
	error:1|2|3|OTHERCODE
	data:localip:LOCALIP|roomname:ROOMNAME|tvip:TVIP
	state:STARTING|
	log:
	cmd: 给插件管理器执行的命令

改进
c->s
	bind caster
	caster CMD
	unbind caster
s->c
	caster bind ok
	caster notify state:initing
	caster notify state:started
	caster notify date:....
	caster notify log:....
*/

// type msgWebSocket struct {
// 	msg string
// }

type CliSessionActor struct {
	sub   *eventstream.Subscription
	binds map[string]bool
	c     net.Conn
}

func (csa *CliSessionActor) dealCliMessage(context actor.Context, msg string) {
	args := strings.SplitN(msg, " ", 2)
	log.Println(args)
	for _, v := range args {
		log.Println([]byte(v), len(v))
	}
	cmd := args[0]
	switch cmd {
	case "b":
		_, ok := actorRegistery.Load("plugin:" + args[1])
		if !ok {
			csa.c.Write([]byte("error plugin:" + args[1] + " not exist\n"))
			return
		}
		csa.binds[args[1]] = true
		csa.c.Write([]byte("plugin:" + args[1] + " bound\n"))
	case "u":
		delete(csa.binds, args[1])
		csa.c.Write([]byte(("plugin:" + args[1] + " unbound\n")))
	default:
		val, ok := actorRegistery.Load("plugin:" + cmd)
		if !ok {
			csa.c.Write([]byte("error plugin:" + cmd + " not exist\n"))
			return
		}
		if _, ok := csa.binds[cmd]; !ok {
			csa.c.Write([]byte("error plugin:" + cmd + " not bound\n"))
			return
		}
		pid := val.(*actor.PID)
		context.Request(pid, &msgClientCommand{cmd: args[1]})
	}
}

func (csa *CliSessionActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// csa.bw = bufio.NewWriter(csa.c)
		csa.sub = eventStream.Subscribe(func(evt interface{}) {
			log.Println("got evt:", evt)
			switch msg := evt.(type) {
			case *msgPlugin:
				// log.Println("1csa subscribe:", msg.msg, msg.pluginName, msg.msgType)
				// ws.send
				if _, ok := csa.binds[msg.pluginName]; !ok {
					return
				}
				// log.Println("2csa subscribe:", msg.msg, msg.pluginName, msg.msgType)
				csa.c.Write([]byte(msg.pluginName + " " + msg.msgType + " " + msg.msg + "\n"))
				// switch msg.msgType {
				// case "notify":
				// 	// log.Println("3csa subscribe:", msg.msg, msg.pluginName, msg.msgType)
				// 	csa.c.Write([]byte(msg.pluginName + " " + msg.msgType + " " + msg.msg + "\n"))
				// }
			}
		})

		go func() {
			defer context.Poison(context.Self())
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
		eventStream.Unsubscribe(csa.sub)
	case string:
		csa.dealCliMessage(context, msg)
	}
}

type CliManagerActor struct {
	l net.Listener
	w sync.WaitGroup
}

func (cma *CliManagerActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Started:
		l, e := net.Listen("tcp4", "localhost:9999")
		panicOnErr(e)
		cma.w.Add(1)
		go func() {
			defer context.Poison(context.Self())
			defer cma.w.Done()
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
