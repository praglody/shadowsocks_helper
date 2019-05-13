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
	"strconv"
)

func main() {
	// 解析命令行参数
	ip, port := getServerIpAndPort()
	if ip == "" || port == 0 {
		return
	}

	if err := initLocalConfig(ip, port); err != nil {
		panic(err)
	}

	if err := logic.InitWorkDir(); err != nil {
		panic(err)
	}

	startLocalServer()
}

func startLocalServer() {
	killSsProcess := "ps -ef|grep 'shadowsocks/local.py -c'|grep -v grep|awk '{print $2}'|xargs kill"
	killSsProcessCmd := exec.Command("/bin/sh", "-c", killSsProcess)
	if err := killSsProcessCmd.Run(); err == nil {
		fmt.Println("关闭已经启动的ss服务器")
	}

	fmt.Println("开始启动ss local")

	ssCmd := "nohup python " + config.WorkDir + "/shadowsocks/shadowsocks/local.py -c " + config.WorkDir + "/local_config.json >/tmp/ss.log 2>&1 &"
	cmd2 := exec.Command("/bin/sh", "-c", ssCmd)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		panic(err)
	}
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

		// 这里需要做服务器端口可用性检测

		localConfig.Upstream = append(localConfig.Upstream, upstream)
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

	fmt.Println(string(j))
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
