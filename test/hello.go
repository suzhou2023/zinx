package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"zinx/utils"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	jsonPath := filepath.Join(rootDir, "conf/zinx.json")
	fmt.Println(jsonPath)

	// 打开 JSON 文件
	file, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)

	var config utils.GlobalObj
	// 使用解码器解码 JSON 数据
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println("config.WorkerPoolSize:", config.WorkerPoolSize)
}
