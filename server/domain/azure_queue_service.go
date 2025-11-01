package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

var (
	azureQueueClient     *azqueue.ServiceClient
	azureQueueClientOnce sync.Once
	azureQueueClientErr  error
	azureQueueName       = "ocr-tasks-queue"
	azureQueueClientInstance *azqueue.QueueClient // ???????????????
	azureQueueClientMutex sync.Mutex
)

// azureQueueService Azure Queue Storage????
// S3?GCS???????Azure Queue Storage API???
type azureQueueService struct {
	serviceClient *azqueue.ServiceClient
	queueClient   *azqueue.QueueClient
	queueURL      string
	fallback      QueueService
}

// NewAzuriteQueueService Azurite Queue Storage???????
// S3?GCS???????Azure Queue Storage API???
func NewAzuriteQueueService() QueueService {
	return &azureQueueService{
		fallback: NewCommonQueueService(),
	}
}

// getAzureQueueServiceClient Azure Queue Storage??????????????????
func (q *azureQueueService) getAzureQueueServiceClient(ctx context.Context) (*azqueue.ServiceClient, error) {
	azureQueueClientOnce.Do(func() {
		// ???????????Azurite???? http://azurite:10000?
		azureEndpoint := os.Getenv("AZURE_STORAGE_ENDPOINT")
		if azureEndpoint == "" {
			azureEndpoint = "http://azurite:10000"
		}
		
		// Queue Storage?????????10001???
		// Blob Storage?10000????Queue?10001???
		queueEndpoint := ""
		if azureEndpoint != "" {
			// http://azurite:10000 -> http://azurite:10001
			// ??????10001???
			if strings.HasSuffix(azureEndpoint, ":10000") {
				queueEndpoint = strings.TrimSuffix(azureEndpoint, ":10000") + ":10001"
			} else {
				// ??????????????????
				queueEndpoint = azureEndpoint + ":10001"
			}
		}

		accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
		if accountName == "" {
			accountName = "devstoreaccount1"
		}

		accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
		if accountKey == "" {
			accountKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
		}

		if queueEndpoint == "" {
			azureQueueClientErr = fmt.Errorf("AZURE_STORAGE_ENDPOINT not set")
			return
		}

		// Azure Queue Storage????URL
		serviceURL := fmt.Sprintf("%s/%s", queueEndpoint, accountName)
		
		credential, err := azqueue.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			azureQueueClientErr = fmt.Errorf("failed to create Azure Queue credential: %w", err)
			return
		}

		serviceClient, err := azqueue.NewServiceClientWithSharedKeyCredential(serviceURL, credential, nil)
		if err != nil {
			azureQueueClientErr = fmt.Errorf("failed to create Azure Queue service client: %w", err)
			return
		}

		azureQueueClient = serviceClient
		azureQueueClientErr = nil
		log.Printf("Azure Queue Storage service client initialized: %s", serviceURL)
	})

	if azureQueueClientErr != nil {
		return nil, azureQueueClientErr
	}

	if azureQueueClient == nil {
		return nil, fmt.Errorf("Azure Queue Storage service client is nil")
	}

	return azureQueueClient, nil
}

// getQueueClient ??????????????????????????????
func (q *azureQueueService) getQueueClient(ctx context.Context) (*azqueue.QueueClient, error) {
	azureQueueClientMutex.Lock()
	defer azureQueueClientMutex.Unlock()
	
	// ??????????????????????????????
	if azureQueueClientInstance != nil {
		log.Printf("DEBUG: Using existing global queue client")
		return azureQueueClientInstance, nil
	}

	serviceClient, err := q.getAzureQueueServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	queueClient := serviceClient.NewQueueClient(azureQueueName)
	
	// ??????????????????????
	_, err = queueClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if err, ok := err.(*azcore.ResponseError); ok && err != nil {
			respErr = err
		}
		
		// ??????????????
		if respErr != nil && respErr.StatusCode == 404 {
			_, createErr := queueClient.Create(ctx, nil)
			if createErr != nil {
				log.Printf("Warning: Failed to create Azure Queue: %v", createErr)
				return nil, fmt.Errorf("failed to create Azure Queue: %w", createErr)
			}
			log.Printf("Azure Queue created: %s", azureQueueName)
		} else {
			// 404??????????????????????
			_, createErr := queueClient.Create(ctx, nil)
			if createErr != nil {
				log.Printf("Warning: Failed to check/create Azure Queue: %v, will retry", err)
			} else {
				log.Printf("Azure Queue created: %s", azureQueueName)
			}
		}
	} else {
		log.Printf("Azure Queue found: %s", azureQueueName)
	}

	log.Printf("DEBUG: Storing global queue client instance")
	azureQueueClientInstance = queueClient
	q.queueClient = queueClient
	return queueClient, nil
}

