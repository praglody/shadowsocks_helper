# shadowsocks_helper
自动部署ss客户端和ss服务器的程序

## 编译

在项目根目录运行执行 `make`，会在根目录生成 `ss_server` 和 `ss_local` 两个二进制文件

清理已编译生成的文件： `make clean`

在 mac 平台交叉编译 linux 平台二进制包：`make linux_release`

## 启动

服务端

`nohup ./ss_server >/tmp/ss_server.log 2>&1 &`

客户端

``
`nohup ./ss_local -i [serverip]:8090 >/tmp/ss_local.log 2>&1 &`


## More

- ss_server

ss_server 会随机监听100个端口，配置文件会写在 `/data/software/server_config.json`,同时还会监听 8090 和 8091 端口

8090 端口用来给 ss_local 端提供配置查询的接口

8091 端口是用来维持 local 和 server 之间的心跳的，在未来将会实现通过给 server 端发 kill 信号来重启所有的 local 端程序

- ss_local

ss_local 会先去 server 端拉取配置，然后通过配置启动来做负载均衡

ss_local 还会维持一个到 server 的长连接心跳

## TODO

第一阶段：
- GO 守护进程
- 完善 log 输出

第二阶段：
- 使用 go 重写 python 的 ss，不完全遵守 ss 的协议，使用自己定义的私有协议
- 开发一个代理中间件，让客户端和服务端的流量通过中间件去转发（考虑过nginx4层代理，不过nginx不好做动态的端口监听）

## CONTACT

目前我正在召集一些同学来做点有意思的开源项目，目的是提升自身的编程能力，大家共同进步。

这个项目就是待完善的项目之一，现在做了一个雏形，配合 python 的 ss 已经做到基本可用了。后面针对这个项目，还有很多待完成的想法。

之后，我们会整理出大家都感兴趣的方向，每次选一个议题，并且要把这个议题落实好，做成一个比较完善的开源项目。

技术栈以 GO 为主，php、python、lua、c 等也可以涉及，有兴趣的同学可以加我的微信：`PrageMelody`

![PrageMelody](wx.jpg)
