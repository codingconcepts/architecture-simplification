package main

import (
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
	time.Sleep(time.Second * 20)

	stockDB, err := pgxpool.New(context.Background(), "postgres://stock@stock_db:26257/stock?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer stockDB.Close()

	productDB, err := pgxpool.New(context.Background(), "postgres://stock@stock_db:26257/product?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer productDB.Close()

	rawReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"redpanda:29092"},
		GroupID:     uuid.NewString(),
		Topic:       "products",
		StartOffset: kafka.LastOffset,
	})

	go updateProducts(productDB, rawReader)

	router := fiber.New()
	router.Get("/stock/:id", handleGetStock(stockDB, productDB))

	log.Fatal(router.Listen(":3000"))
}

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Stock struct {
	ProductID      string `json:"product_id"`
	QuantityOnHand int    `json:"quantity_on_hand"`
}

func handleGetStock(stock, product *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID := c.Params("id")
		if productID == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "missing product id")
		}

		// Fetch stock data.
		const stockStmt = `SELECT quantity_on_hand FROM stock WHERE product_id = $1`

		row := stock.QueryRow(c.Context(), stockStmt, productID)

		s := Stock{
			ProductID: productID,
		}
		if err := row.Scan(&s.QuantityOnHand); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// Fetch product data.
		const productStmt = `SELECT name FROM products WHERE id = $1`

		row = product.QueryRow(c.Context(), productStmt, productID)

		var p Product
		if err := row.Scan(&p.Name); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"id":       productID,
			"name":     p.Name,
			"quantity": s.QuantityOnHand,
		})
	}
}

func updateProducts(db *pgxpool.Pool, reader *kafka.Reader) {
	for {
		// Wait for message and parse it once received.
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading product message: %v", err)
			continue
		}

		var p Product
		if err = json.Unmarshal(msg.Value, &p); err != nil {
			log.Printf("error parsing product message: %v", err)
			continue
		}

		// Write the update to the local database.
		const insertStmt = `UPSERT INTO products (id, name) VALUES ($1, $2)`

		if _, err = db.Exec(context.Background(), insertStmt, p.ID, p.Name); err != nil {
			log.Printf("error upserting product: %v", err)
			continue
		}
	}
}