func (q *azureQueueService) EnqueueOCRTask(ctx context.Context, filename string, storageProvider string) error {
	log.Printf("DEBUG: Azure Queue EnqueueOCRTask called: file=%s, provider=%s", filename, storageProvider)
	
	queueClient, err := q.getQueueClient(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to get Azure Queue client for file=%s: %v, using fallback", filename, err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}
	
	log.Printf("DEBUG: Azure Queue client obtained successfully for file=%s", filename)

	// OCRTask?JSON???????
	task := &OCRTask{
		Filename:        filename,
		StorageProvider: storageProvider,
	}
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal OCR task: %w", err)
	}

	// Azure Queue Storage?????????
	log.Printf("DEBUG: Sending message to Azure Queue: content length=%d", len(taskJSON))
	response, err := queueClient.EnqueueMessage(ctx, string(taskJSON), nil)
	if err != nil {
		log.Printf("ERROR: Failed to enqueue message to Azure Queue: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	log.Printf("DEBUG: EnqueueMessage response received (response type: %T)", response)
	log.Printf("SUCCESS: OCR task enqueued to Azure Queue: file=%s, provider=%s", filename, storageProvider)
	return nil
}

func (q *azureQueueService) DequeueOCRTask(ctx context.Context) (*OCRTask, error) {
	log.Printf("DEBUG: Azure Queue DequeueOCRTask called - THIS LOG MUST APPEAR")
	
	queueClient, err := q.getQueueClient(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to get Azure Queue client for dequeue: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	log.Printf("DEBUG: Azure Queue client obtained successfully for dequeue - queueClient=%v", queueClient != nil)

	// Azure Queue Storage??????????
	// DequeueMessages?????1????????????
	log.Printf("DEBUG: Attempting to dequeue messages from Azure Queue: %s", azureQueueName)
	response, err := queueClient.DequeueMessages(ctx, &azqueue.DequeueMessagesOptions{
		NumberOfMessages:  to.Ptr(int32(1)),
		VisibilityTimeout: to.Ptr(int32(30)), // 30???????????
	})
	if err != nil {
		log.Printf("ERROR: Failed to dequeue message from Azure Queue: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	log.Printf("DEBUG: DequeueMessages response received, message count: %d", len(response.Messages))

	if len(response.Messages) == 0 {
		log.Printf("DEBUG: No messages in Azure Queue")
		return nil, nil
	}

	message := response.Messages[0]
	log.Printf("DEBUG: Message received from Azure Queue, MessageID: %v", message.MessageID)
	
	// ????????????
	messageText := ""
	if message.MessageText != nil {
		messageText = *message.MessageText
		log.Printf("DEBUG: Message text length: %d", len(messageText))
	}

	if messageText == "" {
		log.Printf("DEBUG: Message text is empty")
		return nil, nil
	}

	// JSON?OCRTask????????
	var task OCRTask
	if err := json.Unmarshal([]byte(messageText), &task); err != nil {
		log.Printf("ERROR: Failed to unmarshal OCR task from Azure Queue: %v", err)
		// ???????????????????????
		if message.MessageID != nil && message.PopReceipt != nil {
			queueClient.DeleteMessage(ctx, *message.MessageID, *message.PopReceipt, nil)
			log.Printf("DEBUG: Deleted invalid message from Azure Queue")
		}
		return nil, fmt.Errorf("failed to unmarshal OCR task: %w", err)
	}

	log.Printf("SUCCESS: OCR task dequeued from Azure Queue: file=%s, provider=%s", task.Filename, task.StorageProvider)
	
	// ??????VisibilityTimeout?????????????????????????
	// ???OCR??????????DeleteOCRTaskMessage???
	return &task, nil
}

// DeleteOCRTaskMessage ??????Azure?????????????
func (q *azureQueueService) DeleteOCRTaskMessage(ctx context.Context, messageID string, popReceipt string) error {
	queueClient, err := q.getQueueClient(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to get Azure Queue client for delete: %v", err)
		return err
	}
	
	if messageID == "" || popReceipt == "" {
		log.Printf("WARNING: Cannot delete message without MessageID and PopReceipt")
		return fmt.Errorf("messageID and popReceipt are required")
	}
	
	_, err = queueClient.DeleteMessage(ctx, messageID, popReceipt, nil)
	if err != nil {
		log.Printf("ERROR: Failed to delete message from Azure Queue: %v", err)
		return err
	}
	
	log.Printf("SUCCESS: Deleted message from Azure Queue: MessageID=%s", messageID)
	return nil
}
