package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//创建一个用户的API（server.go中也有类似的方法）
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string), //make channel的时候不加容量，表示无缓冲的channel，保证【同步读写】
		conn:   conn,
		server: server,
	}

	//服务器监听每个客户的channel，如果有消息就把channel中的消息取出，送入conn套接字中通过网络发送
	go user.ListenMessage()
	return user
}

func (this *User) Online() {
	//用户上线，将新上线的用户加入OnlineMap中

	this.server.mapLock.Lock() //————————————加锁————————————//
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播新用户上限消息（广播用到总的channel）
	this.server.BroadCast(this, "用户已上线!")
}

func (this *User) Offline() {
	//用户下线，将新下线的用户从OnlineMap中删除

	this.server.mapLock.Lock() //————————————加锁————————————//
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//广播用户下线消息（广播用到总的channel）
	this.server.BroadCast(this, "用户已下线!")
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//"who"指令查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": 在线..."
			this.C <- onlineMsg //把这个消息发送到管道里
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		/*
			这里要【一锁二查三更新】
		*/
		this.server.mapLock.Lock()
		//消息格式:"rename|南海鑫"
		newName := strings.Split(msg, "|")[1] //————————————strings.Split()————————————————//
		// fmt.Println(newName)
		//判断name是否已经被占用
		_, ok := this.server.OnlineMap[newName] //在map中查询是否已经存在
		if ok {
			this.C <- ">>>>用户名 [" + newName + "] 已被占用"
		} else {
			// this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.Name = newName
			this.C <- ">>>>您的用户名已被修改为: " + newName
			// this.server.mapLock.Unlock()
		}
		this.server.mapLock.Unlock()

	} else if len(msg) >= 3 && msg[:3] == "to|" {
		//消息格式：to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.C <- ">>>>指令格式错误，正确格式:\"to|name|content\""
			return
		}
		this.server.mapLock.Lock()
		remoteUser, ok := this.server.OnlineMap[remoteName]
		this.server.mapLock.Unlock()
		if !ok {
			this.C <- ">>>>用户 " + remoteName + " 不存在!"
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.C <- ">>>>消息内容无效!"
			return
		}
		remoteUser.C <- ">>>>" + this.Name + " 对您说: \"" + content + "\""
	} else {
		this.server.BroadCast(this, msg)
	}

}

//监听当前user channel的方法，一旦有消息就发送给对端客户端
//服务器监听每个客户的channel，如果有消息就把channel中的消息取出，送入conn套接字中通过网络发送
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
