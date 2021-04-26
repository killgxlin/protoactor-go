package main

import (
	"fmt"
	"net"
)

func ensureFireWall(name, binPath string) error {
	return nil
}

func findAvailableAddr(begin, end int) (string, error) {
	for i := begin; i < end; i++ {
		addr := fmt.Sprintf("127.0.0.1:%d", i)
		l, err := net.Listen("tcp4", addr)
		if err != nil {
			continue
		}
		l.Close()
		return addr, nil
	}

	return "", fmt.Errorf("no available port")
}
