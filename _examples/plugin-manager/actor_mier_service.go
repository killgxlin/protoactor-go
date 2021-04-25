package main

import (
	"fmt"
	"log"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
)

// --------------------------------------------------------------------------------
type MierServiceActor struct {
}

func (msa *MierServiceActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		{
			props := actor.PropsFromProducer(func() actor.Actor { return &CliManagerActor{} }).WithReceiverMiddleware(middleware.Logger)
			_, err := context.SpawnNamed(props, "CliManagerActor")
			panicOnErr(err)
		}

		{
			props := actor.PropsFromProducer(func() actor.Actor { return &PluginManagerActor{pluginPIDs: map[string]*actor.PID{}} }).WithReceiverMiddleware(middleware.Logger)
			_, err := context.SpawnNamed(props, "PluginManagerActor")
			panicOnErr(err)
		}
	case *actor.Stopping:
	case *actor.Stopped:
	case *actor.Restarting:
	case string:
		switch msg {
		case "init":
			fmt.Printf("Hello %v\n", msg)
			context.Respond("ok")
		default:
			log.Fatalln("unknown message:", msg)
		}
	}
}
