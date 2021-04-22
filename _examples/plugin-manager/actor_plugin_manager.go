package main

import (
	"bufio"
	gcontext "context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
)

type PluginConfig struct {
}

var pluginConfigs = map[string]*PluginConfig{
	"caster":    {},
	"wireguard": {},
}

type PluginSessionActor struct {
}

func (ps *PluginSessionActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		_ = msg
	case *actor.Stopping:
	case *actor.Stopped:
	case *actor.Restarting:
	}
}

type msgClientCommand struct {
	cmd string
}

type msgPluginActor struct {
	c net.Conn
}

type PluginActor struct {
	name   string
	config *PluginConfig
	cancel gcontext.CancelFunc
	c      net.Conn
}

func (p *PluginActor) fakeSession(context actor.Context, cmdStr string) {
	args := strings.SplitN(cmdStr, " ", 2)
	switch args[0] {
	case "start":
		if p.cancel != nil {
			eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "notify", msg: "log:" + url.QueryEscape((cmdStr))})
			return
		}
		ctx, cancel := gcontext.WithCancel(gcontext.TODO())
		p.cancel = cancel
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			cnt := 0
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cnt++
					eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "notify", msg: fmt.Sprint("tick:", cnt)})
				}
			}
		}()
	case "stop":
		if p.cancel != nil {
			p.cancel()
			p.cancel = nil
		}
	}
}

func (p *PluginActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		actorRegistery.Store("plugin:"+p.name, context.Self())
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "started"})
	case *actor.Stopping:
		if p.c != nil {
			p.c.Close()
			p.c = nil
		}
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "stopping"})
	case *actor.Stopped:
		actorRegistery.Delete("plugin:" + p.name)
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "stopped"})
	case *actor.Restarting:
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "restarting"})
	case *msgClientCommand:
		cmdStr, err := url.QueryUnescape(msg.cmd)
		panicOnErr(err)
		log.Println("recv cmdStr:", cmdStr, p.c)
		if p.c == nil {
			log.Println("p.c == nil")
			eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "err", msg: "session not connected"})
			return
		}
		p.c.Write([]byte(cmdStr + "\n"))
		// eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "notify", msg: "log:" + url.QueryEscape((cmdStr))})
		// p.fakeSession(context, cmdStr)
	case *msgPluginActor:
		log.Println("recv Conn:", msg)
		if p.c != nil {
			p.c.Close()
			p.c = nil
		}
		p.c = msg.c
		if p.c != nil {
			log.Println("-----------")
			go func() {
				defer func() {
					context.Request(context.Self(), &msgPluginActor{})
				}()
				eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "got connection"})
				br := bufio.NewReader(p.c)
				for line, _, err := br.ReadLine(); err == nil; line, _, err = br.ReadLine() {
					str := string(line)
					log.Println("recv str from plugin: ", str)
					context.Send(context.Self(), str)
				}
			}()
			return
		}
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "actor", msg: "lose connection"})
	case string:
		eventStream.Publish(&msgPlugin{pluginName: p.name, msgType: "notify", msg: url.QueryEscape(msg)})
	}
}

type PluginManagerActor struct {
	pluginPIDs map[string]*actor.PID
	l          net.Listener
}

func (pma *PluginManagerActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		_ = msg
		for name, config := range pluginConfigs {
			props := actor.PropsFromProducer(func() actor.Actor { return &PluginActor{name: name, config: config, cancel: nil} }).WithReceiverMiddleware(middleware.Logger)
			pid, err := context.SpawnNamed(props, "Plugin:"+name)
			panicOnErr(err)
			pma.pluginPIDs[name] = pid
		}

		l, e := net.Listen("tcp4", "localhost:8888")
		panicOnErr(e)
		go func() {
			defer context.Poison(context.Self())
			for {
				c, e := pma.l.Accept()
				if e != nil {
					return
				}
				br := bufio.NewReader(c)
				line, _, err := br.ReadLine()
				if err != nil {
					c.Close()
					return
				}
				pluginName := strings.TrimSpace(string(line))
				pid, ok := pma.pluginPIDs[pluginName]
				if !ok {
					c.Close()
					return
				}

				context.Request(pid, &msgPluginActor{c: c})
			}
		}()
		pma.l = l
	case *actor.Stopping:
		if pma.l != nil {
			pma.l.Close()
		}
	case *actor.Stopped:
	case *actor.Restarting:
		context.Poison(context.Self())
	}
}
