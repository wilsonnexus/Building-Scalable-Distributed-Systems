package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"reliability-copilot/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Store struct {
	DDB       *dynamodb.Client
	SQS       *sqs.Client
	TableName string
	QueueURL  string
	AWSRegion string
}

type queueMessage struct {
	EventID string `json:"event_id"`
}

func New(tableName, queueURL, region string) (*Store, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &Store{
		DDB:       dynamodb.NewFromConfig(cfg),
		SQS:       sqs.NewFromConfig(cfg),
		TableName: tableName,
		QueueURL:  queueURL,
		AWSRegion: region,
	}, nil
}

func (s *Store) Enqueue(e model.FailureEvent) error {
	ctx := context.Background()

	eventItem := map[string]interface{}{
		"pk":               "EVENT#" + e.EventID,
		"sk":               "EVENT#" + e.EventID,
		"event_id":         e.EventID,
		"pipeline_id":      e.PipelineID,
		"log_text":         e.LogText,
		"source":           e.Source,
		"timestamp":        e.Timestamp,
		"status":           "queued",
		"failure_category": "",
		"recommendation":   "",
		"risk_score":       0.0,
		"processing_mode":  "",
		"error_message":    "",
		"created_at":       time.Now().UTC().Format(time.RFC3339),
	}

	av, err := attributevalue.MarshalMap(eventItem)
	if err != nil {
		return err
	}

	_, err = s.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item:      av,
	})
	if err != nil {
		return err
	}

	msgBody, _ := json.Marshal(queueMessage{EventID: e.EventID})

	_, err = s.SQS.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.QueueURL),
		MessageBody: aws.String(string(msgBody)),
	})
	return err
}

func (s *Store) GetEvent(eventID string) (*model.FailureEvent, error) {
	ctx := context.Background()

	out, err := s.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
			"sk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, errors.New("not found")
	}

	var raw map[string]interface{}
	if err := attributevalue.UnmarshalMap(out.Item, &raw); err != nil {
		return nil, err
	}

	return &model.FailureEvent{
		EventID:    raw["event_id"].(string),
		PipelineID: raw["pipeline_id"].(string),
		LogText:    raw["log_text"].(string),
		Source:     raw["source"].(string),
		Timestamp:  raw["timestamp"].(string),
	}, nil
}

func (s *Store) Complete(eventID, category, recommendation string, risk float64, mode string) error {
	ctx := context.Background()

	_, err := s.DDB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
			"sk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
		},
		UpdateExpression: aws.String("SET #status = :s, failure_category = :c, recommendation = :r, risk_score = :score, processing_mode = :m, error_message = :e"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":s":     &ddbtypes.AttributeValueMemberS{Value: "completed"},
			":c":     &ddbtypes.AttributeValueMemberS{Value: category},
			":r":     &ddbtypes.AttributeValueMemberS{Value: recommendation},
			":score": &ddbtypes.AttributeValueMemberN{Value: formatFloat(risk)},
			":m":     &ddbtypes.AttributeValueMemberS{Value: mode},
			":e":     &ddbtypes.AttributeValueMemberS{Value: ""},
		},
	})
	return err
}

func (s *Store) Fail(eventID, msg, mode string) error {
	ctx := context.Background()

	_, err := s.DDB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
			"sk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
		},
		UpdateExpression: aws.String("SET #status = :s, processing_mode = :m, error_message = :e"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":s": &ddbtypes.AttributeValueMemberS{Value: "failed"},
			":m": &ddbtypes.AttributeValueMemberS{Value: mode},
			":e": &ddbtypes.AttributeValueMemberS{Value: msg},
		},
	})
	return err
}

func (s *Store) GetResult(eventID string) (*model.AnalysisResult, error) {
	ctx := context.Background()

	out, err := s.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
			"sk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#" + eventID},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, errors.New("not found")
	}

	var raw map[string]interface{}
	if err := attributevalue.UnmarshalMap(out.Item, &raw); err != nil {
		return nil, err
	}

	risk := 0.0
	if v, ok := raw["risk_score"].(float64); ok {
		risk = v
	}

	return &model.AnalysisResult{
		EventID:         raw["event_id"].(string),
		FailureCategory: raw["failure_category"].(string),
		Recommendation:  raw["recommendation"].(string),
		RiskScore:       risk,
		Status:          raw["status"].(string),
		ProcessingMode:  raw["processing_mode"].(string),
		ErrorMessage:    raw["error_message"].(string),
	}, nil
}

func formatFloat(v float64) string {
	return json.Number(string([]byte(fmtFloat(v)))).String()
}

func fmtFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
