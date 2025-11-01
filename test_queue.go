package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"grpc-sample-minimal/server/domain"
)

func main() {
	ctx := context.Background()

	// QueueManager???
	queueManager := domain.GetQueueManager()
	
	// QueueManager????
	queueManager.Enable()
	
	log.Println("=== Queue Enqueue/Dequeue Test ===")
	
	// GCS (PubSub) ????
	testProvider(ctx, queueManager, "gcs")
	
	// S3 (SQS) ????
	testProvider(ctx, queueManager, "s3")
	
	// Azure ????
	testProvider(ctx, queueManager, "azure")
	
	log.Println("=== Test Completed ===")
}

func testProvider(ctx context.Context, queueManager *domain.QueueManager, provider string) {
	log.Printf("\n--- Testing %s provider ---", provider)
	
	// ????????
	testFilename := fmt.Sprintf("test-file-%s-%d.txt", provider, time.Now().Unix())
	
	// 1. ????????
	log.Printf("1. Enqueuing task: file=%s, provider=%s", testFilename, provider)
	err := queueManager.EnqueueOCRTask(ctx, testFilename, provider)
	if err != nil {
		log.Printf("ERROR: Failed to enqueue: %v", err)
		return
	}
	log.Printf("SUCCESS: Task enqueued successfully")
	
	// ???????PubSub????Receive?????????????????
	time.Sleep(2 * time.Second)
	
	// 2. ???????
	log.Printf("2. Dequeuing task for provider: %s", provider)
	
	// ??????????????
	dequeueCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	
	task, err := queueManager.DequeueOCRTask(dequeueCtx, provider)
	if err != nil {
		log.Printf("ERROR: Failed to dequeue: %v", err)
		return
	}
	
	if task == nil {
		log.Printf("WARNING: No task received (timeout or queue empty)")
		return
	}
	
	log.Printf("SUCCESS: Task dequeued successfully")
	log.Printf("  - Filename: %s", task.Filename)
	log.Printf("  - StorageProvider: %s", task.StorageProvider)
	
	if task.Filename != testFilename {
		log.Printf("ERROR: Filename mismatch! Expected: %s, Got: %s", testFilename, task.Filename)
		return
	}
	
	if task.StorageProvider != provider {
		log.Printf("ERROR: Provider mismatch! Expected: %s, Got: %s", provider, task.StorageProvider)
		return
	}
	
	log.Printf("VERIFICATION: All checks passed for %s", provider)
}
