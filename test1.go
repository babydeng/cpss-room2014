package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

func queryLight() {
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

func main() {
	// 连接RabbitMQ服务器
	conn, err := amqp.Dial("amqp://admin:admin@10.177.29.226:5672/em_vhost")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 创建Channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 创建Exchange
	err = ch.ExchangeDeclare("Person_Entry.meetingroom02", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	// 创建Queue
	_, err = ch.QueueDeclare("Person_Entry.meetingroom02", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// 绑定Queue到Exchange
	err = ch.QueueBind("Person_Entry.meetingroom02", "Person_Entry.meetingroom02", "Person_Entry.meetingroom02", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	fmt.Println("RabbitMQ setup done!")
}
