package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

//? 实例（Instance）是某个类型或类的具体对象。
//? 是类型的实际使用体现，通过实例化可以将抽象的定义（如结构体）变为具体可操作的对象

//! 类型 是结构体（struct）的定义，它只是一个模板或蓝图，定义了一类对象的属性和行为。
//! 实例 是通过类型创建的实际对象，用于存储和操作具体数据

// 客户端结构体
// ? 封装了客户端所需的信息和状态：
// ? 服务器连接信息：包括服务器 IP 和端口。
// ? 用户信息：如用户名。
// ? 通信通道：客户端到服务器的网络连接。
type Client struct {
	ServerIp   string
	ServerPort int
	Name       string   //当前客户端的名称（即用户名称）    //? 标识当前客户端用户，通常会被发送到服务器进行用户管理和显示
	conn       net.Conn //客户端与服务器之间的网络连接     //? 通过它与服务器进行通信，包括发送和接收数据。
	flag       int      //当前client所处模式(公聊、私聊、改名等)
}

// 用于创建客户端实例
// ! 定义 NewClient 函数的主要原因是为了封装客户端的初始化逻辑，提供一种方便、统一的方法来创建客户端对象。
func NewClient(serverIp string, serverPort int) *Client {

	//创建客户端对象Client并初始化     //? 将 serverIp 和 serverPort 存入结构体字段中。
	Client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort)) //? 以 TCP 协议连接到服务器。返回成功建立的连接：conn
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	//将成功建立的连接赋值给 Client 的 conn 字段
	Client.conn = conn

	//返回创建好的Client对象
	return Client
}

// 处理server返回的消息(比如已更新用户名信息，聊天消息等)，并自动输出显示到客户端的标准输出
// ? 从服务器端读取消息并打印到终端，实现客户端实时接收服务器发送的消息
func (client *Client) DealResponse() {
	//不断从 client.conn读取数据，并输出到os.Stdout（即终端）
	io.Copy(os.Stdout, client.conn) //! io.Copy(dst, src)是Go语言I/O数据流的快速复制函数，用于将数据从src复制到dst，直到 src 结束或发生错误。
	// ? os.Stdout 代表标准输出（控制台），即把数据输出到终端。
	// ? client.conn是 客户端的TCP 连接，它从服务器接收数据。

	/*	//相当于
		for {
			buf := make([]byte, 4096)
			msg, err := client.conn.Read(buf)
			if err != nil {
				return
			}
			fmt.Println(string(buf[:msg])) //打印服务器发送的消息
		}
	*/
}

// 显示菜单功能方法
func (client *Client) Menu() bool {
	var flag int //存储用户输入值，用于更换客户端运行模式

	//菜单模板
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("0.退出")

	//获取用户输入并存入flag
	fmt.Scanln(&flag)

	//判断flag
	if flag >= 0 && flag <= 3 {
		client.flag = flag //
		return true
	} else {
		fmt.Println(">>>>>>>>>>请输入合法范围内的数字<<<<<<<<<<<")
		return false
	}
}

// 执行用户选择的模式
func (client *Client) Run() {
	for client.flag != 0 {
		for client.Menu() != true { //循环调用menu()，直到输入正确
		}

		//根据不同模式处理不同业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
			break
		case 2:
			//私聊模式
			client.PrivateChat()
			break
		case 3:
			//更换用户名
			client.UpdateName()
			break
		}

	}
}

// 公聊模式功能方法
func (client *Client) PublicChat() {
	for {
		//提示用户输入消息
		fmt.Println("请输入聊天内容（exit 退出）。")
		var chatMsg string
		fmt.Scanln(&chatMsg)

		/* 过时
		for chatMsg != "exit" {
			//发给服务器

			//判断消息是否为空,不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break //如果发送失败，会退出循环。
				}
			}

			//发送成功后，清空 chatMsg 并提示用户再次输入
			chatMsg = ""
			fmt.Println("请输入聊天内容（exit 退出）。")
			fmt.Scanln(&chatMsg)
		}
		*/

		// 退出公聊模式
		if chatMsg == "exit" {
			fmt.Println("已退出公聊模式")
			break
		}

		//避免发送空消息
		if len(chatMsg) == 0 {
			continue //? 如果chatMsg为空，则跳过当前循环，重新提示用户输入消息。
		}

		//发送消息
		sendMsg := chatMsg + "\n" //添加换行符（\n），使服务器能正确解析消息  //!在TCP连接中，数据是以流（stream）形式传输的，服务器需要一个分隔符来判断一条消息的结束
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn Write err:", err)
			break
		}

	}
}

