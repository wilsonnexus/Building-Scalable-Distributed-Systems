package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var coldStart = true

func handler(ctx context.Context, event events.SNSEvent) error {
	if coldStart {
		log.Println("COLD START: first invocation in this execution environment")
		coldStart = false
	}

	for _, record := range event.Records {
		log.Printf("processing SNS message id=%s", record.SNS.MessageID)
		time.Sleep(3 * time.Second)
		log.Printf("completed SNS message id=%s", record.SNS.MessageID)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
