package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"zinx/ziface"
)

/*
存储一切有关Zinx框架的全局参数，供其他模块使用
一些参数也可以通过 用户根据 zinx.json来配置
*/
type GlobalObject struct {
	TcpServer ziface.IServer //当前Zinx的全局Server对象
	Host      string         //当前服务器主机IP
	TcpPort   int            //当前服务器主机监听端口号
	Name      string         //当前服务器名称
	Version   string         //当前Zinx版本号

	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 // 工作池数量
	MaxWorkerTaskLen uint32 // 最大任务数
	MaxMsgChanLen    uint32
}

/*
定义一个全局的对象
*/
var GlobalObj *GlobalObject

// 读取用户的配置文件
func (g *GlobalObject) Reload() {
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	jsonPath := filepath.Join(rootDir, "conf/zinx.json")

	// 打开 JSON 文件
	file, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)

	// 使用解码器解码 JSON 数据
	if err := decoder.Decode(&GlobalObj); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
}

/*
提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObj = &GlobalObject{
		Name:          "ZinxServerApp",
		Version:       "V0.4",
		TcpPort:       7777,
		Host:          "127.0.0.1",
		MaxConn:       100,
		MaxPacketSize: 4096,
	}

	//从配置文件中加载一些用户配置的参数
	GlobalObj.Reload()
}
