package main

import (
	"log"

	"github.com/gocql/gocql"
)

func main() {
	cassandra := mustConnectCassandra()
	defer cassandra.Close()
}

func mustConnectCassandra() *gocql.Session {
	cluster := gocql.NewCluster("localhost:9042")

	cassandra, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("error connecting to cassandra: %v", err)
	}

	return cassandra
}

func mustConnectRedshift() {
	// // Initialize a session
	// sess, _ := session.NewSession(&aws.Config{
	// 	Region:           aws.String("us-east-1"),
	// 	Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	// 	S3ForcePathStyle: aws.Bool(true),
	// 	Endpoint:         aws.String("http://localhost:4566"),
	// })

	// // Create S3 service client
	// client := s3.New(sess)
}
