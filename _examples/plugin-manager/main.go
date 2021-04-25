package main

import (
	"log"
	"sync"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
	"github.com/AsynkronIT/protoactor-go/eventstream"
)

/*
调研方向
	启动销毁
	调用方式
		消息
			ctx.Send
			ctx.Request
		rpc
			同步
				future.wait
			异步
				ctx.AwaitFuture
		eventstream
			subscribe
			publish
	日志


确定对象
确定对象通信-输入输出
用例
	启动
		初始化插件管理器
			初始化插件
		初始化CliManager
	检查插件更新
	客户端上线
		客户端绑定插件
		获取插件状态
	客户端绑定插件
	客户端发命令
	插件状态变更同步客户端
	关闭
		停止CliManager
			杀掉所有client
		停止插件管理器
			停止插件
				停止进程


mier-service
	event-stream
	plugin-manager
		plugin
			plugin-session
			process
	ws-server
		ws-session

服务
	?防火墙检查放行
	?启动停止服务
客户端
	-处理连接
插件
	?客户端命令启动停止插件
	-处理连接
	?安装更新
	?删除
		应用卸载需要删除
	?签名校验

*/

func panicOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var eventStream = eventstream.NewEventStream()
var actorRegistery sync.Map

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor { return &MierServiceActor{} }).WithReceiverMiddleware(middleware.Logger)
	pid, err := system.Root.SpawnNamed(props, "MierServiceActor")
	panicOnErr(err)
	_, _ = console.ReadLine()
	system.Root.Stop(pid)
	_, _ = console.ReadLine()
}
