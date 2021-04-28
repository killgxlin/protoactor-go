package main

type MsgType string

const (
	Actor  MsgType = "actor"
	Err    MsgType = "error"
	Plugin MsgType = "plugin"
	Binary MsgType = "binary"
)

type msgPlugin struct {
	pluginName string
	msgType    MsgType
	msg        string
}
