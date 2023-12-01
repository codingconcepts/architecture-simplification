package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	if err = purgeOrders(db); err != nil {
		log.Fatalf("error generating orders: %v", err)
	}
}

func purgeOrders(db *pgxpool.Pool) error {
	// Delete orders 5 years and older.
	const stmt = `DELETE FROM orders WHERE ts <= now() - INTERVAL '43800h' LIMIT 10000`
	for range time.NewTicker(time.Minute).C {
		affected, err := db.Exec(context.Background(), stmt)
		if err != nil {
			return fmt.Errorf("purging orders: %w", err)
		}

		fmt.Printf("deleted %d rows\n", affected.RowsAffected())
	}

	return fmt.Errorf("finished purgeOrders unexpectedly")
}
