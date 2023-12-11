package main

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://root@cockroachdb-public.crdb.svc.cluster.local:26257/defaultdb?sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to cockroachdb: %v", err)
	}
	defer db.Close()

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

		// Select database version.
		stmt = `SELECT version()`
		row = db.QueryRow(context.Background(), stmt)

		var version string
		if err := row.Scan(&version); err != nil {
			log.Printf("selecting version: %v", err)
			continue
		}

		// Feedback.
		log.Printf("ok (%s)", strings.Split(version, "(")[0])
	}

	log.Fatal("application unexpectedly exited")
}
