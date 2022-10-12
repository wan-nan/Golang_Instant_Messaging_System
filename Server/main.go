package main

import "flag"

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "要连接的服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "要连接的服务器端口(默认是8888)")
	//命令行解析
	flag.Parse()
}

//当前进程主入口
func main() {
	server := NewServer(serverIp, serverPort)
	server.Start()
}
