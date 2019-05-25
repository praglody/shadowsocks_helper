package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"shadowsocks_helper/config"
	"shadowsocks_helper/library/slog"
	"shadowsocks_helper/logic"
	"syscall"
	"time"
)

var connections = make(map[string]*net.Conn)
var configString []byte

func main() {

	if err := logic.InitWorkDir(); err != nil {
		slog.Emergency(err)
	}

	if err := logic.CreateCodeFiles(); err != nil {
		slog.Emergency(err)
	}

	// 启动 ss 服务器
	startShadowSocksServer()

	// 启动 web 服务器，供查询配置信息用
	go startWebServer()

	// 启动tcp服务器，用于和客户端建立心跳连接
	go func() {
		listen, err := net.Listen("tcp4", "0.0.0.0:8091")
		if err != nil {
			slog.Info(err)
			return
		}
		for {
			conn, err := listen.Accept()
			if err != nil {
				slog.Info(err)
				continue
			}
			go handleTcpConn(conn)
		}
	}()

	var c = make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGUSR1, syscall.SIGUSR2:
			slog.Info("shadowsocks_helper signal", s)
			startShadowSocksServer()
			time.Sleep(time.Second * 3)
			for _, conn := range connections {
				if _, err := io.WriteString(*conn, "restart\r\n"); err != nil {
					slog.Error(err)
				}
			}
			break
		case syscall.SIGINT, syscall.SIGTERM:
			slog.Info("shadowsocks_helper server exit")
			return
		}
	}
}

func handleTcpConn(conn net.Conn) {
	fd_s := conn.RemoteAddr().String()
	if connections[fd_s] != nil {
		if err := (*connections[fd_s]).Close(); err != nil {
			slog.Error(err)
		}
	}
	connections[fd_s] = &conn

	var buf = make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			slog.Info(err.Error())
			if err := conn.Close(); err != nil {
				slog.Error(err.Error())
			}
			// 清理连接
			if connections[fd_s] != nil {
				delete(connections, fd_s)
			}
			return
		}

		if string(buf[:4]) == "ping" {
			if _, err := conn.Write([]byte("pong\r\n")); err != nil {
				slog.Error(err.Error())
			}
		}

		if n == 512 {
			continue
		}

		time.Sleep(time.Second * 3)
	}
}

func startWebServer() {
	http.HandleFunc("/getssconfig", func(w http.ResponseWriter, req *http.Request) {
		//file, _ := os.Open("/data/software/server_config.json")
		//defer func(f *os.File) {
		//	if err := f.Close(); err != nil {
		//		fmt.Println(err)
		//	}
		//}(file)
		//
		//buffer, _ := ioutil.ReadAll(file)
		if _, err := w.Write(configString); err != nil {
			slog.Error(err)
		}
	})

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		slog.Error(err)
	}
}

func startShadowSocksServer() {
	workDir := config.WorkDir

	killSsProcess := "ps -ef|grep 'shadowsocks/server.py -c'|grep -v grep|awk '{print $2}'|xargs kill"
	killSsProcessCmd := exec.Command("/bin/bash", "-c", killSsProcess)
	if err := killSsProcessCmd.Run(); err == nil {
		slog.Info("关闭已经启动的ss服务器")
	}

	slog.Info("开始生成配置文件...")

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
			slog.Error(err)
		}
	}

	// 写入配置文件
	configString, _ = json.MarshalIndent(configObj, "", "  ")
	var configFilePath = workDir + "/server_config.json"
	configFile, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		slog.Emergency(err)
	}
	if _, err := configFile.Write(configString); err != nil {
		slog.Emergency(err)
	}
	if err := configFile.Close(); err != nil {
		slog.Emergency(err)
	}

	slog.Info("配置文件创建成功")
	slog.Info("开始启动ss服务器")

	ssCmd := "nohup python " +
		workDir + "/shadowsocks/shadowsocks/server.py -c " +
		workDir + "/server_config.json >/tmp/ss.log 2>&1 &"
	cmd2 := exec.Command("/bin/bash", "-c", ssCmd)
	cmd2.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} //进程退出后保留子进程
	//cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}	//windows实现
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		slog.Emergency(err)
	}
}
