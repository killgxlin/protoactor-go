package main

import (
	"bufio"
	"net"
	"strings"
)

func panicOnErr(e error) {
	if e != nil {
		panic(error)
	}
}

func main() {
	c, e := net.DialTCP("tcp", "localhost:8888")
	panicOnErr(e)

	rwer := bufio.NewReadWriter(c, c)
	for line, _, err := rwer.ReadLine(); err == nil; line, _, err = rwer.ReadLine() {
		msg := string(line)
		cmd := strings.SplitN(msg, "##", 2)
		switch cmd[0] {
		case "code":
		case "start":
		case "stop":
		case "pause":
		case "resume":
		}

	}
	c.Close()
}
