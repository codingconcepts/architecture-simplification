package main

import (
	"context"
	"encoding/json"
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

	rawReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     uuid.NewString(),
		Topic:       "raw",
		StartOffset: kafka.LastOffset,
	})

	transformedWriter := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  "transformed",
		AllowAutoTopicCreation: true,
	}
	defer transformedWriter.Close()

	transformedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     uuid.NewString(),
		Topic:       "transformed",
		StartOffset: kafka.LastOffset,
	})

	go simulateProducer(db, *writeInterval)
	go simulateConsumer(transformedReader)
	simulateETL(rawReader, transformedWriter)
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

		// Parse message to determine total flight time.
		var a after
		if err = json.Unmarshal(m.Value, &a); err != nil {
			log.Printf("error parsing message: %v", err)
			continue
		}

		ts := time.Unix(a.Timestamp, 0)
		log.Printf("[transformed %s] %s", time.Since(ts), string(m.Value))
	}
}

type before struct {
	After struct {
		OrderID    string    `json:"order_id"`
		ProductID  string    `json:"product_id"`
		CustomerID string    `json:"customer_id"`
		Quantity   int       `json:"quantity"`
		Price      float64   `json:"price"`
		Timestamp  time.Time `json:"ts"`
	} `json:"after"`
}

type after struct {
	OrderID   string `json:"-"`
	Quantity  int    `json:"quantity"`
	Price     int64  `json:"price"` // Integer
	Timestamp int64  `json:"ts"`    // Epoch
}

func simulateETL(reader *kafka.Reader, writer *kafka.Writer) error {
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}
		log.Printf("[raw] %s", string(m.Value))

		a, err := transform(m)
		if err != nil {
			log.Printf("error transforming message: %v", err)
			continue
		}

		abytes, err := json.Marshal(a)
		if err != nil {
			log.Printf("error marshalling transformed message: %v", err)
			continue
		}

		out := kafka.Message{
			Key:   []byte(a.OrderID),
			Value: abytes,
		}
		if err = writer.WriteMessages(context.Background(), out); err != nil {
			log.Printf("writing transformed message: %v", err)
			continue
		}
	}
}

func transform(m kafka.Message) (after, error) {
	var b before
	if err := json.Unmarshal(m.Value, &b); err != nil {
		return after{}, fmt.Errorf("parsing before")
	}

	return after{
		Quantity:  b.After.Quantity,
		Price:     int64(b.After.Price * 100),
		Timestamp: b.After.Timestamp.Unix(),
	}, nil
}
