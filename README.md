# shadowsocks_helper
自动部署ss客户端和ss服务器的程序

## 编译

在项目根目录运行执行 `make`

清理 `make clean`

交叉编译 linux 平台二进制包 `make linux_release`

## 启动

服务端

`./ss_server`

客户端

`./ss_local -i [serverip]:8090`


## More

- ss_server

ss_server 会随机监听100个端口，配置文件会写在 `/data/software/server_config.json`,同时还会监听 8090 和 8091 端口

8090 端口用来给 ss_local 端提供配置查询的接口

8091 端口是用来维持 local 和 server 之间的心跳的，在未来将会实现通过给 server 端发 kill 信号来重启所有的 local 端程序

- ss_local

ss_local 会先去 server 端拉取配置，然后通过配置启动来做负载均衡

ss_local 还会维持一个到 server 的长连接心跳

## TODO

- GO 守护进程
- ~~信号控制~~
- 完善心跳长连接
- 使用 log 包管理日志

## CONTACT

目前我正在召集一些同学来做点有意思的开源项目，目的是提升自身的编程能力，大家共同进步。

这个项目就是待完善的项目之一，现在做了一个雏形，配合 python 的 ss 已经做到基本可用了。后面针对这个项目，还有很多待完成的想法。

之后，我们会整理出大家都感兴趣的方向，每次选一个议题，并且要把这个议题落实好，做成一个比较完善的开源项目。

有兴趣的同学可以加我的微信：

![PrageMelody](wx.jpg)