package netx

import (
	"fmt"
	"net"
	"strings"
)

// GetOutBoundIp 获得对外发消息的 IP 地址
// 在注册的时候，有一个难点是获得当前节点的 IP。正常来说，我们都不会从配置文件读。有些公司会在自己的节点的环境变量里面注入 IP，
// 那么可以考虑从环境变量读。
// 这里通过对外面发送 UDP 报文，来确定了自己的 IP。
// PS： 你肯定不能用 127.0.0.1。在注册的时候，基本上都是注册自己和客户端所在的同一局域网的 IP 地址。
func GetOutBoundIp() string {
	// DNS 的地址，国内可以用 114.114.114.114
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// 获取本机ip地址
func getHostIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("get current host ip err: ", err)
		return ""
	}
	addr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(addr.String(), ":")[0]
	return ip
}
