package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	data := `{
    "eventIdentify": {
        "eventId": "mixiaochao_boy",
        "name": "Person_Entry",
        "topic": "Person_Entry.meetingroom02",
        "location": "meetingroom02"
    },
    "timestamp": 1679849540,
    "perceptionEventType": 1,
    "eventData": {
        "location": "meetingroom02",
        "objectId": "plmm",
        "data": {
            "location": "meetingroom02"
        }
    }
}`
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

	// 发送消息
	err = ch.Publish(
		"Person_Entry.meetingroom02", // Exchange名称
		"Person_Entry.meetingroom02", // RoutingKey
		false,                        // Mandatory标志
		false,                        // Immediate标志
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}

	fmt.Println("Message published!")
}
