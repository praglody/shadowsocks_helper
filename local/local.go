package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"shadowsocks_helper/config"
	"strconv"
)

func main() {
	// 解析命令行参数
	ip, port := getServerIpAndPort()
	if ip == "" || port == 0 {
		return
	}
	if err := config.InitWorkDir(); err != nil {
		panic(err)
	}
	initLocalConfig(ip, port)
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
	var port int
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
		} else if args[i] == "-p" {
			if (i + 1) < len(args) {
				i++
				port, _ = strconv.Atoi(args[i])
			} else {
				fmt.Fprint(os.Stderr, "请输入正确的端口号\n")
				return "", 0
			}
		}
	}

	if ip == "" {
		fmt.Fprint(os.Stderr, "请输入正确的服务器IP\n")
		return "", 0
	} else if port < 1 || port > 65535 {
		fmt.Fprint(os.Stderr, "请输入正确的端口号\n")
		return "", 0
	}

	address := net.ParseIP(ip)
	if address == nil {
		fmt.Fprint(os.Stderr, "请输入正确的服务器IP\n")
		return "", 0
	}
	return ip, port
}
