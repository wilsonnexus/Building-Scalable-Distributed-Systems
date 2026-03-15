package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Item struct {
	ItemID   int    `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	mode         = getenv("APP_MODE", "api") // api or processor
	awsRegion    = getenv("AWS_REGION", "us-east-1")
	snsTopicArn  = os.Getenv("SNS_TOPIC_ARN")
	sqsQueueURL  = os.Getenv("SQS_QUEUE_URL")
	port         = getenv("PORT", "8080")
	workerCount  = atoi(getenv("WORKERS", "1"))
	paymentSlots = make(chan struct{}, 1) // buffered channel bottleneck
	processorSem chan struct{}
	snsClient    *sns.Client
	sqsClient    *sqs.Client
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("failed loading AWS config: %v", err)
	}

	snsClient = sns.NewFromConfig(cfg)
	sqsClient = sqs.NewFromConfig(cfg)
	processorSem = make(chan struct{}, workerCount)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	if mode == "api" {
		log.Println("starting in API mode")
		mux.HandleFunc("/orders/sync", syncOrderHandler)
		mux.HandleFunc("/orders/async", asyncOrderHandler)
	} else {
		log.Printf("starting in PROCESSOR mode with %d workers\n", workerCount)
		go processorLoop()
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("listening on :%s\n", port)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func syncOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order.OrderID = newOrderID()
	order.Status = "processing"
	order.CreatedAt = time.Now()

	// Bottleneck simulation: only one payment can be verified at a time.
	paymentSlots <- struct{}{}
	time.Sleep(3 * time.Second)
	<-paymentSlots

	order.Status = "completed"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(order)
}

func asyncOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order.OrderID = newOrderID()
	order.Status = "pending"
	order.CreatedAt = time.Now()

	body, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "marshal failed", http.StatusInternalServerError)
		return
	}

	_, err = snsClient.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(snsTopicArn),
		Message:  aws.String(string(body)),
	})
	if err != nil {
		http.Error(w, "publish failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":  "order accepted",
		"order_id": order.OrderID,
		"status":   "pending",
	})
}

func processorLoop() {
	for {
		out, err := sqsClient.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsQueueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
			VisibilityTimeout:   30,
		})
		if err != nil {
			log.Printf("receive message error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(out.Messages) == 0 {
			continue
		}

		for _, msg := range out.Messages {
			msgCopy := msg
			processorSem <- struct{}{}
			go func(m sqstypes.Message) {
				defer func() { <-processorSem }()
				processMessage(m)
			}(msgCopy)
		}
	}
}

func processMessage(msg sqstypes.Message) {
	var envelope struct {
		Type      string `json:"Type"`
		Message   string `json:"Message"`
		MessageID string `json:"MessageId"`
	}

	rawBody := aws.ToString(msg.Body)
	bodyToUse := rawBody

	if err := json.Unmarshal([]byte(rawBody), &envelope); err == nil && envelope.Message != "" {
		bodyToUse = envelope.Message
	}

	var order Order
	if err := json.Unmarshal([]byte(bodyToUse), &order); err != nil {
		log.Printf("unmarshal order failed: %v, raw=%s", err, rawBody)
		return
	}

	log.Printf("processing order %s", order.OrderID)
	time.Sleep(3 * time.Second)
	log.Printf("completed order %s", order.OrderID)

	_, err := sqsClient.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(sqsQueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		log.Printf("delete message failed: %v", err)
	}
}

func newOrderID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
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
		return 1
	}
	return n
}
