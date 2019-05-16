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
	"time"
)

var connections = make(map[string]*net.Conn)

func main() {
	// 启动 ss 服务器
	startShadowSocksServer()

	// 启动 web 服务器，供查询配置信息用
	go startWebServer()

	// 启动tcp服务器，用于和客户端建立心跳连接
	go func() {
		listen, err := net.Listen("tcp", "0.0.0.0:8091")
		if err != nil {
			fmt.Println(err)
			return
		}
		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go handleTcpConn(conn)
		}
	}()

	for {
		time.Sleep(time.Hour)
	}
}

func handleTcpConn(conn net.Conn) {
	fd_s := conn.RemoteAddr().String()
	if connections[fd_s] != nil {
		if err := (*connections[fd_s]).Close(); err != nil {
			panic(err)
		}
	}
	connections[fd_s] = &conn

	var buf = make([]byte, 12)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			if err := conn.Close(); err != nil {
				fmt.Println(err.Error())
			}
			// 清理连接
			if connections[fd_s] != nil {
				delete(connections, fd_s)
			}
			return
		}

		if string(buf[:4]) == "ping" {
			if _, err := conn.Write([]byte("pong\r\n")); err != nil {
				fmt.Println(err.Error())
			}
		}

		time.Sleep(time.Second * 3)
	}
}

func startWebServer() {
	http.HandleFunc("/getssconfig", func(w http.ResponseWriter, req *http.Request) {
		file, _ := os.Open("/data/software/server_config.json")
		defer func(f *os.File) {
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}(file)

		buffer, _ := ioutil.ReadAll(file)
		if _, err := w.Write(buffer); err != nil {
			fmt.Println(err)
		}
	})

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		fmt.Println(err)
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

		if port == 8090 || port == 8091 {
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
