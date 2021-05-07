// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"encoding/base64"
	"strings"

	"github.com/gorilla/websocket"
)

func base64Encode(src string) string {
    return base64.StdEncoding.EncodeToString([]byte(src))
}

func base64Decode(src string) string {
	decodeBytes, _ := base64.StdEncoding.DecodeString(src)
    return string(decodeBytes)
}

var addr = flag.String("addr", "localhost:8000", "http service address")

//cmd##start|1080|1|SRFZ
func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/cli"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	writeMsg := func(msg string) {
		c.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
	}

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			log.Printf("recv: %s", string(message))

			if string(message) == "is conneted"{
				continue
			}

			// decode := strings.Split(string(message), "##")
			// decodeMsg := decode[1]
			// if decode[0] == "1" {
			// 	decodeMsg = base64Decode(decode[1])
			// }
			// log.Printf("recv: %s", decodeMsg)
		}
	}()

	for {
		var name1 string
		var name2 string
		var name3 string
		var name string

		fmt.Scanln(&name1, &name2, &name3)
		name = strings.TrimSpace(name1 + " " + name2 + " " + name3)
		fmt.Print(name)	
		writeMsg(name)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}
