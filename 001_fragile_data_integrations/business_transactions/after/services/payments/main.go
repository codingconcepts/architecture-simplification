package main

import (
	"architecture_simplification/001_fragile_data_integrations/business_transactions/after/models"
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetFlags(0)
	time.Sleep(time.Second * 20)

	db, err := pgxpool.New(context.Background(), "postgres://payments_user@cockroachdb:26257/payments?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	go handleOrderMessages(kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"redpanda:29092"},
		GroupID:     uuid.NewString(),
		Topic:       "orders",
		StartOffset: kafka.LastOffset,
	}))

	go handleInventoryMessages(kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"redpanda:29092"},
		GroupID:     uuid.NewString(),
		Topic:       "inventory",
		StartOffset: kafka.LastOffset,
	}))

	runtime.Goexit()
}

type order struct {
	After struct {
		ID       string          `json:"id"`
		UserID   string          `json:"user_id"`
		Total    float64         `json:"total"`
		TS       string          `json:"ts"`
		Failures json.RawMessage `json:"failures"`
	} `json:"after"`
	Key []string `json:"key"`
}

func handleOrderMessages(reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var o order
		if err = json.Unmarshal(msg.Value, &o); err != nil {
			log.Printf("error unmarshalling order message: %v", err)
			continue
		}

		ff, err := models.JSONToFailureFlags(o.After.Failures)
		if err != nil {
			log.Printf("reading failure flags: %v", err)
		}

		if ff.Payment {
			log.Printf("payment failure: %v", err)
		}
	}
}

type inventory struct {
	After struct {
		ProductID string          `json:"product_id"`
		Quantity  int             `json:"quantity"`
		Failures  json.RawMessage `json:"failures"`
	} `json:"after"`
	Key []string `json:"key"`
}

func handleInventoryMessages(reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var i inventory
		if err = json.Unmarshal(msg.Value, &i); err != nil {
			log.Printf("error unmarshalling inventory message: %v", err)
			continue
		}

		// If inventory failed, rollback payment.
		ff, err := models.JSONToFailureFlags(i.After.Failures)
		if err != nil {
			log.Printf("reading failure flags: %v", err)
		}

		if ff.Inventory {

		}

		log.Printf("[inventory]: %+v", i)
	}
}
