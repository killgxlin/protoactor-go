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
	conn net.Conn
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
	conn      net.Conn
	pluginCmd *exec.Cmd
	bindNum   int
	port      int
}

func (pa *PluginActor) getCmdStatus() PluginStatus {
	switch {
	case pa.pluginCmd == nil && pa.conn != nil:
		log.Panicln("actor status error")
	case pa.pluginCmd == nil && pa.conn == nil:
		return Idle
	case pa.pluginCmd != nil && pa.conn == nil:
		return Launching
	case pa.pluginCmd != nil && pa.conn != nil:
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
	pa.publish(Binary, "process launched")
}

func (pa *PluginActor) terminateCmd(context actor.Context) {
	if pa.getCmdStatus() == Idle {
		return
	}

	if pa.conn != nil {
		pa.conn.Close()
		pa.conn = nil
	} else {
		pa.pluginCmd.Process.Kill()
	}

	pa.pluginCmd.Wait()
	pa.pluginCmd = nil
	pa.publish(Binary, "process terminated")
}

func (pa *PluginActor) publish(msgType MsgType, msg string) {
	eventStream.Publish(&msgPlugin{pluginName: pa.name, msgType: msgType, msg: msg})
}

func (pa *PluginActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		pa.bindNum = 0
		actorRegistery.Store("plugin:"+pa.name, context.Self())
		pa.publish(Actor, "started")
	case *actor.Stopping:
		pa.terminateCmd(context)
		pa.publish(Actor, "stopping")
	case *actor.Stopped:
		actorRegistery.Delete("plugin:" + pa.name)
		pa.publish(Actor, "stopped")
	case *actor.Restarting:
		pa.publish(Actor, "restarting")
	case *msgClientCommand:
		cmdStr, err := url.QueryUnescape(msg.cmd)
		panicOnErr(err)
		log.Println("recv cmdStr:", cmdStr, pa.conn)

		if pa.getCmdStatus() != Launched {
			pa.publish(Err, "plugin cmd not launched")
			return
		}

		pa.conn.Write([]byte(cmdStr + "\n"))
	case *evtPluginConn:
		if pa.conn != nil {
			pa.conn.Close()
			pa.conn = nil
		}
		pa.conn = msg.conn
		if pa.conn != nil {
			go func() {
				defer func() {
					context.Request(context.Self(), &evtPluginConn{})
				}()
				pa.publish(Actor, "got connection")
				br := bufio.NewReader(pa.conn)
				for line, _, err := br.ReadLine(); err == nil; line, _, err = br.ReadLine() {
					str := string(line)
					log.Println("recv str from plugin: ", str)
					context.Request(context.Self(), str)
				}
			}()
			return
		}
		pa.publish(Actor, "lose connection")
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
		pa.publish(Plugin, url.QueryEscape(msg))
	}
}
