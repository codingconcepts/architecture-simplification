package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	kafkaURL, ok := os.LookupEnv("KAFKA_URL")
	if !ok {
		log.Fatal("missing KAFKA_URL env var")
	}

	indexURL, ok := os.LookupEnv("INDEX_URL")
	if !ok {
		log.Fatalf("missing INDEX_URL env var")
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaURL},
		GroupID:     uuid.NewString(),
		Topic:       "products.store.product",
		StartOffset: kafka.LastOffset,
	})

	db, err := pgxpool.New(context.Background(), indexURL)
	if err != nil {
		log.Fatalf("error connecting to index: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error testing index connection: %v", err)
	}

	consume(kafkaReader, db)
}

type cdcEvent struct {
	Payload struct {
		Op    string `json:"op"`
		After struct {
			ID struct {
				Value      string `json:"value"`
				DeletionTs any    `json:"deletion_ts"`
				Set        bool   `json:"set"`
			} `json:"id"`
			Ts struct {
				Value      string `json:"value"`
				DeletionTs any    `json:"deletion_ts"`
				Set        bool   `json:"set"`
			} `json:"ts"`
			Description struct {
				Value      string `json:"value"`
				DeletionTs any    `json:"deletion_ts"`
				Set        bool   `json:"set"`
			} `json:"description"`
			Name struct {
				Value      string `json:"value"`
				DeletionTs any    `json:"deletion_ts"`
				Set        bool   `json:"set"`
			} `json:"name"`
		} `json:"after"`
	} `json:"payload"`
}

func consume(reader *kafka.Reader, db *pgxpool.Pool) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading event: %v", err)
			continue
		}

		var event cdcEvent
		if err = json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("error parsing event: %v", err)
			continue
		}

		if err = updateIndex(db, event); err != nil {
			log.Printf("error updating index: %v", err)
			continue
		}
	}
}

func updateIndex(db *pgxpool.Pool, e cdcEvent) error {
	const stmt = `INSERT INTO product (id, name, description, ts) VALUES ($1, $2, $3, $4)
								ON CONFLICT (id)
								DO UPDATE SET
									name = EXCLUDED.name,
									description = EXCLUDED.description`

	a := e.Payload.After
	if _, err := db.Exec(context.Background(), stmt, a.ID, a.Name, a.Description, a.Ts); err != nil {
		return fmt.Errorf("upserting product: %w", err)
	}

	return nil
}
