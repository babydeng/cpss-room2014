package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/streadway/amqp"
)

func getLight() (int, error) {
	light := 0
	// 创建HTTP客户端
	client := &http.Client{}

	// 准备请求数据
	data := []byte("meetingroom02_light001")

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://10.176.34.90:9312/structure/device/query", bytes.NewBuffer(data))
	if err != nil {
		// fmt.Printf("Failed to create request: %v\n", err)
		return light, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		// fmt.Printf("Failed to send request: %v\n", err)
		return light, err
	}
	defer resp.Body.Close()

	var responseMap map[string]interface{}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Println("Error reading response body:", err)
		return light, err
	}
	err = json.Unmarshal(respBody, &responseMap)
	if err != nil {
		// fmt.Println("Error decoding response body:", err)
		return light, err
	}

	fmt.Printf("data %v\n", responseMap)
	// 处理解析结果
	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		if meetingroom02_light001, ok := data["meetingroom02_light001"].(map[string]interface{}); ok {
			if state, ok := meetingroom02_light001["state"].(map[string]interface{}); ok {
				if lightStr, ok := state["current_light_state"].(string); ok {
					lightInt, err := strconv.ParseInt(lightStr, 10, 32)
					if err != nil {
						fmt.Println("Error parsing light value:", err)
						return light, err
					}
					return int(lightInt), nil
					// fmt.Printf("current_light_state %v\n", lightStr)
				} else {
					return light, fmt.Errorf("error parsing current_light_state")
				}
			} else {
				return light, fmt.Errorf("error parsing state")
			}
		} else {
			return light, fmt.Errorf("error parsing meetingroom02_light001")
		}
	} else {
		return light, fmt.Errorf("error parsing data")
	}
}

func changeLight(selection string) error {

}

func main() {
	conn, err := amqp.Dial("amqp://admin:admin@10.177.29.226:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test_queue", // 队列名称
		false,        // 持久化
		false,        // 自动删除
		false,        // 独占队列
		false,        // 不等待服务器响应
		nil,          // 额外的参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// 发送消息
	msg := "Hello, RabbitMQ!"
	err = ch.Publish(
		"",     // 交换机名称
		q.Name, // 目标队列名称
		false,  // 不强制使用
		false,  // 不等待服务器响应
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}
	log.Printf("Sent message: %s", msg)

	// 接收消息
	msgs, err := ch.Consume(
		q.Name, // 目标队列名称
		"",     // 消费者名称
		true,   // 自动确认
		false,  // 独占队列
		false,  // 不等待服务器响应
		false,  // 额外参数
		nil,    // 额外选项
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 处理消息
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
