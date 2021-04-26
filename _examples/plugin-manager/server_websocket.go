package main

import (
	"io"
	"net"
)

/*
type PluginInfo struct {
	name string
	proc string
	Auth string
	conn *websocket.Conn
}

var srv *http.Server
var exit = false

var OsName string

var TextMessage = 1

var encodeType = "0"

var upgrader = websocket.Upgrader{} // use default options

func base64Encode(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}

func base64Decode(src string) string {
	decodeBytes, _ := base64.StdEncoding.DecodeString(src)
	return string(decodeBytes)
}

var listClient = make(map[string]PluginInfo)

func echo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	ws, err := upgrader.Upgrade(w, r, nil)
	var plugin_name string = ""

	empty := func() {
		delete(listClient, plugin_name)
		plugin_name = ""
	}

	quit := func() {
		Logger.Info("echo websocket客户端连接断开...")
		ws.Close()
		PluginExit(plugin_name)
		empty()
		cancel()
	}

	if err != nil {
		Logger.Errorf("websocket upgrader.Upgrade error: %s", err)
		return
	}
	defer quit()
	Logger.Info("echo websocket客户端连接...")
	// inited := false
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			Logger.Errorf("websocket ReadMessage error: %s", err)
			return
		}

		Logger.Infof("read from client, %s", string(message))
		// if !inited && strings.HasPrefix(string(message), "Hello!") {
		//     inited = true
		//     ws.WriteMessage(TextMessage, []byte("is conneted"))

		//     encodeType = string(message)[6:]
		//     Logger.Infof("encodeType %s", encodeType)
		//     continue
		// }

		// decode := strings.Split(string(message), "##")
		// if decode[0] == "1" {
		//     decodeMsg = base64Decode(decode[1])
		// }
		// Logger.Infof("read from client, %s", decodeMsg)

		decodeMsg := string(message)
		if decodeMsg == "Hello!" {
			ws.WriteMessage(TextMessage, []byte("is conneted"))
			continue
		} else if decodeMsg == "quit" {
			return
		}

		s := strings.Split(decodeMsg, "##")
		if len(s) <= 1 {
			sendMsgToClient(ws, "write cmd error")

			continue
		}

		switch {
		case s[0] == "init":
			if plugin_name != "" {
				sendMsgToClient(ws, plugin_name+":is inited")
				continue
			}
			if _, ok := listClient[s[1]]; ok {
				sendMsgToClient(ws, s[1]+" is exist")
				continue
			}

			plugin_name = s[1]
			go PluginsManager(ctx, plugin_name, ws)
		case s[0] == "close":
			Logger.Infof("close %s\n", plugin_name)
			PluginExit(plugin_name)
			empty()
		// case s[0] == "status":
		//     Logger.Infof("status %s\n", s[1])
		//     if _,ok :=listClient[s[1]]; ok!=false {
		//         if listClient[s[1]].conn != nil {
		//             listClient[plugin_name].conn.WriteMessage(TextMessage, []byte("plugins is inited"))
		//             return
		//         }
		//     }
		//     listClient[plugin_name].conn.WriteMessage(TextMessage, []byte("plugins is not inited"))
		//     return
		case s[0] == "cmd":
			ret := SendCommandToPlugin(plugin_name, strings.Replace(s[1], "|", " ", -1))
			if ret == false {
				sendMsgToClient(ws, "not init")
			}
		default:
			sendMsgToClient(ws, "cmd error")
		}
	}
}

func sendMsgToClient(ws *websocket.Conn, info string) {
	if ws != nil {
		ws.WriteMessage(TextMessage, []byte(info))
		// if encodeType == "0" {
		//     msg := "0##" + info
		//     ws.WriteMessage(TextMessage, []byte(msg))
		// } else if encodeType == "1"{
		//     msg := "1##" + base64Encode(info)
		//     ws.WriteMessage(TextMessage, []byte(msg))
		// }
	}
}

func SendMsgToClient(plugin_name string, msg string) {
	if _, ok := listClient[plugin_name]; ok != false {
		if listClient[plugin_name].conn != nil {
			sendMsg := plugin_name + "::" + msg
			sendMsgToClient(listClient[plugin_name].conn, sendMsg)
		} else {
			Logger.Warning("conn destroyed")
		}
	}
}

func websocketServer() {
	Logger.Info("websocketServer running")
	flag.Parse()
	handler := http.HandlerFunc(echo)
	http.Handle("/", handler)
	srv = &http.Server{
		Addr:    ":59421",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		}

	}()
}

func InitWebsocket() {
	OsName = runtime.GOOS
	Logger.Infof("OsName:%s", OsName)

	EnsureFirewall("mierpluginsmanager", os.Args[0])
	websocketServer()
}

func ExitWebsocket() {
	for pluginName := range listClient {
		PluginExit(pluginName)
	}

	if err := srv.Shutdown(nil); err != nil {
		panic(err)
	}

	Logger.Info("srv.Shutdown")
	exit = true
}
*/

// type WSListener struct {
// }

// func (wsl *WSListener) Accept() (net.Conn, error) {
// 	return nil, nil
// }

// func (wsl *WSListener) Close() error {
// 	return nil
// }

// func (wsl *WSListener) Addr() net.Addr {
// 	return nil
// }

// func ListenWS() *WSListener {
// 	wsl := &WSListener{}
// 	return wsl
// }

// websocket的Listener和Conn需要分别满足以下两个接口
var l net.Listener
var c io.ReadWriteCloser

func Listen(netName, addr string) (net.Listener, error) {
	return net.Listen("tcp4", "127.0.0.1:9999")
}

type WSListener struct {
}

// Accept waits for and returns the next connection to the listener.
func (wsl *WSListener) Accept() (net.Conn, error) {
	return nil, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (wsl *WSListener) Close() error {
	return nil
}

// Addr returns the listener's network address.
func (wsl *WSListener) Addr() net.Addr {
	return nil
}

type WSConn struct {
}

func (wsc *WSConn) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (wsc *WSConn) Write(p []byte) (n int, err error) {
	return 0, nil
}
func (wsc *WSConn) Close() error {
	return nil
}
