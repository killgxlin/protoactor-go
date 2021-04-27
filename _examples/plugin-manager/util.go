package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

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

func ensureFireWall(name, binPath string) error {

	return _ensureFireWall(name, binPath)
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
