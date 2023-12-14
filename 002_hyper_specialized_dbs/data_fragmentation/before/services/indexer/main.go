package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.SetFlags(0)

	redisURL, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		log.Fatalf("missing REDIS_URL env var")
	}

	indexURL, ok := os.LookupEnv("INDEX_URL")
	if !ok {
		log.Fatalf("missing INDEX_URL env var")
	}

	time.Sleep(time.Second * 10)

	// Connect to database.
	db, err := pgxpool.New(context.Background(), indexURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	// Connect to Redis.
	cache := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})
	defer cache.Close()

	if err := cache.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	mustEnsureIndex(db)
	pipeToIndex(cache, db)
}

func mustEnsureIndex(db *pgxpool.Pool) {
	const stmt = `CREATE TABLE IF NOT EXISTS "product" (
									"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
									"name" VARCHAR(255) NOT NULL,
									"description" VARCHAR(255)
								)`

	if _, err := db.Exec(context.Background(), stmt); err != nil {
		log.Fatalf("error creating product table: %v", err)
	}
}

type product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func pipeToIndex(cache *redis.Client, db *pgxpool.Pool) {
	sub := cache.PSubscribe(context.Background(), "__keyevent@*__:set")
	defer sub.PUnsubscribe(context.Background(), "__keyevent@*__:set")

	for msg := range sub.Channel() {
		value := cache.Get(context.Background(), msg.Payload)
		raw, err := value.Bytes()
		if err != nil {
			log.Printf("error reading update: %v", err)
			continue
		}

		var p product
		if err = json.Unmarshal(raw, &p); err != nil {
			log.Printf("error parsing update: %v", err)
			continue
		}

		if err = writeToIndex(db, p); err != nil {
			log.Printf("error writing update: %v", err)
			continue
		}
	}
}

func writeToIndex(db *pgxpool.Pool, p product) error {
	const stmt = `UPSERT INTO product (id, name, description) VALUES ($1, $2, $3)`

	if _, err := db.Exec(context.Background(), stmt, p.ID, p.Name, p.Description); err != nil {
		return fmt.Errorf("inserting product into index: %w", err)
	}

	return nil
}
