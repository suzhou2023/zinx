package main

import (
	"fmt"
	"time"
)

func main() {
	// 创建一个大小为 3 的整数通道
	intChannel := make(chan int, 3)

	// 启动三个协程，每个协程往通道写入一个整数
	go func() {
		intChannel <- 1
	}()
	go func() {
		intChannel <- 2
	}()
	go func() {
		intChannel <- 3
	}()

	// 主协程等待一段时间
	time.Sleep(time.Second)

	// 从通道中读取数据
	fmt.Println(<-intChannel)
	fmt.Println(<-intChannel)
	fmt.Println(<-intChannel)

}
