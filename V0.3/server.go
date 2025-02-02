package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

//server.go主要用于服务端的基本构建

// * 一个server服务器主要包括两个属性：ip与port（端口）
type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User // 保存在线用户，key是用户名，value是用户对象
	mapLock   sync.RWMutex     //? 读写锁。 sync.RWMutex 用于在多线程（Goroutine）环境下保护对 OnlineMap 的读写操作，避免并发访问时的数据竞争。锁

	// 消息广播的channel
	Message chan string //用于传递消息的channel

}

//? 在go中，基本封装通常通过定义结构体和方法来实现
//? 首字母大小写控制访问权限
//? 通过方法访问私有字段
//!  基本封装：将数据（属性）和操作数据的方法（函数）组合到一个类中，同时通过访问控制来保护数据不被外界直接修改，从而实现数据的隐藏和模块化管理。
//!  1.隐藏实现细节：将数据和方法封装在一起，不让外界直接访问内部的实现细节。
//!  2.提供访问接口：通过公开的方法（Getter 和 Setter），让外界访问或修改数据时保持受控。
//!  3.提高代码可维护性：修改实现时，接口不变，减少对外部代码的影响。

// * 创建一个工厂函数,用于创建Server实例
// ?为了更加安全、灵活和统一的方式来初始化对象或结构体实例
func NewServer(ip string, port int) *Server {
	server := &Server{ //初始化Server，创建了一个指向 Server 结构体的指针，并将其赋值给变量 server。 “&” 是取地址符号，它返回一个指向新创建的 Server 实例的指针
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server //返回实例指针（server）
}

// *持续监听message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message //从message channel中取出消息
		this.mapLock.Lock()

		//? 将msg发送给全部在线的User
		for _, cli := range this.OnlineMap { //遍历所有在线用户

			cli.C <- msg //将消息发送到每个用户的消息channel

		}

		this.mapLock.Unlock()
	}
}

// * 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) { //? 两个参数分别代表用户信息和要广播的消息内容。

	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg //? 将格式化的信息发送给Message通道

}

// * 用于处理链接需要做的业务。定义了 Server 类型的方法 Handler，该方法接收一个 net.Conn 类型的参数 conn，它表示与客户端的连接。
func (this *Server) Handler(conn net.Conn) {
	//当前链接的业务
	////fmt.Println("链接建立成功！")

	//创建用户实例,获取信息
	user := NewUser(conn)

	//用户上线，将用户加入到onlineMap中
	this.mapLock.Lock()              //? 先上锁OnlineMap，防止数据竞争
	this.OnlineMap[user.Name] = user // 将用户对象user添加到OnlineMap中，user.Name作为key，user作为值
	this.mapLock.Unlock()            //? 解锁，释放

	//广播当前用户上线消息
	this.BroadCast(user, "已上线")

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
				return
			}
			//提取用户消息（去掉换行符“\n”）
			msg := string(buf[:n-1])

			//将得到的消息进行广播（及显示给其他用户）
			this.BroadCast(user, msg)
		}
	}()

	//当前handler阻塞
	select {} //? 阻塞当前连接，使 Goroutine 不会退出。这样可以保持与用户的连接，等待进一步的消息传递。

}

// * 给server提供一个启动服务器的方法(API)
func (this *Server) Start() {

	// socket listen
	//! 开启监听，连接网络通信的端点，监听表示服务器等待并接收客户端的连接请求。
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port)) //? %s表示格式化为字符串;%d表示格式化为整数
	if err != nil {
		fmt.Println("net.Listen err:", err) //打印错误信息
		return
	}

	//close listen socket
	//! 关闭连接，当服务器不再需要接受连接时，关闭监听的 socket，释放资源，避免内存泄漏或其他潜在问题。
	defer listener.Close() //? defer表示在函数执行结束时再执行 listener.Close()，即使出现错误或提前返回也能确保资源被释放。

	//启动监听Message的goroutine
	go this.ListenMessager()

	// accept
	//! 接收连接，处理客户端请求的连接。
	for { //? for来用于服务器不断接受连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue //如果接受连接时发生错误（例如连接被中断或客户端异常关闭），则打印错误信息，并继续循环，等待下一个连接
		}

		//handler
		//!处理连接，一旦服务器接受到客户端的连接，就需要用特定的方法来处理这个连接（比如处理请求，发送响应，执行某些任务等）。
		go this.Handler(conn) //启动一个 goroutine 来处理这个连接,这段代码会在一个独立的执行线程中运行，不会阻塞当前的主线程。每个连接会并发地交由 Handler 方法处理。
	}

}
