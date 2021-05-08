package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func Listen(netName, addr string) (net.Listener, error) {
	if (netName == "tcp4") {
		return net.Listen("tcp4", addr)
	}

	wsl := new(WSListener)
	wsl.init(netName, addr)
	return wsl, nil
}

type WSListener struct {
	srv    *http.Server
	ctx    context.Context
	cancel context.CancelFunc
	c      net.Conn
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

	c := new(WSConn)
	c.conn = ws
	wsl.c = c 
	wsl.cancel()
}

func (wsl *WSListener) init(cmd, addr string) error {
	handler := http.HandlerFunc(wsl.echo)
	http.Handle(fmt.Sprintf("/%s", cmd), handler)
	wsl.srv = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var ret error = nil
	go func() {
		if err := wsl.srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe: err%v\n", err)
			ret = err
			return
		}

	}()

	return ret
}

// Accept waits for and returns the next connection to the listener.
func (wsl *WSListener) Accept() (net.Conn, error) {
	wsl.ctx, wsl.cancel = context.WithCancel(context.Background())
	for {
		select {
		case <-wsl.ctx.Done():
			return wsl.c, nil
		default:
			continue
		}
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (wsl *WSListener) Close() error {
	wsl.c.Close()
	return nil
}

// Addr returns the listener's network address.
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
