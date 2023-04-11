package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 创建HTTP客户端
	client := &http.Client{}

	// 准备请求数据
	data := []byte("meetingroom02_light001")

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://10.176.34.90:9312/structure/device/query", bytes.NewBuffer(data))
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

	var responseMap map[string]interface{}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	err = json.Unmarshal(respBody, &responseMap)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		return
	}

	fmt.Printf("data %v\n", responseMap)
	// 处理解析结果
	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		if meetingroom02_light001, ok := data["meetingroom02_light001"].(map[string]interface{}); ok {
			if state, ok := meetingroom02_light001["state"].(map[string]interface{}); ok {
				if light, ok := state["current_light_state"].(string); ok {
					fmt.Printf("current_light_state %v\n", light)
				} else {
					fmt.Println("Error parsing current_light_state.")
				}
			} else {
				fmt.Println("Error parsing state value.")
			}
		} else {
			fmt.Println("Error parsing meetingroom02_light001.")
		}
	} else {
		fmt.Println("Error parsing data.")
	}

}
