package main

import (
	"net"
)

type User struct {
	Name string
	Addr string      //当前客户端所在ip地址
	C    chan string //当前和用户绑定的channel，每一个用户都有一个channel。用于与其他 goroutine 传递消息。消息通过该通道发送，最终会被写入到与客户端的连接中。
	conn net.Conn    //当前用户与客户端的网络连接。通过它可以与客户端进行数据传输。
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
func NewUser(conn net.Conn) *User { //!函数接受一个 net.Conn 类型的连接作为参数，允许 User 对象与客户端进行有效的通信，读写数据，获取连接的远程地址

	userAddr := conn.RemoteAddr().String() //获取与客户端的远程连接地址（IP与端口）并转化为string

	user := &User{
		Name: userAddr, // 默认信息为当前客户端地址
		Addr: userAddr,
		C:    make(chan string), //! make 是 Go 的内建函数，用于分配内存并初始化切片、映射（map）或通道（channel）。(创建一个通道，用于在不同的 goroutine 之间传递消息)
		conn: conn,
	}

	//启动监听当前user channel消息的goroutine
	//? 每个user都应该有个goroutine用来不断监听是否有新消息
	go user.ListenMessage()

	return user
}