// 查询当前在线用户
func (client *Client) SelectUsers() {
	//根据服务端定义的“who”来查询
	sendMsg := "who\n" //服务器一般按行读取数据，\n 让服务器知道 "who" 命令已经完整传输，使服务器能正确解析消息，
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string //私聊对象用户名
	var chatMsg string    //聊天信息内容

	//先查询当前在线用户
	client.SelectUsers()
	/*过时
	//提示用户输入
	fmt.Println(">>>>>>>请输入想要私聊的对象[用户名]，exit退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>>请输入消息内容，exit退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//发给服务器

			//判断消息是否为空,不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break //如果发送失败，会退出循环。
				}
			}

			//发送成功后，清空 chatMsg 并提示用户再次输入
			chatMsg = ""
			fmt.Println(">>>>>>>请输入消息内容，exit退出：")
			fmt.Scanln(&chatMsg)
		}

		//退出当前对话，重新选择私聊对象
		client.SelectUsers()
		fmt.Println(">>>>>>>请输入想要私聊的对象[用户名]，exit退出：")
		fmt.Scanln(&remoteName)
	}
	*/

	for {
		//提示用户输入
		fmt.Println(">>>>>>>请输入想要私聊的对象[用户名]，exit退出：")
		fmt.Scanln(&remoteName)

		if remoteName == "exit" {

			break
		}
		for {
			fmt.Println(">>>>>>>请输入消息内容，exit退出：")
			fmt.Scanln(&chatMsg)
			if chatMsg == "exit" {
				client.SelectUsers()
				break
			}
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break //如果发送失败，会退出循环。
				}
			}
		}
	}
}

// 更换用户名功能
func (client *Client) UpdateName() bool {

	fmt.Println(">>>>>>>请输入新的用户名：")

	//从标准输入读取用户输入      //! 标准输入（stdin）是程序从键盘或其他输入源读取数据的方式
	fmt.Scanln(&client.Name) //&client.Name取出client.Name的地址，让 Scanln() 直接修改这个变量的值   //? client.Name 是一个string，如果要直接修改它，就要传入它的地址,故使用&（取地址符）

	//构造要发送给服务器的修改用户名请求
	sendMsg := "rename|" + client.Name + "\n"
	//把sendMsg发送到服务器
	_, err := client.conn.Write([]byte(sendMsg)) //? sendMsg是字符串需要转换成 byte数组才能通过Write()发送
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 命令行格式：./client -ip 127.0.0.1 -port 8888
var serverIp string
var serverPort int //? serverIp与serverPort用于接受用户的命令行参数

// ! init函数，会在main()之前自动执行，常用于初始化全局变量
func init() {
	//? 使用flag包来解析命令行参数
	//当用户输入./help的时候可以输出提示说明
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认为127.0.0.1）") //绑定 serverIp 变量到命令行参数 -ip，如果用户不输入，则默认为 "127.0.0.1"。
	flag.IntVar(&serverPort, "port", 8766, "设置服务器端口（默认是8766）")              //绑定 serverPort 变量到 -port，默认值为 8766
}

func main() {

	//使用flag.Parse()解析init()中 flag.StringVar 和 flag.IntVar 绑定的参数
	flag.Parse()

	//调用NewClient创建客户端，尝试连接服务器
	////client := NewClient("127.0.0.1", 8766) 这里ip与端口都写死了
	client := NewClient(serverIp, serverPort) //改成使用命令行参数
	if client == nil {
		fmt.Println(">>>>>>>连接服务器失败……")
		return
	}
	fmt.Println(">>>>>>>连接服务器成功……")

	//开启一个go routine监听服务器信息，不会阻塞客户端其他功能
	go client.DealResponse()

	//启动客户端业务
	client.Run() //启动菜单
}
