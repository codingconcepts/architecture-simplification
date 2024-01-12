package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
)

func main() {
	log.SetFlags(0)

	postgres := mustConnectPostgres("postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	defer postgres.Close()

	cassandra := mustConnectCassandra("localhost:9042")
	defer cassandra.Close()

	bigquery := mustConnectBigQuery("http://localhost:9050")
	defer bigquery.Close()

	if err := simulateWriter(postgres, cassandra, bigquery); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}

type order struct {
	ID     string    `json:"id"`
	UserID string    `json:"user_id"`
	Total  float64   `json:"total"`
	TS     time.Time `json:"ts"`
}

func simulateWriter(pg *pgxpool.Pool, cs *gocql.Session, bq *bigquery.Client) error {
	i := 0
	for range time.NewTicker(time.Millisecond * 10).C {
		o := order{
			ID:     uuid.NewString(),
			UserID: uuid.NewString(),
			Total:  rand.Float64() * 100,
			TS:     time.Now(),
		}

		if err := writePostgres(pg, o); err != nil {
			log.Printf("writing to postgres: %v", err)
			continue
		}

		if err := writeCassandra(cs, o); err != nil {
			log.Printf("writing to cassandra: %v", err)
			continue
		}

		if err := writeBigQuery(bq, o); err != nil {
			log.Printf("writing to bigquery: %v", err)
			continue
		}

		i++
		fmt.Printf("saved %d orders\r", i)
	}

	return fmt.Errorf("finished work unexpectedly")
}

func writePostgres(pg *pgxpool.Pool, o order) error {
	const stmt = `INSERT INTO orders (id, user_id, total, ts) VALUES ($1, $2, $3, $4)`

	// Insert row into Postgres but "fail" 0.1% of the time.
	if rand.Intn(1000) == 42 {
		return fmt.Errorf("simulated error in postgres")
	}

	if _, err := pg.Exec(context.Background(), stmt, o.ID, o.UserID, o.Total, o.TS); err != nil {
		return fmt.Errorf("inserting row into postgres: %w", err)
	}

	return nil
}

func writeCassandra(cs *gocql.Session, o order) error {
	const stmt = `INSERT INTO example.orders (id, user_id, total, ts) VALUES (?, ?, ?, ?)`

	// Insert row into Cassandra but "fail" 0.1% of the time.
	if rand.Intn(1000) == 42 {
		return fmt.Errorf("simulated error in cassandra")
	}

	if err := cs.Query(stmt, o.ID, o.UserID, o.Total, o.TS).Exec(); err != nil {
		return fmt.Errorf("inserting row into cassandra: %w", err)
	}

	return nil
}

func (o order) Save() (map[string]bigquery.Value, string, error) {
	m := map[string]bigquery.Value{
		"id":      o.ID,
		"user_id": o.UserID,
		"total":   o.Total,
		"ts":      o.TS,
	}

	return m, o.ID, nil
}

func writeBigQuery(bq *bigquery.Client, o order) error {
	inserter := bq.Dataset("example").Table("orders").Inserter()

	// Insert row into BigQuery but "fail" 0.1% of the time.
	if rand.Intn(1000) == 42 {
		return fmt.Errorf("simulated error in bigquery")
	}

	if err := inserter.Put(context.Background(), o); err != nil {
		return fmt.Errorf("inserting row into bigquery: %w", err)
	}

	return nil
}

func mustConnectPostgres(url string) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error testing database connection: %v", err)
	}

	return db
}

func mustConnectCassandra(url string) *gocql.Session {
	cluster := gocql.NewCluster(url)

	// The connection to Cassandra fails once in a while, try in a back-off loop
	// until a connection attempt is successful but otherwise fail.
	for i := 1; i <= 10; i++ {
		cassandra, err := cluster.CreateSession()
		if err != nil {
			nextTry := time.Duration(100 * i)
			log.Printf("failed to connect to cassandra, trying again in %sms", nextTry)
			time.Sleep(nextTry)
			continue
		}

		return cassandra
	}

	log.Fatalf("unable to establish connection to cassandra")
	return nil
}

func mustConnectBigQuery(url string) *bigquery.Client {
	client, err := bigquery.NewClient(
		context.Background(),
		"local",
		option.WithEndpoint(url),
	)
	if err != nil {
		log.Fatalf("error connecting to big query: %v", err)
	}

	return client
}
