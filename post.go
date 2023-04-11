package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	// 创建HTTP客户端
	client := &http.Client{}

	// 准备请求数据
	data := []byte(`{"selection": "right"}`)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://10.177.11.124:5400/led", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to send request: %v\n", resp.Status)
		return
	}

	fmt.Println("Request sent successfully")
}
