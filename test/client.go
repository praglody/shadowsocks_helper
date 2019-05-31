package main

import (
	"encoding/binary"
	"net"
	"shadowsocks_helper/library/slog"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "192.168.31.177:14213")
	if err != nil {
		slog.Emergency(err)
	}

	// 握手阶段
	n, err := conn.Write([]byte{0x05, 0x01, 0x00})
	if n < 3 || err != nil {
		slog.Emergencyf("Write String err: %v", err)
	}
	var resp = make([]byte, 2)
	n, err = conn.Read(resp)
	if err != nil {
		slog.Info("server resp: %v", resp)
	} else {
		slog.Emergency("Read resp error: %v", err)
	}

	// 建立连接
	n, err = conn.Write([]byte{0x05, 0x01, 0x00, 0x03, 0x0D})
	n, err = conn.Write([]byte("www.baidu.com"))
	var port = make([]byte, 2)
	binary.BigEndian.PutUint16(port, 443)
	n, err = conn.Write(port)

	var resp2 = make([]byte, 10)
	n, err = conn.Read(resp2)
	if err != nil {
		slog.Info("server resp: %v", resp2)
	} else {
		slog.Emergency("Read resp2 error: %v", err)
	}

	time.Sleep(time.Hour)
}
