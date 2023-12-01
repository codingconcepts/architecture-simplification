package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func main() {
	productClient := newClient("http://products:3000")
	stockClient := newClient("http://stock:3000")

	router := fiber.New()
	router.Get("/products/:id", handleGetProduct(productClient, stockClient))

	log.Fatal(router.Listen(":3000"))
}

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Stock struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func handleGetProduct(products, stock *http.Client) fiber.Handler {
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

		// Fetch stock from stock service.
		sresp, err := stock.Get(fmt.Sprintf("/stock/%s", productID))
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		var stock Stock
		if err = json.NewDecoder(sresp.Body).Decode(&stock); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		defer sresp.Body.Close()

		return c.JSON(fiber.Map{
			"product_id": productID,
			"stock":      stock.Quantity,
		})
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
