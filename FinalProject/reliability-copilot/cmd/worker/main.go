package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"reliability-copilot/internal/classifier"
	"reliability-copilot/internal/store"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type queueMessage struct {
	EventID string `json:"event_id"`
}

func main() {
	region := getenv("AWS_REGION", "us-east-1")
	tableName := getenv("DDB_TABLE_NAME", "reliability-copilot-events")
	queueURL := getenv("SQS_QUEUE_URL", "")

	st, err := store.New(tableName, queueURL, region)
	if err != nil {
		log.Fatal(err)
	}

	workers := atoi(getenv("WORKERS", "1"))
	inferenceDelayMs := atoi(getenv("INFERENCE_DELAY_MS", "0"))
	dbDelayMs := atoi(getenv("DB_DELAY_MS", "0"))
	failRatePct := atoi(getenv("FAIL_RATE_PCT", "0"))
	protectionEnabled := getenv("PROTECTION_ENABLED", "true") == "true"

	log.Printf("worker starting with WORKERS=%d INFERENCE_DELAY_MS=%d DB_DELAY_MS=%d FAIL_RATE_PCT=%d PROTECTION_ENABLED=%v",
		workers, inferenceDelayMs, dbDelayMs, failRatePct, protectionEnabled)

	for {
		out, err := st.SQS.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
			VisibilityTimeout:   30,
		})
		if err != nil {
			log.Printf("receive error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(out.Messages) == 0 {
			continue
		}

		for _, msg := range out.Messages {
			var qm queueMessage
			if err := json.Unmarshal([]byte(*msg.Body), &qm); err != nil {
				continue
			}

			event, err := st.GetEvent(qm.EventID)
			if err != nil {
				continue
			}

			mode := "rule-based-worker"

			if dbDelayMs > 0 {
				time.Sleep(time.Duration(dbDelayMs) * time.Millisecond)
			}

			if failRatePct > 0 && time.Now().UnixNano()%100 < int64(failRatePct) {
				if protectionEnabled {
					_ = st.Complete(event.EventID, "unknown_failure", "Fallback recommendation due to temporary worker failure.", 0.50, "fallback-protected")
				} else {
					_ = st.Fail(event.EventID, "simulated worker failure", "unprotected")
				}
			} else {
				if inferenceDelayMs > 0 {
					if protectionEnabled && inferenceDelayMs > 1500 {
						_ = st.Complete(event.EventID, "unknown_failure", "Fallback recommendation due to slow inference path.", 0.52, "fallback-timeout")
					} else {
						time.Sleep(time.Duration(inferenceDelayMs) * time.Millisecond)
						category, recommendation, score := classifier.Classify(event.LogText)
						_ = st.Complete(event.EventID, category, recommendation, score, mode)
					}
				} else {
					category, recommendation, score := classifier.Classify(event.LogText)
					_ = st.Complete(event.EventID, category, recommendation, score, mode)
				}
			}

			_, _ = st.SQS.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
		}
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
