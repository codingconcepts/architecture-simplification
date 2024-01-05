package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx/v5/pgxpool"
	scyllacdc "github.com/scylladb/scylla-cdc-go"
)

func main() {
	scyllaURL, ok := os.LookupEnv("SCYLLA_URL")
	if !ok {
		log.Fatal("missing SCYLLA_URL env var")
	}

	indexURL, ok := os.LookupEnv("INDEX_URL")
	if !ok {
		log.Fatalf("missing INDEX_URL env var")
	}

	session, err := scyllaConnect(scyllaURL, 30)
	if err != nil {
		log.Fatalf("error connecting to scylla: %v", err)
	}
	defer session.Close()

	db, err := pgxpool.New(context.Background(), indexURL)
	if err != nil {
		log.Fatalf("error connecting to index: %v", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error testing index connection: %v", err)
	}

	cfg := &scyllacdc.ReaderConfig{
		Session:               session,
		TableNames:            []string{"store.product"},
		ChangeConsumerFactory: scyllacdc.MakeChangeConsumerFactoryFromFunc(createChangeConsumer(db)),
		Advanced: scyllacdc.AdvancedReaderConfig{
			ConfidenceWindowSize:   time.Second,
			PostNonEmptyQueryDelay: time.Second,
			PostEmptyQueryDelay:    time.Second,
			QueryTimeWindowSize:    time.Second,
			ChangeAgeLimit:         time.Second,
		},
	}

	reader, err := scyllacdc.NewReader(context.Background(), cfg)
	if err != nil {
		log.Fatalf("error creating cdc reader: %v", err)
	}

	if err := reader.Run(context.Background()); err != nil {
		log.Fatalf("error running cdc reader: %v", err)
	}
}

func createChangeConsumer(db *pgxpool.Pool) func(ctx context.Context, tableName string, c scyllacdc.Change) error {
	return func(ctx context.Context, tableName string, c scyllacdc.Change) error {
		for _, change := range c.PostImage {
			idRaw, ok := change.GetValue("id")
			if !ok {
				log.Printf("unable to get id from change row")
				continue
			}

			id := idRaw.(*gocql.UUID)
			name := change.GetAtomicChange("name")
			description := change.GetAtomicChange("description")

			err := updateIndex(
				db,
				*id,
				*name.Value.(*string),
				*description.Value.(*string),
			)

			if err != nil {
				log.Printf("error updating index: %v", err)
			}
		}

		return nil
	}
}

func updateIndex(db *pgxpool.Pool, id gocql.UUID, name, description string) error {
	const stmt = `INSERT INTO product (id, name, description) VALUES ($1, $2, $3)
								ON CONFLICT (id)
								DO UPDATE SET
									name = EXCLUDED.name,
									description = EXCLUDED.description`

	if _, err := db.Exec(context.Background(), stmt, id.String(), name, description); err != nil {
		return fmt.Errorf("upserting product: %w", err)
	}

	return nil
}

func scyllaConnect(url string, attempts int) (*gocql.Session, error) {
	for i := 0; i < attempts; i++ {
		cluster := gocql.NewCluster(url)
		session, err := cluster.CreateSession()
		if err != nil {
			log.Printf("error connecting to scylla: %v", err)
			time.Sleep(time.Second * time.Duration(i))
			continue
		}

		return session, nil
	}

	return nil, fmt.Errorf("exhausted connection attempts")
}
