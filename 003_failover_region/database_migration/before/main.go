package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.SetFlags(0)

	url, ok := os.LookupEnv("CONNECTION_STRING")
	if !ok {
		log.Fatalf("missing CONNECTION_STRING env var")
	}

	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatalf("error connecting to postgres: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error testing database connection: %v", err)
	}

	work(db)
}

func work(db *pgxpool.Pool) {
	for range time.NewTicker(time.Second).C {
		id := uuid.NewString()

		// Insert purchase.
		stmt := `INSERT INTO purchase (id, basket_id, member_id, amount) VALUES ($1, $2, $3, $4)`
		if _, err := db.Exec(context.Background(), stmt, id, uuid.NewString(), uuid.NewString(), rand.Float64()*100); err != nil {
			log.Printf("inserting purchase: %v", err)
			continue
		}

		// Select purchase.
		stmt = `SELECT amount FROM purchase WHERE id = $1`
		row := db.QueryRow(context.Background(), stmt, id)

		var value float64
		if err := row.Scan(&value); err != nil {
			log.Printf("selecting purchase: %v", err)
			continue
		}

		// Feedback.
		log.Println("ok")
	}
}
