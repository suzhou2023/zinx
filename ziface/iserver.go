package ziface

// 定义服务器接口
type IServer interface {
	// 启动服务器方法
	Start()

	// 停止服务器方法
	Stop()

	// 开启业务服务方法
	Serve()

	// 路由功能：给当前服务注册一个路由业务方法
	AddRouter(msgId uint32, router IRouter)

	// 获取连接管理器
	GetConnMgr() IConnManager

	//设置该Server的连接创建时Hook函数
	SetOnConnStart(func(IConnection))

	//设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))

	//调用连接OnConnStart Hook函数
	CallOnConnStart(conn IConnection)

	//调用连接OnConnStop Hook函数
	CallOnConnStop(conn IConnection)
}
