# 客户端协议
## 协议

```
// cli->svc
cs b|bind caster
cs u|unbind caster

// cli->plug
cp caster CMD

// plug->cli
pc state:initing			// 通知cli当前状态
pc state:started
pc date:....
pc log:....

// svc->plug
sp cmd						// 转发cli的命令
sp query					// 查询当前所有状态

// plug->svc
ps idle						// 空闲状态，可以切换
ps working					// 工作中，不可切换

// svc->ccli
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

打包发布
	数据
		https://url.path.to/plugins
			(alpha|beta|latest)[-(linux|mac)]?.yml
			caster-?.?.?[-(linux|mac)].zip
			wireguard-?.?.?[-(linux|mac)].zip
		pluginSrcDir
			caster
				version
			wireguard
				version
		pluginAssetsDir
			caster
			wireguard
		buildCacheDir
			caster-?.?.?[-(linux|mac)].zip
			wireguard-?.?.?[-(linux|mac)].zip
			publish plugins ?.?.? linux mac alpha
				upload packages
				fetch yaml
				modify yaml
				upload yaml

	过程
		打包
			获得
			编译二进制
			签名
			压缩
			
		发布
			准备manifest
			上传
				先上传版本
				再上传manifest
			回滚
				更新到对应版本的manifest

plugin管理
	数据
		installDir
			caster-?.?.?.zip
		pluginCacheDir
			caster
				alpha.yml
					version
				beta.yml
					version
				latest.yml
					version
				?.?.?
					manifest.yml
					plugin.exe
					*
				?.?.?
					manifest.yml
					plugin.exe
					*
	过程
		给cli的接口
			bind(pluginName, minVersion)
			cmd(pluginName, command)
		给svc的接口
			init(installDir, pluginCacheDir, channel)
			restart(pluginName)
			uninstall(pluginName)
			checkUpdate(pluginName)
				
		