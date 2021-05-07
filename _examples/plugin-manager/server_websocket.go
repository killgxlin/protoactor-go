package main

import (
	"io"
	"net"
	"net/http"
    "context"
	"fmt"
	"bytes"
	// "flag"
	"time"

	"github.com/gorilla/websocket"
)

// websocket的Listener和Conn需要分别满足以下两个接口
var l net.Listener
var c io.ReadWriteCloser

// func Listen(netName, addr string) (net.Listener, error) {
// 	return net.Listen("tcp4", addr)
// }


func Listen(netName, addr string) (*WSListener, error) {
	// wsl := &WSListener{}
	wsl := new(WSListener)
	wsl.init(netName, addr)
	return wsl, nil
}

type WSListener struct {
	srv *http.Server
	ctx context.Context
	cancel context.CancelFunc
	c *WSConn
}

var upgrader = websocket.Upgrader{  
	ReadBufferSize: 64,
	WriteBufferSize: 128,
	HandshakeTimeout: 50 * time.Second,
}

func (wsl *WSListener) echo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("websocket upgrader.Upgrade error: %v\n", err)
		return
	}

	wsl.c = new(WSConn)
	wsl.c.conn = ws
	wsl.cancel()
}

func (wsl *WSListener) init (cmd, addr string) error {
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
func (wsl *WSListener) Accept() (*WSConn, error) {
	wsl.ctx, wsl.cancel = context.WithCancel(context.Background())
	for{
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
	return bytes.Count(message, nil)-1, nil
}

func (wsc *WSConn) Write(p []byte) (n int, err error) {
	ret := wsc.conn.WriteMessage(websocket.TextMessage, p)
	if ret!= nil {
		fmt.Printf("Write err %v\n", err)
		return 0, ret
	}

	return bytes.Count(p, nil)-1, nil
}
func (wsc *WSConn) Close() error {
	return wsc.conn.Close()
}
