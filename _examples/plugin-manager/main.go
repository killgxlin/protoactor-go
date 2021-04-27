package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
	"github.com/AsynkronIT/protoactor-go/eventstream"
	"github.com/kardianos/service"
)

func panicOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var eventStream = eventstream.NewEventStream()
var actorRegistery sync.Map
var mierSvcPid *actor.PID
var system = actor.NewActorSystem()

type program struct{}

func (p *program) Start(s service.Service) error {
	props := actor.PropsFromProducer(func() actor.Actor { return &MierServiceActor{} }).WithReceiverMiddleware(middleware.Logger)
	pid, err := system.Root.SpawnNamed(props, "MierServiceActor")
	if err != nil {
		return err
	}

	mierSvcPid = pid
	return nil
}
func (p *program) run() {
	// Do work here
}
func (p *program) Stop(s service.Service) error {
	if mierSvcPid == nil {
		return fmt.Errorf("mier service actor is not started")
	}
	return system.Root.StopFuture(mierSvcPid).Wait()
}

func main() {
	ensureFireWall("plugin-manager", os.Args[0])

	svcConfig := &service.Config{
		Name:        "xmpm",
		DisplayName: "xiaomi plugins manage service",
		Description: "This is xiaomi plugins manage service.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 1 {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
		return
	}

	switch os.Args[1] {
	case "i", "init":
		s.Install()
		s.Start()
	case "u", "uninit":
		s.Stop()
		s.Uninstall()
	default:
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	}

}
