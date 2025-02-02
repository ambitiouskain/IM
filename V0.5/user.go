package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string      //当前客户端所在ip地址
	C    chan string //当前和用户绑定的channel，每一个用户都有一个channel。用于与其他 goroutine 传递消息。消息通过该通道发送，最终会被写入到与客户端的连接中。
	conn net.Conn    //当前用户与客户端的网络连接。通过它可以与客户端进行数据传输。

	server *Server //绑定serve，方便实用server里的方法
}

// * 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端(对应的那个客户端)
func (this *User) ListenMessage() {
	for {
		msg := <-this.C //从channel读取数据并保存为msg

		this.conn.Write([]byte(msg + "\n")) //this.conn.Write 用于将数据写入网络连接（即发送数据给客户端）。

		// ! []byte 用于类型转换操作，将一个字符串转换为字节切片（[]byte）。字节切片是 Go 中用于表示原始数据的类型，可以用来进行二进制数据的处理或网络传输。
	}

}

// * 创建一个用户的API，方便创建User实例
func NewUser(conn net.Conn, server *Server) *User {
	//!  传入 net.Conn 类型的连接以及*Server作为参数，使 User 对象与客户端进行有效的通信，读写数据，获取连接的远程地址

	userAddr := conn.RemoteAddr().String() //获取与客户端的远程连接地址（IP与端口）并转化为string

	user := &User{
		Name: userAddr, // 默认信息为当前客户端地址
		Addr: userAddr,
		C:    make(chan string), //! make 是 Go 的内建函数，用于分配内存并初始化切片、映射（map）或通道（channel）。(创建一个通道，用于在不同的 goroutine 之间传递消息)
		conn: conn,

		server: server,
	}

	//启动监听当前user channel消息的goroutine
	//? 每个user都应该有个goroutine用来不断监听是否有新消息
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (this *User) Online() {
	//用户上线，将用户加入到onlineMap中
	this.server.mapLock.Lock()              //? 先上锁OnlineMap，防止数据竞争
	this.server.OnlineMap[this.Name] = this // 将用户对象user添加到OnlineMap中，user.Name作为key，user作为值
	this.server.mapLock.Unlock()            //? 解锁，释放

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")

}

// 用户的下线业务
func (this *User) Offline() {
	//用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()               //? 先上锁OnlineMap，防止数据竞争
	delete(this.server.OnlineMap, this.Name) // 将用户对象user删除
	this.server.mapLock.Unlock()             //? 解锁，释放

	//广播当前用户下线消息
	this.server.BroadCast(this, "已下线")
}

// 给当前User关联的客户端发送消息
func (this *User) SendMsg(msg string) {
	_, err := this.conn.Write([]byte(msg)) //? 将 string 类型的消息 msg 转换为字节切片（[]byte），因为网络传输需要发送的是字节流，而不是直接的字符串。
	if err != nil {
		fmt.Println("Error sending message:", err)

	}
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	//添加用户查询功能查询，如果user输入“who”消息，则查询当前在线用户并（只向查询者）返回结果
	if msg == "who" {

		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "当前在线...\n"
			//将查询结果返回给查询者，即只向查询者广播
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else {
		//否则正常广播消息
		this.server.BroadCast(this, msg)
	}
}
