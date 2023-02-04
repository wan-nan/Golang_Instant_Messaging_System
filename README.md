# Golang——即时通信系统

幼儿编程项目，也是学Golang做的第一个项目，其实就是个简陋的聊天室 ~~（QQ 2002）~~

做的过程中忘记用GitHub进行版本控制了，下次一定😶

[项目地址](https://github.com/wan-nan/Golang_Instant_Messaging_System)

[参考教程](https://www.yuque.com/aceld/mo95lb/dsk886)

## 功能

- 用户上线及用户下线的广播提醒
- 广播消息
- 私聊消息
- 修改用户名
- 活跃处理（超时强踢）

## 运行

- 客户端运行：在项目目录下打开终端

  `./client.exe -ip [server ip address] -port [server ip port]`

  例如

  `./client.exe -ip 127.0.0.1 -port 8888`（本机回环测试用的127.0.0.1）

  也可以通过``./client.exe -h`查看命令行参数使用帮助

- 服务端运行：在项目目录下打开终端

  `./server.exe -ip [server ip address] -port [server ip port]`

  例如

  `./server.exe -ip 127.0.0.1 -port 8888`（本机回环测试用的127.0.0.1）

  也可以通过``./server.exe -h`查看命令行参数使用帮助

- 服务端和客户端需要在同一个局域网下（例如校园网）

  ~~当然有钱租个公网ip运行服务端也可以~~
  
- 要在Linux下运行需要重新编译

  `go build -o server main.go server.go user.go`

  `go build -o client client.go`

## 个人总结

- 其实还是有很多问题的，例如

  - 客户端可以通过**广播消息**这一功能发送`who`/`rename`/`to|xx|xx`这样的指令直接调用服务端的api，倒是不会引起什么bug，但是这并不合理

  - 客户端中，系统消息与来自其他客户端的消息应该在不同的窗口显示，可惜没做UI

  - ~~超时强踢之后服务器会close掉conn以及channel，但是客户端这边程序不会退出~~

    修改想法历程：（原本从conn TCP中读消息是个单独的协程，即`go client.DealResponse()`，main负责的是`client.Run()`）

    1. **（×）**搞个channel用于协程通信，`go client.DealResponse()`协程中一停main中阻塞的select中的`<-client.stop`就继续执行，然后退出，但是由于`client.Run()`是个循环，很难搞**select**

       ```go
       if msg == "[exit]" {
       			client.stop <- true
       			return
       		}
       ```

    2. **（×）**改变想法，把`client.stop`改成个bool类型，但还是受限于`client.Run()`中的循环嵌套，要加很多对`client.stop`的判断，这与“服务器一旦关闭，客户端立刻退出”的理念不符

    3. **（√）**原本代码结构有问题，`io.Copy(os.Stdout, client.conn)`会导致无法根据服务端反馈的连接关闭的消息来做出反馈，而是直接输出到Stdout（例如用户被踢掉后用户的进程需要根据被踢的消息来自动退出）

       我们首先要建立缓冲区，将conn中的消息手动读到缓冲区中，判断消息内容是否为**服务器关闭客户端的指令**的同时，判断连接conn是否还开着；

       然后用协程执行`client.Run()`，在main中执行`client.DealResponse()`，**保证读消息的进程也可以结束main进程**

  - ~~服务器ip和端口在程序中写死了（已完善）~~

    ```go
    import "flag"
    
    var serverIp string
    var serverPort int
    
    func init() {
    	flag.StringVar(&serverIp, "ip", "127.0.0.1", "要连接的服务器IP地址(默认是127.0.0.1)")
    	flag.IntVar(&serverPort, "port", 8888, "要连接的服务器端口(默认是8888)")
    	//命令行解析
    	flag.Parse()
    }
    ```

- 这个项目一方面熟悉了Golang的基本语法，同时也对C/S模式的应用有了初步的认识

  - [ ] 后续打算接触一下**设计模式**

## Security

[![Security Status](https://www.murphysec.com/platform3/v3/badge/1618015309080866816.svg?t=1)](https://www.murphysec.com/accept?code=a64ac5408ad1622070984da40540f197&type=1&from=2&t=2)