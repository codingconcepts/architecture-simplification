package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	products = []string{
		"93410c29-1609-484d-8662-ae2d0aa93cc4",
		"47b0472d-708c-4377-aab4-acf8752f0ecb",
		"a1a879d8-58c0-4357-a570-a57c3b1fe059",
		"5ded80d3-fb55-4a2f-b339-43fc9c89894a",
		"b6afe0c5-9cab-4971-8c61-127fe5b4acd1",
		"7098227b-4883-4992-bc32-e12335efbc8c",
	}

	writeQuantities *quantities
	readQuantities  *quantities
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://root@localhost:26257/defaultdb?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging db: %v", err)
	}

	writeQuantities = &quantities{
		products: map[string]int{},
	}
	readQuantities = &quantities{
		products: map[string]int{},
	}

	go simulateReads(db, time.Millisecond*10)
	go simulateWrites(db, time.Millisecond*100)
	printLoop()
}

type quantities struct {
	productsMu sync.RWMutex
	products   map[string]int
}

func (q *quantities) set(product string, stock int) {
	q.productsMu.Lock()
	defer q.productsMu.Unlock()

	q.products[product] = stock
}

func simulateReads(db *pgxpool.Pool, rate time.Duration) error {
	for range time.NewTicker(rate).C {
		if err := simulateRead(db); err != nil {
			log.Printf("error simulating read: %v", err)
		}
	}

	return fmt.Errorf("finished simulateReads unexectedly")
}

func simulateRead(db *pgxpool.Pool) error {
	// Pick a product.
	productID := products[rand.Intn(len(products))]

	// Read stock from database.
	stock, err := readFromDB(db, productID)
	if err != nil {
		return fmt.Errorf("reading from db: %w", err)
	}

	readQuantities.set(productID, stock)

	return nil
}

func simulateWrites(db *pgxpool.Pool, rate time.Duration) error {
	stock := 0
	for range time.NewTicker(rate).C {
		stock++

		if err := simulateWrite(db, stock); err != nil {
			log.Printf("error simulating write: %v", err)
		}
	}

	return fmt.Errorf("finished simulateWrites unexectedly")
}

func simulateWrite(db *pgxpool.Pool, stock int) error {
	// Pick a product.
	productID := products[rand.Intn(len(products))]

	// Update stock in database.
	const stmt = `UPDATE stock SET quantity = $1 WHERE product_id = $2`
	if _, err := db.Exec(context.Background(), stmt, stock, productID); err != nil {
		return fmt.Errorf("updating database stock: %w", err)
	}

	writeQuantities.set(productID, stock)

	return nil
}

func readFromDB(db *pgxpool.Pool, productID string) (int, error) {
	const stmt = `SELECT quantity FROM stock WHERE product_id = $1`

	row := db.QueryRow(context.Background(), stmt, productID)

	var quantity int
	if err := row.Scan(&quantity); err != nil {
		return 0, fmt.Errorf("getting stock from database: %w", err)
	}

	return quantity, nil
}

func printLoop() {
	for range time.NewTicker(time.Second).C {
		writeQuantities.productsMu.RLock()
		readQuantities.productsMu.RLock()

		lines := []string{}

		for dbk, dbv := range writeQuantities.products {
			cv := readQuantities.products[dbk]

			if cv != dbv {
				lines = append(lines, fmt.Sprintf("%s (write: %d vs read: %d)\n", dbk, dbv, cv))
			}
		}

		writeQuantities.productsMu.RUnlock()
		readQuantities.productsMu.RUnlock()

		fmt.Println("\033[H\033[2J")
		if len(lines) > 0 {
			fmt.Println(strings.Join(lines, "\n"))
		} else {
			fmt.Println("write and read values match")
		}
	}
}
