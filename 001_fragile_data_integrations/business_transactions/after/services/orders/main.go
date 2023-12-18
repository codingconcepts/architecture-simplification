package main

import (
	"architecture_simplification/001_fragile_data_integrations/business_transactions/after/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetFlags(0)
	time.Sleep(time.Second * 20)

	db, err := pgxpool.New(context.Background(), "postgres://orders_user@cockroachdb:26257/orders?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	go handlePaymentMessages(kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"redpanda:29092"},
		GroupID:     uuid.NewString(),
		Topic:       "payments",
		StartOffset: kafka.LastOffset,
	}))

	router := fiber.New()
	router.Post("/orders", handlePlaceOrder(db))

	log.Fatal(router.Listen(":3000"))
}

type order struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Total        float64         `json:"total"`
	FailureFlags json.RawMessage `json:"failure_flags"`
}

func handlePlaceOrder(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var o order
		if err := c.BodyParser(&o); err != nil {
			log.Printf("invalid order: %v", err)
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid order")
		}

		ff, err := models.JSONToFailureFlags(o.FailureFlags)
		if err != nil {
			log.Printf("reading failure flags: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "reading failure flags")
		}

		if ff.Order {
			log.Printf("order failure: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "order failure")
		}

		const stmt = `INSERT INTO orders (id, user_id, total, failures) VALUES ($1, $2, $3, $4)`
		if _, err := db.Exec(context.Background(), stmt, o.ID, o.UserID, o.Total, o.FailureFlags); err != nil {
			log.Printf("error storing order: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "error storing order")
		}

		return nil
	}
}

func handlePaymentMessages(reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		log.Printf("[payment msg]: %s", string(msg.Value))
	}
}
