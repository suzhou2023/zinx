package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Connection struct {
	//当前Conn属于哪个Server，在Conn初始化的时候添加即可
	TcpServer ziface.IServer

	//当前连接的socket TCP套接字
	Conn *net.TCPConn

	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32

	//当前连接的关闭状态
	isClosed bool

	//MsgId和对应处理方法的消息管理模块
	msgHandler ziface.IMsgHandle

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte
}

// 创建连接的方法
func NewConntion(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		msgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObj.MaxMsgChanLen),
	}

	//将当前新创建的连接添加到ConnManager中
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

/* 处理conn读数据的Goroutine */
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		// 创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitBuffChan <- true
			continue
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObj.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.msgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.msgHandler.DoMsgHandler(&req)
		}
	}
}

/*
写消息Goroutine，用户将数据发送给客户端
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}

			//针对有缓冲channel需要些的数据处理
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				break
				fmt.Println("msgBuffChan is Closed")
			}

		case <-c.ExitBuffChan:
			//conn已经关闭
			return
		}
	}
}

// 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg := NewMsgPackage(msgId, data)
	fmt.Println("msg.DataLen:", msg.DataLen)

	bytes, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取
	c.msgChan <- bytes

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgBuffChan <- msg

	return nil
}

// 启动连接，让当前连接开始工作
func (c *Connection) Start() {

	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()

	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()

	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			//得到退出消息，不再阻塞
			return
		}
	}
}

// 停止连接，结束当前连接状态
func (c *Connection) Stop() {
	//1. 如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该链接已经关闭
	c.ExitBuffChan <- true

	//将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c) //删除conn从ConnManager中

	//关闭该链接全部管道
	close(c.ExitBuffChan)
}

// 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
