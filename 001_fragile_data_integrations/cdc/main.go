package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

var (
	paymentsMu sync.Mutex
	payments   = make(map[string]time.Time)

	messagesPublished uint64
	avgDelay          = rollingAverage(10)
)

func main() {
	url := flag.String("url", "", "connection string")
	flag.Parse()

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     uuid.NewString(),
		Topic:       "events.public.payment",
		StartOffset: kafka.LastOffset,
	})

	go consumePayments(kafkaReader)
	work(db)
}

func work(db *pgxpool.Pool) {
	for range time.NewTicker(time.Millisecond * 100).C {
		if err := insertPayment(db); err != nil {
			log.Printf("error inserting payment: %v", err)
		}
	}
}

func insertPayment(db *pgxpool.Pool) error {
	paymentsMu.Lock()
	defer paymentsMu.Unlock()

	id := uuid.NewString()
	now := time.Now()
	payments[id] = now

	const stmt = `INSERT INTO payment (id, amount, ts) VALUES ($1, $2, $3)`
	if _, err := db.Exec(context.Background(), stmt, id, round(rand.Float64()*100, 2), now); err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}

	atomic.AddUint64(&messagesPublished, 1)
	return nil
}

func round(val float64, precision int) float64 {
	return math.Round(val*(math.Pow10(precision))) / math.Pow10(precision)
}

type paymentEvent struct {
	After struct {
		ID     string    `json:"id"`
		Amount string    `json:"amount"`
		Ts     time.Time `json:"ts"`
	} `json:"after"`
}

func consumePayments(reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		if err = compareAndPrint(msg); err != nil {
			log.Printf("error comparing message: %v", err)
			continue
		}
	}
}

func compareAndPrint(msg kafka.Message) error {
	var pe paymentEvent
	if err := json.Unmarshal(msg.Value, &pe); err != nil {
		return fmt.Errorf("parsing event: %w", err)
	}

	diff := time.Since(pe.After.Ts)
	fmt.Printf("average delay: %s (%d messages)\n", avgDelay(diff), atomic.LoadUint64(&messagesPublished))
	return nil
}

func rollingAverage(period int) func(time.Duration) time.Duration {
	var i int
	var sum time.Duration
	var storage = make([]time.Duration, 0, period)

	return func(input time.Duration) (avrg time.Duration) {
		if len(storage) < period {
			sum += input
			storage = append(storage, input)
		}

		sum += input - storage[i]
		storage[i], i = input, (i+1)%period
		avrg = sum / time.Duration(len(storage))

		return
	}
}
