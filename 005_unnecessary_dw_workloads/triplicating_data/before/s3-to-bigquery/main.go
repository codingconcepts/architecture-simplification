package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handle)
}

func handle(ctx context.Context, e event) {
	for _, record := range e.Records {
		log.Printf("message id: %s", record.MessageID)

		var b body
		if err := json.Unmarshal([]byte(record.Body), &b); err != nil {
			log.Fatalf("error unmarshalling body: %v", err)
		}

		for _, s3 := range b.Records {
			fmt.Printf("s3:     %+v\n", s3.S3.Object.Key)
		}
	}
}

type event struct {
	Records []struct {
		Attributes struct {
			ApproximateFirstReceiveTimestamp string `json:"ApproximateFirstReceiveTimestamp"`
			ApproximateReceiveCount          string `json:"ApproximateReceiveCount"`
			SenderID                         string `json:"SenderId"`
			SentTimestamp                    string `json:"SentTimestamp"`
		} `json:"attributes"`
		AwsRegion         string `json:"awsRegion"`
		Body              string `json:"body"`
		EventSource       string `json:"eventSource"`
		EventSourceARN    string `json:"eventSourceARN"`
		Md5OfBody         string `json:"md5OfBody"`
		MessageAttributes struct {
		} `json:"messageAttributes"`
		MessageID     string `json:"messageId"`
		ReceiptHandle string `json:"receiptHandle"`
	} `json:"Records"`
}

type body struct {
	Records []struct {
		EventVersion string    `json:"eventVersion"`
		EventSource  string    `json:"eventSource"`
		AwsRegion    string    `json:"awsRegion"`
		EventTime    time.Time `json:"eventTime"`
		EventName    string    `json:"eventName"`
		UserIdentity struct {
			PrincipalID string `json:"principalId"`
		} `json:"userIdentity"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		ResponseElements struct {
			XAmzRequestID string `json:"x-amz-request-id"`
			XAmzID2       string `json:"x-amz-id-2"`
		} `json:"responseElements"`
		S3 struct {
			S3SchemaVersion string `json:"s3SchemaVersion"`
			ConfigurationID string `json:"configurationId"`
			Bucket          struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Sequencer string `json:"sequencer"`
				Size      int    `json:"size"`
				ETag      string `json:"eTag"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}
