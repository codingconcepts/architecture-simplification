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

	}
}
