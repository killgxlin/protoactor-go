# 客户端协议
## 协议
### c->S
```
cs b|bind caster
cs u|unbind caster
cs caster CMD
```
### p->c
```
pc state:initing			// 通知cli当前状态
pc state:started
pc date:....
pc log:....
```
### s->p
```
sp cmd						// 转发cli的命令
sp query					// 查询当前所有状态
```
### p->s
```
ps idle						// 空闲状态，可以切换
ps working					// 工作中，不可切换
```
### s->c
```
sc caster actor started		// caster 相关的actor状态
sc caster actor stopping
sc caster actor stopped

sc caster artifact updating		// caster 相关的可执行文件状态
sc caster artifact checking
sc caster artifact downloading
sc caster artifact decompressing
sc caster artifact updated
sc caster artifact uninstalling
sc caster artifact uninstalled

sc caster process launching		// caster 相关的进程状态
sc caster process launched
sc caster process terminating
sc caster process terminated

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
		latest.yml
		alpha.yml
		beta.yml
		?.?.?
			manifest.yml
			plugin.exe
			*
		?.?.?
			manifest.yml
			plugin.exe
			*
```