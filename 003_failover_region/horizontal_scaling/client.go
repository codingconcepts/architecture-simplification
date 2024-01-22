package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := flag.String("url", "", "database connection string")
	flag.Parse()

	if *url == "" {
		flag.Usage()
		os.Exit(2)
	}

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error testing database connection: %v", err)
	}

	work(db)
}

func work(db *pgxpool.Pool) {
	for range time.NewTicker(time.Second * 1).C {
		const stmt = `SELECT COUNT(*) FROM customer`

		start := time.Now()
		row := db.QueryRow(context.Background(), stmt)

		var count int
		if err := row.Scan(&count); err != nil {
			log.Printf("error getting row count: %v", err)
		}

		log.Printf("customers: %d (took %s)", count, time.Since(start))
	}
}
