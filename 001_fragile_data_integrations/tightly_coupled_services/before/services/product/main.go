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

	db, err := pgxpool.New(context.Background(), "postgres://product@cockroachdb:26257/product?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	router := fiber.New()
	router.Get("/products/:id", handleGetProduct(db))

	log.Fatal(router.Listen(":3000"))
}

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleGetProduct(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID := c.Params("id")
		if productID == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "missing product id")
		}

		const stmt = `SELECT name FROM products WHERE id = $1`

		row := db.QueryRow(c.Context(), stmt, productID)

		p := Product{
			ID: productID,
		}
		if err := row.Scan(&p.Name); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"id":   p.ID,
			"name": p.Name,
		})
	}
}
