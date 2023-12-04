package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetFlags(0)

	writeInterval := flag.Duration("w", time.Millisecond*10, "interval between writes")
	flag.Parse()

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
		Topic:       "transformed_2",
		StartOffset: kafka.LastOffset,
	})

	go simulateProducer(db, *writeInterval)
	simulateConsumer(transformedReader)
}

func simulateProducer(db *pgxpool.Pool, rate time.Duration) error {
	const stmt = `INSERT INTO order_line_item (order_id, product_id, customer_id, quantity, price, ts) VALUES
								(gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), $1, $2, now())`

	for range time.NewTicker(rate).C {
		if _, err := db.Exec(context.Background(), stmt, rand.Intn(10), rand.Float64()*100); err != nil {
			return fmt.Errorf("inserting event: %w", err)
		}
	}

	return fmt.Errorf("finished simulateProducer unexectedly")
}

func simulateConsumer(reader *kafka.Reader) error {
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		log.Printf("[transformed] %s", string(m.Value))
	}
}
