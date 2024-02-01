package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	// 创建协程的数量
	numGoroutines := 100000 // 可以调整这个值来观察不同数量的协程对内存的影响

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Second) // 让协程做一些工作
		}()
	}

	// 打印当前内存占用
	printMemUsage()

	// 等待所有协程完成
	wg.Wait()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 打印内存占用情况
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
