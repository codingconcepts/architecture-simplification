package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	time.Sleep(time.Second * 20)

	db, err := pgxpool.New(context.Background(), "postgres://stock@cockroachdb:26257/stock?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	router := fiber.New()
	router.Get("/stock/:id", handleGetStock(db))

	log.Fatal(router.Listen(":3000"))
}

type Stock struct {
	ProductID      string `json:"product_id"`
	QuantityOnHand int    `json:"quantity_on_hand"`
}

func handleGetStock(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID := c.Params("id")
		if productID == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "missing product id")
		}

		const stmt = `SELECT quantity_on_hand FROM stock WHERE product_id = $1`

		row := db.QueryRow(c.Context(), stmt, productID)

		p := Stock{
			ProductID: productID,
		}
		if err := row.Scan(&p.QuantityOnHand); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"id":       p.ProductID,
			"quantity": p.QuantityOnHand,
		})
	}
}
