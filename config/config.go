package config

import (
	"math/rand"
	"time"
)

const WorkDir = "/data/software"

type SsConfig struct {
	Server           string         `json:"server"`
	PortPassword     map[int]string `json:"port_password"`
	Method           string         `json:"method"`
	Timeout          int            `json:"timeout"`
	FastOpen         bool           `json:"fast_open"`
	DnsServer        [2]string      `json:"dns_server"`
	TunnelRemote     string         `json:"tunnel_remote"`
	TunnelRemotePort int            `json:"tunnel_remote_port"`
	TunnelPort       int            `json:"tunnel_port"`
}

type UpstreamServer struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Weight     int    `json:"weight"`
}

type LocalConfig struct {
	Upstream         []UpstreamServer `json:"upstream"`
	LocalAddress     string           `json:"local_address"`
	LocalPort        int              `json:"local_port"`
	Method           string           `json:"method"`
	Timeout          int              `json:"timeout"`
	FastOpen         bool             `json:"fast_open"`
	DnsServer        [2]string        `json:"dns_server"`
	TunnelRemote     string           `json:"tunnel_remote"`
	TunnelRemotePort int              `json:"tunnel_remote_port"`
	TunnelPort       int              `json:"tunnel_port"`
}

func GetConfig() *SsConfig {
	return &SsConfig{
		"0.0.0.0",
		map[int]string{},
		"aes-256-cfb",
		600,
		true,
		[2]string{"8.8.8.8", "223.5.5.5"},
		"8.8.8.8",
		53,
		53,
	}
}

func GetLocalConfig() *LocalConfig {
	return &LocalConfig{
		[]UpstreamServer{},
		"0.0.0.0",
		14213,
		"aes-256-cfb",
		600,
		true,
		[2]string{"8.8.8.8", "223.5.5.5"},
		"8.8.8.8",
		53,
		53,
	}
}

func GetRandomPassword() string {
	var result []byte
	var str = "0123456789abcdefghijklmnopqrstuvwxyz;,.:|/-+=_#@!~"
	bytes := []byte(str)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 12; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
