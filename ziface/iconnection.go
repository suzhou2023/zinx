package ziface

import "net"

// 定义连接接⼝
type IConnection interface {
	// 启动连接，让当前连接开始⼯工作
	Start()

	// 停⽌止连接，结束当前连接状态M
	Stop()

	// 从当前连接获取原始的socket TCPConn
	GetTCPConnection() *net.TCPConn

	//获取当前连接ID
	GetConnID() uint32

	//获取远程客户端地址信息
	RemoteAddr() net.Addr

	//直接将Message数据发送数据给远程的TCP客户端
	SendMsg(msgId uint32, data []byte) error

	//添加带缓冲发送消息接口
	SendBuffMsg(msgId uint32, data []byte) error
}

// 定义⼀一个统⼀一处理理链接业务的接⼝口
type HandFunc func(*net.TCPConn, []byte, int) error
