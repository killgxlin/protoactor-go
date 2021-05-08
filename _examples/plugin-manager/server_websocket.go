package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func Listen(netName, addr string) (net.Listener, error) {
	if netName == "tcp4" {
		return net.Listen("tcp4", addr)
	}
	return listen(netName, addr)
}

func listen(netName, addr string) (net.Listener, error) {
	wsl := new(WSListener)

	handler := http.HandlerFunc(wsl.echo)
	http.Handle(fmt.Sprintf("/%s", netName), handler)
	wsl.srv = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	wsl.ch = make(chan *websocket.Conn)

	go func() {
		defer close(wsl.ch)
		if err := wsl.srv.ListenAndServe(); err != http.ErrServerClosed {
			panicOnErr(err)
		}
	}()

	return wsl, nil
}

type WSListener struct {
	srv *http.Server
	ch  chan *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   256,
	WriteBufferSize:  256,
	HandshakeTimeout: 50 * time.Second,
}

func (wsl *WSListener) echo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("websocket upgrader.Upgrade error: %v\n", err)
		return
	}
	wsl.ch <- ws
}

func (wsl *WSListener) Accept() (net.Conn, error) {
	ws, ok := <-wsl.ch
	if !ok {
		return nil, fmt.Errorf("Listenr closed")
	}
	return &WSConn{conn: ws}, nil
}

func (wsl *WSListener) Close() error {
	wsl.srv.Close()
	return nil
}

func (wsl *WSListener) Addr() net.Addr {
	return nil
}

type WSConn struct {
	conn *websocket.Conn
}

func (wsc *WSConn) Read(p []byte) (n int, err error) {
	_, message, err := wsc.conn.ReadMessage()
	if err != nil {
		fmt.Printf("Read err %v\n", err)
		return 0, err
	}

	copy(p, string(message))
	return bytes.Count(message, nil) - 1, nil
}

func (wsc *WSConn) Write(p []byte) (n int, err error) {
	ret := wsc.conn.WriteMessage(websocket.TextMessage, p)
	if ret != nil {
		fmt.Printf("Write err %v\n", err)
		return 0, ret
	}

	return bytes.Count(p, nil) - 1, nil
}
func (wsc *WSConn) Close() error {
	return wsc.conn.Close()
}

func (wsc *WSConn) LocalAddr() net.Addr {
	return wsc.conn.LocalAddr()
}
func (wsc *WSConn) RemoteAddr() net.Addr {
	return wsc.conn.RemoteAddr()
}

func (wsc *WSConn) SetDeadline(t time.Time) error {
	return nil
}
func (wsc *WSConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (wsc *WSConn) SetWriteDeadline(t time.Time) error {
	return nil
}
