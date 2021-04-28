# 客户端协议
## 协议
### c->s
```
b|bind caster
u|unbind caster
caster CMD
```
### s->c
```
caster notify state:initing
caster notify state:started
caster notify date:....
caster notify log:....

caster actor started
caster actor stopping
caster actor stopped

caster binary installing
caster binary checking
caster binary downloading
caster binary decompressing
caster binary installed
caster binary launching
caster binary launched
caster binary terminating
caster binary terminated
caster binary uninstalling
caster binary uninstalled
```

## TODO
服务
- 防火墙检查放行
- 启动停止服务
客户端
- 处理连接
- websocket
插件
- ?客户端命令启动停止插件
- 处理连接
- ?安装更新
- ?删除
- ?签名校验

## 架构
cli和plugin通信方式

## plugin
### fds目录结构
```
https://url.path.to/plugins
	caster
		latest.yml
		alpha.yml
		beta.yml
		plugin-?.?.?.zip
```
### 解压后目录结构
```
plugin-root
	caster
		?.?.?
			manifest.yml
			plugin.exe
			*
		?.?.?
			plugin.exe
			*
```