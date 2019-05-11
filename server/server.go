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
	"shadowsocks_helper/logic"
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
	if err := logic.InitWorkDir(); err != nil {
		panic(err)
	}

	if err := logic.CreateCodeFiles(); err != nil {
		panic(err)
	}

	workDir := config.WorkDir

	killSsProcess := "ps -ef|grep 'shadowsocks/server.py -c'|grep -v grep|awk '{print $2}'|xargs kill"
	killSsProcessCmd := exec.Command("/bin/sh", "-c", killSsProcess)
	if err := killSsProcessCmd.Run(); err == nil {
		fmt.Println("关闭已经启动的ss服务器")
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
	var configFilePath = workDir + "/server_config.json"
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
	fmt.Println("开始启动ss服务器")

	ssCmd := "nohup python " + workDir + "/shadowsocks/shadowsocks/server.py -c " + workDir + "/server_config.json >/tmp/ss.log 2>&1 &"
	cmd2 := exec.Command("/bin/sh", "-c", ssCmd)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		panic(err)
	}
}
