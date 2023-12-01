package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	dbQuantities    *quantities
	cacheQuantities *quantities
)

func main() {
	readInterval := flag.Duration("r", time.Millisecond*100, "interval between reads")
	writeInterval := flag.Duration("w", time.Millisecond*1000, "interval between writes")
	flag.Parse()

	cache := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer cache.Close()

	if err := cache.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging db: %v", err)
	}

	dbQuantities = &quantities{
		products: map[string]int{},
	}
	cacheQuantities = &quantities{
		products: map[string]int{},
	}

	go simulateReads(db, cache, *readInterval)
	go simulateWrites(db, cache, *writeInterval)
	printLoop()
}

type quantities struct {
	productsMu sync.RWMutex
	products   map[string]int
}

type product struct {
	id    string
	stock int
}

func (q *quantities) set(product string, stock int) {
	q.productsMu.Lock()
	defer q.productsMu.Unlock()

	q.products[product] = stock
}

func simulateReads(db *pgxpool.Pool, cache *redis.Client, rate time.Duration) error {
	for range time.NewTicker(rate).C {
		if err := simulateRead(db, cache); err != nil {
			log.Printf("error simulating read: %v", err)
		}
	}

	return fmt.Errorf("finished simulateReads unexectedly")
}

func simulateRead(db *pgxpool.Pool, cache *redis.Client) error {
	// Pick a product.
	productID := products[rand.Intn(len(products))]

	// Read stock from cache.
	cmd := cache.Get(context.Background(), productID)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			// No stock in cache, read from DB.
			dbQuantity, err := readFromDB(db, productID)
			if err != nil {
				return fmt.Errorf("reading from db: %w", err)
			}

			// Set stock in cache.
			if err = cache.Set(context.Background(), productID, dbQuantity, 0).Err(); err != nil {
				return fmt.Errorf("setting cache stock: %w", err)
			}

			// Update new cache value.
			cacheQuantities.set(productID, dbQuantity)
		}
		return nil
	}

	// Update known cache value.
	val, err := cmd.Int()
	if err != nil {
		return fmt.Errorf("stock value int conversion: %w", err)
	}
	cacheQuantities.set(productID, val)

	return nil
}

func simulateWrites(db *pgxpool.Pool, cache *redis.Client, rate time.Duration) error {
	stock := 0
	for range time.NewTicker(rate).C {
		stock++

		if err := simulateWrite(db, cache, stock); err != nil {
			log.Printf("error simulating write: %v", err)
		}
	}

	return fmt.Errorf("finished simulateWrites unexectedly")
}

func simulateWrite(db *pgxpool.Pool, cache *redis.Client, stock int) error {
	// Pick a product.
	productID := products[rand.Intn(len(products))]

	// Update stock in database.
	const stmt = `UPDATE stock SET quantity = $1 WHERE product_id = $2`
	if _, err := db.Exec(context.Background(), stmt, stock, productID); err != nil {
		return fmt.Errorf("updating database stock: %w", err)
	}

	// Update known database value.
	dbQuantities.set(productID, stock)

	// Invalidate cache.
	if err := cache.Del(context.Background(), productID).Err(); err != nil {
		return fmt.Errorf("invalidating cache: %w", err)
	}

	return nil
}

func printLoop() {
	for range time.NewTicker(time.Second).C {
		dbQuantities.productsMu.RLock()
		cacheQuantities.productsMu.RLock()

		lines := []string{}

		for dbk, dbv := range dbQuantities.products {
			cv := cacheQuantities.products[dbk]

			if cv != dbv {
				lines = append(lines, fmt.Sprintf("%s (db: %d vs cache: %d)\n", dbk, dbv, cv))
			}
		}

		dbQuantities.productsMu.RUnlock()
		cacheQuantities.productsMu.RUnlock()

		fmt.Println("\033[H\033[2J")
		if len(lines) > 0 {
			fmt.Println(strings.Join(lines, "\n"))
		} else {
			fmt.Println("cache and database match")
		}
	}
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
