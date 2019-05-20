package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"shadowsocks_helper/config"
	"shadowsocks_helper/logic"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	// 解析命令行参数
	ip, port := getServerIpAndPort()
	if ip == "" || port == 0 {
		return
	}

	if err := logic.InitWorkDir(); err != nil {
		panic(err)
	}

	if err := logic.CreateCodeFiles(); err != nil {
		panic(err)
	}

	if err := initLocalConfig(ip, port); err != nil {
		panic(err)
	}

	if err := startLocalServer(); err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp4", ip+":8091")
	if err != nil {
		panic(err)
	}

	go func(conn net.Conn) {
		for {
			_, err := io.WriteString(conn, "ping\r\n")
			if err != nil {
				fmt.Println(err)
				break
			}
			time.Sleep(time.Second * 100)
		}
		conn.Close()
	}(conn)

	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			conn.Close()
			break
		}

		line = strings.TrimSpace(line)
		if line == "restart" {
			fmt.Println("收到restart信号，重启本地客户端")

			time.Sleep(time.Second * 3)

			// 重启客户端进程
			if err := initLocalConfig(ip, port); err != nil {
				fmt.Println(err)
			}

			if err := startLocalServer(); err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(time.Second * 3)
	}
}

func startLocalServer() error {
	killSsProcess := "ps -ef|grep 'shadowsocks/local.py -c'|grep -v grep|awk '{print $2}'|xargs kill"
	killSsProcessCmd := exec.Command("/bin/bash", "-c", killSsProcess)
	if err := killSsProcessCmd.Run(); err == nil {
		fmt.Println("关闭已经启动的ss服务器")
	}

	fmt.Println("开始启动ss local")
	defer fmt.Println("启动完毕 ...")

	ssCmd := "nohup python " + config.WorkDir + "/shadowsocks/shadowsocks/local.py -c " +
		config.WorkDir + "/local_config.json >/tmp/ss.log 2>&1 &"
	cmd2 := exec.Command("/bin/bash", "-c", ssCmd)
	cmd2.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} //进程退出后保留子进程
	//cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}	//windows实现
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	return cmd2.Run()
}

func initLocalConfig(ip string, port int) error {
	// 获取服务器配置
	httpClient := &http.Client{}
	response, err := httpClient.Get(fmt.Sprintf("http://%s:%d/getssconfig", ip, port))
	if err != nil {
		return err
	}
	var configStr []byte
	configStr, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// 解析配置文件
	var configObj config.SsConfig
	err = json.Unmarshal(configStr, &configObj)
	if err != nil {
		return err
	}

	var localConfig = config.GetLocalConfig()
	for k, v := range configObj.PortPassword {
		upstream := config.UpstreamServer{
			Weight:     1,
			Server:     ip,
			ServerPort: k,
			Password:   v,
		}

		// 这里做服务器端口可用性检测
		ipaddr := net.JoinHostPort(ip, strconv.Itoa(k))
		if conn, err := net.Dial("tcp", ipaddr); err == nil {
			_ = conn.Close()
		} else {
			fmt.Printf("服务器端口 %s 不可用，已自动摘除\n", ipaddr)
			continue
		}

		localConfig.Upstream = append(localConfig.Upstream, upstream)
	}

	if len(localConfig.Upstream) < 1 {
		panic(errors.New("没有可用的服务器端口"))
	}

	j, _ := json.MarshalIndent(localConfig, "", "  ")

	var configFilePath = config.WorkDir + "/local_config.json"
	configFile, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	if _, err := configFile.Write(j); err != nil {
		return err
	}
	if err := configFile.Close(); err != nil {
		return err
	}

	return nil
}

func getServerIpAndPort() (string, int) {
	var ip string
	var port string
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		if args[i] == "-i" {
			if (i + 1) < len(args) {
				i++
				ip = args[i]
			} else {
				fmt.Fprint(os.Stderr, "请输入正确的服务器IP\n")
				return "", 0
			}
		}
	}

	ip, port, err := net.SplitHostPort(ip)
	if err != nil {
		panic(err)
	}

	if address := net.ParseIP(ip); address == nil {
		fmt.Fprint(os.Stderr, "请输入正确的服务器IP\n")
		return "", 0
	}

	portInt, err := strconv.Atoi(port)
	if err != nil || portInt < 1 || portInt > 65535 {
		fmt.Fprint(os.Stderr, "请输入正确的服务器端口号\n")
		return "", 0
	}

	return ip, portInt
}
