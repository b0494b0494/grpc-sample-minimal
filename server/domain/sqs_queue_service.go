package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	sqsQueueURL     string
	sqsQueueUROnce  sync.Once
	sqsClient       *sqs.Client
	sqsClientOnce   sync.Once
	sqsQueueName    = "ocr-tasks-queue"
)

// sqsQueueService ?AWS SQS????????????
type sqsQueueService struct {
	client   *sqs.Client
	queueURL string
	fallback QueueService
}

// NewLocalstackSQSService ?Localstack SQS???????????????
func NewLocalstackSQSService() QueueService {
	return &sqsQueueService{
		fallback: NewCommonQueueService(),
	}
}

// getSQSClient ?SQS????????????????
func (q *sqsQueueService) getSQSClient(ctx context.Context) (*sqs.Client, error) {
	var err error
	sqsClientOnce.Do(func() {
		localstackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
		if localstackEndpoint == "" {
			localstackEndpoint = "http://localstack:4566"
		}

		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}

		// ????: S3????????????Localstack???????????????
		accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		if accessKeyID == "" {
			accessKeyID = "test"
		}
		secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if secretAccessKey == "" {
			secretAccessKey = "test"
		}

		// S3??????????????: ????????????Localstack??????????
		cfg, loadErr := config.LoadDefaultConfig(ctx,
			config.WithRegion(awsRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: localstackEndpoint, SigningRegion: awsRegion}, nil
				},
			)),
		)
		if loadErr != nil {
			err = fmt.Errorf("failed to load AWS SDK config: %w", loadErr)
			return
		}

		// SQS??????????Localstack??HTTP????
		sqsOptions := func(o *sqs.Options) {
			// Localstack?HTTP??????????????????????
		}
		sqsClient = sqs.NewFromConfig(cfg, sqsOptions)
	})

	if err != nil {
		return nil, err
	}
	return sqsClient, nil
}

// getQueueURL ?SQS???URL??????????
func (q *sqsQueueService) getQueueURL(ctx context.Context) (string, error) {
	if q.queueURL != "" {
		return q.queueURL, nil
	}

	client, err := q.getSQSClient(ctx)
	if err != nil {
		return "", err
	}

	// ???????????
	queueURL := ""
	getURLOutput, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(sqsQueueName),
	})
	if err != nil {
		// ?????????
		var queueDoesNotExist *types.QueueDoesNotExist
		if !errors.As(err, &queueDoesNotExist) {
			log.Printf("Warning: Failed to get SQS queue URL: %v, trying to create queue", err)
		}

		// ??????
		createOutput, createErr := client.CreateQueue(ctx, &sqs.CreateQueueInput{
			QueueName: aws.String(sqsQueueName),
			Attributes: map[string]string{
				"VisibilityTimeout":     "30",
				"MessageRetentionPeriod": "345600", // 4?
			},
		})
		if createErr != nil {
			log.Printf("Warning: Failed to create SQS queue: %v, using fallback", createErr)
			return "", createErr
		}

		queueURL = aws.ToString(createOutput.QueueUrl)
		log.Printf("SQS queue created: %s", queueURL)
	} else {
		queueURL = aws.ToString(getURLOutput.QueueUrl)
		log.Printf("SQS queue found: %s", queueURL)
	}

	q.queueURL = queueURL
	return queueURL, nil
}

func (q *sqsQueueService) EnqueueOCRTask(ctx context.Context, filename string, storageProvider string) error {
	client, err := q.getSQSClient(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get SQS client: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	queueURL, err := q.getQueueURL(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get SQS queue URL: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	// OCRTask?JSON???????
	task := &OCRTask{
		Filename:        filename,
		StorageProvider: storageProvider,
	}
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal OCR task: %w", err)
	}

	// SQS?????????
	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(taskJSON)),
	})
	if err != nil {
		log.Printf("Warning: Failed to send message to SQS: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	log.Printf("OCR task enqueued to SQS: file=%s, provider=%s", filename, storageProvider)
	return nil
}

func (q *sqsQueueService) DequeueOCRTask(ctx context.Context) (*OCRTask, error) {
	client, err := q.getSQSClient(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get SQS client: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	queueURL, err := q.getQueueURL(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get SQS queue URL: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	// SQS??????????
	output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20, // ????????
		VisibilityTimeout:   30, // ??????????????????
	})
	if err != nil {
		log.Printf("Warning: Failed to receive message from SQS: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	if len(output.Messages) == 0 {
		return nil, nil
	}

	message := output.Messages[0]

	// JSON?OCRTask????????
	var task OCRTask
	if err := json.Unmarshal([]byte(aws.ToString(message.Body)), &task); err != nil {
		log.Printf("Warning: Failed to unmarshal OCR task: %v", err)
		// ???????????
		q.deleteMessage(ctx, client, queueURL, message.ReceiptHandle)
		return nil, fmt.Errorf("failed to unmarshal OCR task: %w", err)
	}

	// ???????????????????
	// ??: ????OCR??????????????????
	// ????????????????????????????
	// q.deleteMessage(ctx, client, queueURL, message.ReceiptHandle)

	log.Printf("OCR task dequeued from SQS: file=%s, provider=%s", task.Filename, task.StorageProvider)
	return &task, nil
}

// deleteMessage ?SQS????????????
func (q *sqsQueueService) deleteMessage(ctx context.Context, client *sqs.Client, queueURL string, receiptHandle *string) {
	if receiptHandle == nil {
		return
	}

	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("Warning: Failed to delete message from SQS: %v", err)
	}
}

// DeleteOCRTaskMessage ?OCR?????????????SQS????????????
// ???????????????????
func (q *sqsQueueService) DeleteOCRTaskMessage(ctx context.Context, receiptHandle *string) error {
	client, err := q.getSQSClient(ctx)
	if err != nil {
		return err
	}

	queueURL, err := q.getQueueURL(ctx)
	if err != nil {
		return err
	}

	q.deleteMessage(ctx, client, queueURL, receiptHandle)
	return nil
}
