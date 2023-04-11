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

type Entity struct {
	EntityID string `json:"entity_id"`
}

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhYmM1OGZiMTE5NDM0MjNmOGE3NTJiMjlhOGFhMWM5OSIsImlhdCI6MTY3NzgwODE4NywiZXhwIjoxOTkzMTY4MTg3fQ.Y-SZk5JoyBEpSNKbCzLiHhAUt4cwxTgeAyLEg8cqOIU"
var switchDisplay = "switch.cuco_v3_f645_switch_2"

func SetSwitch(entityId string, state string) error {
	url := "http://10.177.11.124:8123/api/services/switch/turn_" + state
	entity := Entity{EntityID: entityId}

	jsonData, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// 创建HTTP客户端
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	return nil
}

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

				if queueName == "Person_Entry.meetingroom02" {
					// do something
					go SetSwitch(switchDisplay, "on")

				}

				if queueName == "Person_Leave.meetingroom02" {
					// do something
					go SetSwitch(switchDisplay, "off")

					var threshold int
					now := time.Now()
					hour := now.Hour()
					if hour >= 6 && hour < 18 {
						threshold = 200
					} else {
						threshold = 100
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
		}()
	}

	// 阻塞程序不退出
	select {}
}
