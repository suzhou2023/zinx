package ziface

/*
消息管理抽象层
*/
type IMsgHandle interface {
	//马上以非阻塞方式处理消息
	DoMsgHandler(request IRequest)

	//为消息添加具体的处理逻辑
	AddRouter(msgId uint32, router IRouter)

	//启动worker工作池
	StartWorkerPool()

	//将消息交给TaskQueue,由worker进行处理
	SendMsgToTaskQueue(request IRequest)
}
