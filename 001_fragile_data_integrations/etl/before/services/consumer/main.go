package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetFlags(0)

	db, err := pgxpool.New(context.Background(), "postgres://root@localhost:26257/defaultdb?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging db: %v", err)
	}

	transformedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     uuid.NewString(),
		Topic:       "transformed",
		StartOffset: kafka.LastOffset,
	})

	simulateConsumer(transformedReader)
}

func simulateConsumer(reader *kafka.Reader) error {
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		// Parse message to determine total flight time.
		var a after
		if err = json.Unmarshal(m.Value, &a); err != nil {
			log.Printf("error parsing message: %v", err)
			continue
		}

		ts := time.Unix(a.Timestamp, 0)
		log.Printf("\n[transformed %s] %s", time.Since(ts), string(m.Value))
	}
}

type after struct {
	OrderID   string `json:"-"`
	Quantity  int    `json:"quantity"`
	Price     int64  `json:"price"` // Integer
	Timestamp int64  `json:"ts"`    // Epoch
}
