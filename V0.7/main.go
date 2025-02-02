package main

// main与server拆开写的目的：模块基本划分
// main作为主入口
func main() {
	server := NewServer("127.0.0.1", 8766) //实例化
	server.Start()                         //启动
}
