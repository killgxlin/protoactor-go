package main

import (
	gcontext "context"
	"fmt"
	"os"
	"strings"
	"time"
	"net/url"
	"flag"

	"github.com/gorilla/websocket"
)

func panicOnErr(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	InitLogger("c:\\Myself\\code\\git\\protoactor-go\\_examples\\plugin-manager\\plugins\\")

	var addr = flag.String("addr", "localhost:"+os.Args[2], "http service address")

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/pm"}
	c, _, e := websocket.DefaultDialer.Dial(u.String(), nil)
	if e != nil {
		Logger.Errorf("dial:%v", e)
	}
	defer c.Close()

	writeMsg := func(msg string) {
		c.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
	}

	writeMsg(os.Args[1])

	castRunner := func(ctx gcontext.Context, ctrl chan string) {
		paused := false
		ticker := time.NewTicker(time.Second)
		cnt := 0
		for {
			select {
			case msg := <-ctrl:
				switch msg {
				case "pause":
					paused = true
				case "resume":
					paused = false
				}
			case <-ctx.Done():
				return
			case <-ticker.C:
				if paused {
					continue
				}
				cnt++
				writeMsg(fmt.Sprintf("%s heartbeat count:%d", os.Args[1], cnt))
			}
		}
	}

	// -----------------------------------------------

	var (
		ctrl                          = make(chan string)
		ctx       gcontext.Context    = nil
		cancelFun gcontext.CancelFunc = nil
	)

	cancelCast := func() {
		if cancelFun == nil {
			return
		}

		cancelFun()
		cancelFun = nil
	}
	defer cancelCast()

	ch := make(chan int, 1)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				Logger.Errorf("read err:%v", err)
				fmt.Println("read:", err)
				ch<-1
				return
			}

			msg := string(message)

			cmd := strings.SplitN(msg, "##", 2)
			Logger.Infof("msg:%v", cmd[0])

			switch cmd[0] {
			case "start":
				if cancelFun != nil {
					Logger.Info("already started")
					writeMsg("already started")
					continue
				}
				ctx, cancelFun = gcontext.WithCancel(gcontext.TODO())
				go castRunner(ctx, ctrl)
				Logger.Info("started")
				writeMsg("started")
			case "stop":
				if cancelFun == nil {
					writeMsg("not started")
					continue
				}
				cancelCast()
				writeMsg("stopped")
			case "pause":
				ctrl <- "pause"
				writeMsg("paused")
			case "resume":
				ctrl <- "resume"
				writeMsg("resumed")
			}
		}
	}()


	<-ch
}
