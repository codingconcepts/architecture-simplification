package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://user:password@localhost:5430/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging db: %v", err)
	}

	readWrite(db)
}

func readWrite(db *pgxpool.Pool) {
	for range time.NewTicker(time.Second).C {
		// Write.
		const writeStmt = `INSERT INTO product (name, price) VALUES ($1, $2)
											 RETURNING id`

		name := mustRandomString(10)
		price := round(rand.Float64()*100, 2)

		row := db.QueryRow(context.Background(), writeStmt, name, price)

		var id string
		if err := row.Scan(&id); err != nil {
			log.Printf("error inserting product: %v", err)
			continue
		}

		// Read.
		const readStmt = `SELECT name, price FROM product WHERE id = $1`
		row = db.QueryRow(context.Background(), readStmt, id)

		var dbName string
		var dbPrice float64
		if err := row.Scan(&dbName, &dbPrice); err != nil {
			log.Printf("error reading product: %v", err)
			continue
		}

		// Print
		fmt.Printf("inserted %s\n", id)
	}
}

func mustRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func round(val float64, precision int) float64 {
	return math.Round(val*(math.Pow10(precision))) / math.Pow10(precision)
}
