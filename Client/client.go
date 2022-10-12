/*
	客户端这边的功能就像是
	【把指令发送给tcp套接字，通过tcp套接字传输到服务器，来调用服务器端的api】
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       -1,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn
	return client
}

//这个要并行执行，所以用单独的goroutine执行
//处理server回应的消息， 直接显示到标准输出即可
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上, 永久阻塞监听

	// io.Copy(os.Stdout, client.conn)
	//io.Copy()是将源复制到目标，并且是按默认的缓冲区32k循环操作的，不会将内容一次性全写入内存中

	buf := make([]byte, 4096) //——————————要读进去还是要make声明的——————————//
	for {
		n, err := client.conn.Read(buf)
		if n == 0 { //——————————表示连接关闭——————————//（猜测应该是阻塞式读法）
			fmt.Println(">>>>与服务端的网络连接关闭，程序即将退出")
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err: ", err)
			return
		}
		//提取用户的信息(去除最后的'\n')
		msg := string(buf[:n-1])
		if msg == "[exit]" {
			fmt.Println(">>>>被服务端踢出，程序即将退出")
			return
		}
		fmt.Println(msg)
		// fmt.Println(buf)
	}
}

func (client *Client) msgSender(sendMsg string) {
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

//查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	client.msgSender(sendMsg)
}

//私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容, exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				client.msgSender(sendMsg)
			}

			chatMsg = ""
			fmt.Println(">>>>请输入消息内容, exit退出:")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发给服务器

		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			client.msgSender(sendMsg)
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}

}
func (client *Client) UpdateName() bool {

	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	client.msgSender(sendMsg)

	return true
}

//菜单menu方法：打印菜单；确保用户输入正确数字
func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.退出")

	fmt.Scanln(&flag) //输入不是数字时会赋给flag 0！！！
	// fmt.Println(flag)
	if flag >= 1 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}
func (client *Client) Run() {
	//这种两个for嵌套比较巧妙
	for client.flag != 4 {
		for client.menu() != true {
		}

		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
			// break
		case 2:
			//私聊模式
			client.PrivateChat()
			// break
		case 3:
			//更新用户名
			client.UpdateName()
			// break
		}
	}
}

var serverIp string
var serverPort int

//./client -ip 127.0.0.1 -port 8888
//绑定命令行参数
func init() { //————————命令行解析————————//
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
	//命令行解析
	flag.Parse()
}

func main() {
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>服务器连接失败")
		return
	}
	fmt.Println(">>>>>服务器连接成功")

	/*
		//单独开启一个goroutine去处理server的回执消息【用io库的话根本不需要！！！】
		go client.DealResponse()
		// io.Copy(os.Stdout, client.conn)

		client.Run()
	*/
	go client.Run()
	client.DealResponse()
}
