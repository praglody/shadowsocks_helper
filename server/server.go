package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"shadowsocks_helper/config"
)

func main() {
	// 启动 ss 服务器
	startShadowSocksServer()

	// 启动 web 服务器，供查询配置信息用
	http.HandleFunc("/getssconfig", func(w http.ResponseWriter, req *http.Request) {
		file, _ := os.Open("/data/software/config.json")
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		buffer, _ := ioutil.ReadAll(file)
		fmt.Println(string(buffer))
		if _, err := w.Write(buffer); err != nil {
			fmt.Println(err)
		}
	})

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		panic(err)
	}
}

func startShadowSocksServer() {
	var workDir = "/data/software"
	if _, err := os.Stat(workDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(workDir, os.ModePerm); err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			panic("文件系统错误")
		}
	}

	if _, err := os.Stat(workDir + "/shadowsocks"); err != nil {
		var cmdStr = "cd " + workDir + " && git clone https://github.com/praglody/shadowsocks.git"
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		fmt.Println("代码已下载完毕，项目路径：" + workDir + "/shadowsocks")
	} else {
		fmt.Println("shadowsocks 程序代码存在，项目路径：" + workDir + "/shadowsocks")
	}

	fmt.Println("开始生成配置文件...")

	var configObj = config.GetConfig()
	var listen []*net.Listener
	for i := 0; i < 100; i++ {
		// 获取未被使用的端口号
		l, _ := net.Listen("tcp", "127.0.0.1:0") // listen on localhost
		port := l.Addr().(*net.TCPAddr).Port
		listen = append(listen, &l)

		if port == 8090 {
			i--
			continue
		}
		configObj.PortPassword[port] = config.GetRandomPassword()
	}
	// 关闭端口监听
	for _, l := range listen {
		if err := (*l).Close(); err != nil {
			panic(err)
		}
	}

	// 写入配置文件
	j, _ := json.MarshalIndent(configObj, "", "  ")
	var configFilePath = workDir + "/config.json"
	configFile, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	if _, err := configFile.Write(j); err != nil {
		panic(err)
	}
	if err := configFile.Close(); err != nil {
		panic(err)
	}

	fmt.Println("配置文件创建成功")
	fmt.Println("启动ss服务器")

	killssCmd := "ps -ef|grep 'shadowsocks/server.py -c'|grep -v grep|awk '{print $2}'|xargs kill"
	cmd1 := exec.Command("/bin/sh", "-c", killssCmd)
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		panic(err)
	}

	ssCmd := "nohup python " + workDir + "/shadowsocks/shadowsocks/server.py -c " + workDir + "/config.json >/tmp/ss.log 2>&1 &"
	cmd2 := exec.Command("/bin/sh", "-c", ssCmd)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		panic(err)
	}
}
