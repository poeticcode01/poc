package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type OrderEvent struct {
	OrderID   string `json:"order_id"`
	EventType string `json:"event_type"`
	Amount    int    `json:"amount"`
	Timestamp string `json:"timestamp"`
}

type Consumer struct{}

func (Consumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (Consumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for msg := range claim.Messages() {
		var event OrderEvent
		_ = json.Unmarshal(msg.Value, &event)

		if event.EventType == "ORDER_CREATED" {
			log.Printf(
				"[EMAIL] order=%s partition=%d offset=%d",
				event.OrderID, msg.Partition, msg.Offset,
			)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(
		[]string{"localhost:9092"},
		"email-cg",
		config,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer group.Close()

	ctx := context.Background()
	consumer := Consumer{}

	for {
		err := group.Consume(ctx, []string{"orders"}, consumer)
		if err != nil {
			log.Println("consume error:", err)
		}
	}
}
