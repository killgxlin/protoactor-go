# 客户端协议
## 协议
c->s
	- bind caster   // 绑定插件
	- unbind caster // 解绑插件
	- caster CMD    // 发命令给插件
s->c
	- caster bind ok
	- caster notify state:initing
	- caster notify state:started
	- caster notify date:....
	- caster notify log:....

## TODO
服务
	-防火墙检查放行
	-启动停止服务
客户端
	-处理连接
	-websocket
插件
	?客户端命令启动停止插件
	-处理连接
	?安装更新
	?删除
		应用卸载需要删除
	?签名校验
