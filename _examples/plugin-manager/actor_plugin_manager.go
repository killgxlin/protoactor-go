package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
)

// --------------------------------------------------------------------------------
type PluginConfig struct {
	binPath string
}

var pluginConfigs = map[string]*PluginConfig{
	"caster": {
		binPath: "./plugins/plugin.exe",
	},
	"wireguard": {
		binPath: "./plugins/plugin.exe",
	},
}

// --------------------------------------------------------------------------------
type PluginManagerActor struct {
	pluginPIDs map[string]*actor.PID
	l          net.Listener
}

func (pma *PluginManagerActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Started:
		addr, err := findAvailableAddr(9000, 9999)
		panicOnErr(err)
		log.Println("found available addr for pma", addr)
		l, e := net.Listen("tcp4", addr)
		panicOnErr(e)

		port, _ := strconv.ParseInt(strings.Split(addr, ":")[1], 10, 64)

		for name, config := range pluginConfigs {
			props := actor.PropsFromProducer(func() actor.Actor { return &PluginActor{name: name, config: config, cancel: nil, port: int(port)} }).WithReceiverMiddleware(middleware.Logger)
			pid, err := context.SpawnNamed(props, "Plugin:"+name)
			panicOnErr(err)
			pma.pluginPIDs[name] = pid
		}

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

				context.Request(pid, &evtPluginConn{c: c})
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
