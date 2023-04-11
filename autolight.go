package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 如果光照低于某个值，就打开灯

func main() {
	// 连接到RabbitMQ服务器
	conn, err := amqp.Dial("amqp://admin:admin@10.177.29.226:5672/em_vhost")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 创建一个Channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 定义Queue名称
	queueNames := []string{"Person_Leave.meetingroom02", "Person_Entry.meetingroom02"}

	// 消费消息
	for _, queueName := range queueNames {
		msgs, err := ch.Consume(
			queueName, // Queue名称
			"",        // Consumer名称
			true,      // 自动应答
			false,     // 独占
			false,     // 没有等待
			false,     // 没有额外的参数
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to consume messages from queue %v: %v", queueName, err)
		}

		// 处理消息
		go func() {
			for d := range msgs {
				fmt.Printf("Received a message from queue %v: %s\n", queueName, d.Body)
			}
		}()
	}

	// 阻塞程序不退出
	select {}
}
