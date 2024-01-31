package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)

	go func() {
		// 发送数据到管道
		ch <- 123
	}()

	// 稍后从管道接收数据
	time.Sleep(1 * time.Second)
	value := <-ch
	fmt.Println("Received:", value)

	// 关闭管道
	close(ch)
}
