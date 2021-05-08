package main

import (
	"bufio"
	gcontext "context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func panicOnErr(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	c, e := net.Dial("tcp", "127.0.0.1:"+os.Args[2])
	panicOnErr(e)
	defer c.Close()

	writeMsg := func(msg string) {
		c.Write([]byte(msg + "\n"))
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

	br := bufio.NewReader(c)
	for line, _, err := br.ReadLine(); err == nil; line, _, err = br.ReadLine() {
		msg := string(line)
		cmd := strings.SplitN(msg, "##", 2)
		switch cmd[0] {
		case "start":
			if cancelFun != nil {
				writeMsg("already started")
				continue
			}
			ctx, cancelFun = gcontext.WithCancel(gcontext.TODO())
			go castRunner(ctx, ctrl)
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
}
