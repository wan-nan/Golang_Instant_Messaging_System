package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

//服务端构建
type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex //读写锁 sync包中包含go语言的同步机制

	//消息广播的总channel
	Message chan string
}

//创建一个server的API
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,                   //注意这里":"，且最后一个也要带","
		OnlineMap: make(map[string]*User), //全局的map要加锁
		Message:   make(chan string),
	} //新建一个server
	return server
}

//监听总channel进行广播
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock() //————————————加锁————————————//
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + " : " + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//当前链接的业务
	// fmt.Println("链接建立成功")

	// //用户上线，将新上线的用户加入OnlineMap中
	//先创建一个新用户
	user := NewUser(conn, this)
	// this.mapLock.Lock() //————————————加锁————————————//
	// this.OnlineMap[user.Name] = user
	// this.mapLock.Unlock()
	// //广播新用户上限消息（广播用到总的channel）
	// this.BroadCast(user, "Log in!")
	user.Online()

	isLive := make(chan bool) //查看用户是否活跃的channel

	//接收用户传输来的数据（从conn中Read()）
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 { //——————————表示连接关闭——————————//（猜测应该是阻塞式读法）
				// this.BroadCast(user, "Log out!")
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}
			//提取用户的信息(去除最后的'\n')
			msg := string(buf[:n-1])
			//处理用户发来的消息
			user.DoMessage(msg)
			//发送了消息表示当前用户活跃
			isLive <- true
		}
	}()

	//活跃处理
	for {
		select { //————————select活跃处理————————//
		case <-isLive:
			//当前用户活跃，重置定时器
		case <-time.After(time.Minute * 10): //每次进入select，select阻塞时都会启动定时器
			//10s后触发 本质是个channel
			//进入此case说明已经超时
			user.C <- ">>>>您因长时间不活跃即将被踢出!"
			user.C <- "[exit]"
			//销毁资源
			time.Sleep(time.Second * 5) //等一会再关闭channel
			// user.Offline() 这里不需要再Offline 在conn.Read(buf)==0时会自动Offline
			conn.Close()
			close(user.C)
			close(isLive)
			return //runtime.Goexit()
		}
	}
}

//启动服务器的接口——当前类的一个成员方法
func (this *Server) Start() {
	//socket listen

	//少打一个f,Sprint与Sprintf天壤之别
	// s := fmt.Sprint("%s:%d", this.Ip, this.Port)
	// fmt.Println(s)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port)) //net.Listen(network string, address string)
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	//close listen socket 防止遗忘用defer
	defer listener.Close()

	//启动监听广播的Messager
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept() //conn是连接的句柄
		if err != nil {
			fmt.Println("net.Listen err: ", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}

}
