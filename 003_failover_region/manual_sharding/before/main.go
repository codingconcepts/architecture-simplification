package main

import (
	"context"
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := flag.String("url", "", "connection string")
	flag.Parse()

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	router := fiber.New()
	router.Post("/customers", createCustomer(db))
	router.Post("/orders", createOrder(db))

	log.Fatal(router.Listen(":3000"))
}

type createCustomerRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func createCustomer(db *pgxpool.Pool) fiber.Handler {
	const stmt = `INSERT INTO customer (id, email) VALUES ($1, $2)`

	return func(c *fiber.Ctx) error {
		var req createCustomerRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("invalid create customer request: %v", err)
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid request")
		}

		if _, err := db.Exec(c.Context(), stmt, req.ID, req.Email); err != nil {
			log.Printf("creating customer: %v", err)
			return fiber.NewError(fiber.StatusUnprocessableEntity, "creating customer")
		}

		return nil
	}
}

type createOrderRequest struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

func createOrder(db *pgxpool.Pool) fiber.Handler {
	const stmt = `INSERT INTO purchase (customer_id, amount) VALUES ($1, $2)`

	return func(c *fiber.Ctx) error {
		var req createOrderRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("invalid create order request: %v", err)
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid request")
		}

		if _, err := db.Exec(c.Context(), stmt, req.CustomerID, req.Amount); err != nil {
			log.Printf("creating order: %v", err)
			return fiber.NewError(fiber.StatusUnprocessableEntity, "creating order")
		}

		return nil
	}
}
