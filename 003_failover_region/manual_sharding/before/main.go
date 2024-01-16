package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

func main() {
	url := flag.String("url", "", "connection string")
	flag.Parse()

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	work(db, "uk", 5)
}

func work(db *pgxpool.Pool, country string, customers int) error {
	var eg errgroup.Group

	for i := 0; i < customers; i++ {
		eg.Go(func() error {
			if err := simulateCustomer(db, country); err != nil {
				return fmt.Errorf("creating customer: %w", err)
			}
			return nil
		})
	}

	eg.Wait()
	return fmt.Errorf("didn't expect to finish work function")
}

func simulateCustomer(db *pgxpool.Pool, country string) error {
	crr := createCustomerRequest{
		ID:    uuid.NewString(),
		Email: gofakeit.Email(),
	}
	if err := createCustomer(db, crr); err != nil {
		return fmt.Errorf("creating customer: %w", err)
	}
	log.Println("created customer")

	for range time.NewTicker(time.Second).C {
		cor := createOrderRequest{
			CustomerID: crr.ID,
			Amount:     rand.Float64() * 100,
		}
		if err := createOrder(db, cor); err != nil {
			log.Printf("error creating order: %v", err)
		}
		log.Println("created order")
	}

	return nil
}

type createCustomerRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func createCustomer(db *pgxpool.Pool, r createCustomerRequest) error {
	const stmt = `INSERT INTO customer (id, email) VALUES ($1, $2)`

	if _, err := db.Exec(context.Background(), stmt, r.ID, r.Email); err != nil {
		log.Printf("creating customer: %v", err)
		return fmt.Errorf("creating customer: %w", err)
	}

	return nil
}

type createOrderRequest struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

func createOrder(db *pgxpool.Pool, r createOrderRequest) error {
	const stmt = `INSERT INTO purchase (customer_id, amount) VALUES ($1, $2)`

	if _, err := db.Exec(context.Background(), stmt, r.CustomerID, r.Amount); err != nil {
		return fmt.Errorf("creating order: %w", err)
	}

	return nil
}
