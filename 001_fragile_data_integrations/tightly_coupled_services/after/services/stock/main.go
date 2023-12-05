package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	time.Sleep(time.Second * 20)

	db, err := pgxpool.New(context.Background(), "postgres://stock@stock_db:26257/stock?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	productClient := newClient("http://products:3000")

	router := fiber.New()
	router.Get("/stock/:id", handleGetStock(db, productClient))

	log.Fatal(router.Listen(":3000"))
}

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Stock struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

func handleGetStock(db *pgxpool.Pool, products *http.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID := c.Params("id")
		if productID == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "missing product id")
		}

		// Fetch products from products service.
		presp, err := products.Get(fmt.Sprintf("/products/%s", productID))
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		var product Product
		if err = json.NewDecoder(presp.Body).Decode(&product); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		defer presp.Body.Close()

		// Fetch stock from database.
		const stmt = `SELECT quantity_on_hand FROM stock WHERE product_id = $1`

		row := db.QueryRow(c.Context(), stmt, productID)

		s := Stock{
			ProductID:   productID,
			ProductName: product.Name,
		}
		if err := row.Scan(&s.Quantity); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(s)
	}
}

func newClient(baseURL string) *http.Client {
	transport := &urlPrefixTransport{
		BaseURL: baseURL,
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}

type urlPrefixTransport struct {
	BaseURL string
}

func (t *urlPrefixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base, err := url.Parse(t.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %w", err)
	}

	u, err := url.Parse(req.URL.String())
	if err != nil {
		return nil, fmt.Errorf("parsing request url: %w", err)
	}

	u = base.ResolveReference(u)
	req.URL = u
	req.Header.Add("Content-Type", "application/json")

	return http.DefaultTransport.RoundTrip(req)
}
