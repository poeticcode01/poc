package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type OrderEvent struct {
	OrderID   string `json:"order_id"`
	EventType string `json:"event_type"`
	Amount    int    `json:"amount"`
	Timestamp string `json:"timestamp"`
}

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0

	config.Metadata.AllowAutoTopicCreation = false // Ensure topics are created manually
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("Failed to start producer:", err)
	}
	defer producer.Close()

	events := []OrderEvent{
		{"ORD2", "ORDER_CREATED", 2500, time.Now().Format(time.RFC3339)},
		{"ORD2", "ORDER_PAID", 2500, time.Now().Format(time.RFC3339)},
		{"ORD3", "ORDER_CREATED", 1800, time.Now().Format(time.RFC3339)},
	}

	for _, event := range events {
		value, _ := json.Marshal(event)

		msg := &sarama.ProducerMessage{
			Topic: "orders",
			Key:   sarama.StringEncoder(event.OrderID),
			Value: sarama.ByteEncoder(value),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Println("Failed to send message:", err)
			continue
		}

		fmt.Printf(
			"Produced event %s | partition=%d offset=%d\n",
			event.EventType, partition, offset,
		)
	}
}
