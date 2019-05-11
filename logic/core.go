package logic

import (
	"fmt"
	"os"
	"os/exec"
	"shadowsocks_helper/config"
)

func CreateCodeFiles() error {
	workDir := config.WorkDir
	if _, err := os.Stat(workDir + "/shadowsocks"); err != nil {
		var cmdStr = "cd " + workDir + " && git clone https://github.com/praglody/shadowsocks.git"
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Println("代码已下载完毕，项目路径：" + workDir + "/shadowsocks")
	} else {
		fmt.Println("shadowsocks 程序代码存在，项目路径：" + workDir + "/shadowsocks")
	}
	return nil
}

func InitWorkDir() error {
	if _, err := os.Stat(config.WorkDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(config.WorkDir, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
