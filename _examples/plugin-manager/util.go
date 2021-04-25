package main

import (
	"fmt"
	"net"
)

func ensureFireWall(name, binPath string) error {
	return nil
}

func findAvailablePort(begin, end int) (int, error) {
	for i := begin; i < end; i++ {
		l, err := net.Listen("tcp4", fmt.Sprintf("localhost:%d", i))
		if err != nil {
			continue
		}
		l.Close()
		return i, nil
	}

	return -1, fmt.Errorf("no available port")
}
