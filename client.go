package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

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

	// fmt.Printf("data %v\n", responseMap)
	// 处理解析结果
	if data, ok := responseMap["data"].(map[string]interface{}); ok {
		if meetingroom02_light001, ok := data["meetingroom02_light001"].(map[string]interface{}); ok {
			if state, ok := meetingroom02_light001["state"].(map[string]interface{}); ok {
				if lightStr, ok := state["current_light_state"].(string); ok {
					lightInt, err := strconv.ParseInt(lightStr, 10, 32)
					if err != nil {
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
	// 创建HTTP客户端
	client := &http.Client{}

	// 构造POST请求
	requestBody, _ := json.Marshal(map[string]string{
		"selection": selection,
	})

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://10.177.11.124:5400/led", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Failed to create request: %v\n", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send request: %v", resp.Status)
	}

	return nil
}

func main() {
	now := time.Now()
	hour := now.Hour()
	fmt.Println("Time: ", hour)
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

	// 创建Queue
	q, err := ch.QueueDeclare(
		"Person_Leave.meetingroom02", // Queue名称
		true,                         // 持久化标志
		false,                        // 自动删除标志
		false,                        // 独占标志
		false,                        // No-wait标志
		nil,                          // 参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// 接收消息
	msgs, err := ch.Consume(
		q.Name,             // Queue名称
		"person_detection", // Consumer名称
		true,               // Auto-ack标志
		false,              // Exclusive标志
		false,              // No-local标志
		false,              // No-wait标志
		nil,                // 参数
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 处理消息
	for msg := range msgs {
		var threshold int
		now := time.Now()
		hour := now.Hour()
		if hour >= 6 && hour < 18 {
			threshold = 200
		} else {
			threshold = 100
		}
		// 解析JSON消息
		var data map[string]interface{}
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		num := 0
		for {
			num++
			if num > 2 {
				break
			}
			light, err := getLight()
			if err != nil {
				fmt.Println("Error getting light value:", err)
			}
			fmt.Printf("Light: %d\n", light)
			if light < threshold {
				fmt.Println("Light is below threshold.")
				break
			}
			err = changeLight("left")
			if err != nil {
				fmt.Println("Error changing light:", err)
				return
			}
			fmt.Println("Changed light (left)")
			time.Sleep(30 * time.Second)
			light, err = getLight()
			if err != nil {
				fmt.Println("Error getting light value:", err)
			}
			fmt.Printf("Light: %d\n", light)

			if light < threshold {
				fmt.Println("Light is below threshold.")
				break
			}

			err = changeLight("right")
			fmt.Println("Changed light (right)")
			if err != nil {
				fmt.Println("Error changing light:", err)
				return
			}
			time.Sleep(30 * time.Second)

		}
	}
}
