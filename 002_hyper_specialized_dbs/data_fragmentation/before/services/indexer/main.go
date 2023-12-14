package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.SetFlags(0)

	//time.Sleep(time.Second * 20)

	// // Connect to Postgres.
	// db, err := pgxpool.New(context.Background(), "postgres://postgres:password@index:5432/postgres?sslmode=disable")
	// if err != nil {
	// 	log.Fatalf("error connecting to database: %v", err)
	// }
	// defer db.Close()

	// if err = db.Ping(context.Background()); err != nil {
	// 	log.Fatalf("error pinging database: %v", err)
	// }

	// Connect to Redis.
	cache := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer cache.Close()

	if err := cache.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	pipeToIndex(cache, nil)
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

		log.Printf("value: %s", string(raw))
	}
}
