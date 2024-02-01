package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandler struct {
	//存放每个MsgId所对应的处理方法的map属性
	Routers map[uint32]ziface.IRouter

	//业务工作Worker池的数量
	WorkerPoolSize uint32

	//Worker负责取任务的消息队列
	TaskQueue []chan ziface.IRequest
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Routers:        make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObj.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObj.WorkerPoolSize),
	}
}

// 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " request msgID=", request.GetMsgID(), "to workerID=", workerID)
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

// 启动worker工作池
func (mh *MsgHandler) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObj.MaxWorkerTaskLen)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个Worker工作流程
func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.Handle(request)
		}
	}
}

// 马上以非阻塞方式处理消息
func (mh *MsgHandler) Handle(request ziface.IRequest) {
	router, ok := mh.Routers[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	//执行对应处理方法
	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandler) AddRouter(msgId uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的router处理方法是否已经存在
	if _, ok := mh.Routers[msgId]; ok {
		panic("repeated router , msgId = " + strconv.Itoa(int(msgId)))
	}

	//2 添加msg与api的绑定关系
	mh.Routers[msgId] = router
	fmt.Println("Add router msgId:", msgId)
}
