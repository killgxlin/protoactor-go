package main

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
)

// --------------------------------------------------------------------------------
type MierServiceActor struct {
}

func (msa *MierServiceActor) Receive(context actor.Context) {
	switch context.Message().(type) {
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
	}
}
