package main

import (
	"bufio"
	gcontext "context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os/exec"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// --------------------------------------------------------------------------------
type msgClientCommand struct {
	cmd string
}

type evtPluginConn struct {
	c net.Conn
}

type evtPluginCmdExit struct {
	err error
}

type evtPluginBind struct {
	isUnbind bool
}

type PluginStatus int

const (
	Idle      PluginStatus = 0
	Launching PluginStatus = 1
	Launched  PluginStatus = 2
)

type PluginActor struct {
	name      string
	config    *PluginConfig
	cancel    gcontext.CancelFunc
	c         net.Conn
	pluginCmd *exec.Cmd
	bindNum   int
	port      int
}

func (pa *PluginActor) getCmdStatus() PluginStatus {
	switch {
	case pa.pluginCmd == nil && pa.c != nil:
		log.Panicln("actor status error")
	case pa.pluginCmd == nil && pa.c == nil:
		return Idle
	case pa.pluginCmd != nil && pa.c == nil:
		return Launching
	case pa.pluginCmd != nil && pa.c != nil:
		return Launched
	}
	return Idle
}

func (pa *PluginActor) launchCmd(context actor.Context) {
	if pa.getCmdStatus() != Idle {
		return
	}

	cmd := exec.Command(pa.config.binPath, pa.name, fmt.Sprint(pa.port))

	binRunner := func() {
		err := cmd.Run()
		log.Printf("actor %s process exit %v", pa.name, err)
		context.Request(context.Self(), &evtPluginCmdExit{err})
	}

	go binRunner()

	pa.pluginCmd = cmd
}

func (pa *PluginActor) terminateCmd(context actor.Context) {
	if pa.getCmdStatus() == Idle {
		return
	}

	if pa.c != nil {
		pa.c.Close()
		pa.c = nil
	} else {
		pa.pluginCmd.Process.Kill()
	}

	pa.pluginCmd.Wait()
	pa.pluginCmd = nil
}

func (pa *PluginActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		pa.bindNum = 0
		actorRegistery.Store("plugin:"+pa.name, context.Self())
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "started"})
	case *actor.Stopping:
		pa.terminateCmd(context)
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "stopping"})
	case *actor.Stopped:
		actorRegistery.Delete("plugin:" + pa.name)
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "stopped"})
	case *actor.Restarting:
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "restarting"})
	case *msgClientCommand:
		cmdStr, err := url.QueryUnescape(msg.cmd)
		panicOnErr(err)
		log.Println("recv cmdStr:", cmdStr, pa.c)

		if pa.pluginCmd == nil {
			eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "err", msg: "plugin cmd not start"})
			return
		}

		if pa.c == nil {
			eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "err", msg: "session not connected"})
			return
		}

		pa.c.Write([]byte(cmdStr + "\n"))
	case *evtPluginConn:
		if pa.c != nil {
			pa.c.Close()
			pa.c = nil
		}
		pa.c = msg.c
		if pa.c != nil {
			log.Println("-----------")
			go func() {
				defer func() {
					context.Request(context.Self(), &evtPluginConn{})
				}()
				eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "got connection"})
				br := bufio.NewReader(pa.c)
				for line, _, err := br.ReadLine(); err == nil; line, _, err = br.ReadLine() {
					str := string(line)
					log.Println("recv str from plugin: ", str)
					context.Request(context.Self(), str)
				}
			}()
			return
		}
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "actor", msg: "lose connection"})
	case *evtPluginCmdExit:
		pa.terminateCmd(context)
		if pa.bindNum > 0 {
			pa.launchCmd(context)
		}

	case *evtPluginBind:
		if msg.isUnbind {
			pa.bindNum--
		} else {
			pa.bindNum++
		}
		switch {
		case pa.bindNum > 0 && pa.getCmdStatus() == Idle:
			pa.launchCmd(context)
		case pa.bindNum == 0 && pa.getCmdStatus() != Idle:
			pa.terminateCmd(context)
		case pa.bindNum > 0 && pa.getCmdStatus() != Idle:
		default:
			log.Fatalln("state error bindNum:", pa.bindNum)
		}

	case string:
		eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: "notify", msg: url.QueryEscape(msg)})
	}
}
