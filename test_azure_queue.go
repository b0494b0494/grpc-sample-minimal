package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"os"

	"grpc-sample-minimal/server/domain"
)

func main() {
	ctx := context.Background()

	// ?????????????
	if os.Getenv("AZURE_STORAGE_ENDPOINT") == "" {
		os.Setenv("AZURE_STORAGE_ENDPOINT", "http://localhost:10000")
	}
	if os.Getenv("AZURE_STORAGE_ACCOUNT_NAME") == "" {
		os.Setenv("AZURE_STORAGE_ACCOUNT_NAME", "devstoreaccount1")
	}
	if os.Getenv("AZURE_STORAGE_ACCOUNT_KEY") == "" {
		os.Setenv("AZURE_STORAGE_ACCOUNT_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")
	}

	// QueueManager???
	queueManager := domain.GetQueueManager()
	
	// QueueManager????
	queueManager.Enable()
	
	log.Println("=== Azure Queue Enqueue/Dequeue Test ===")
	
	// ???????????
	testFilename := fmt.Sprintf("test-file-azure-%d.txt", time.Now().Unix())
	
	// 1. ?????
	log.Printf("1. Enqueuing task: file=%s, provider=azure", testFilename)
	err := queueManager.EnqueueOCRTask(ctx, testFilename, "azure")
	if err != nil {
		log.Printf("ERROR: Failed to enqueue: %v", err)
		return
	}
	log.Printf("SUCCESS: Task enqueued successfully")
	
	// ??????????????????????????
	time.Sleep(2 * time.Second)
	
	// 2. ????
	log.Printf("2. Dequeuing task for provider: azure")
	
	// ?????????????????
	dequeueCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	
	task, err := queueManager.DequeueOCRTask(dequeueCtx, "azure")
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
	
	if task.StorageProvider != "azure" {
		log.Printf("ERROR: Provider mismatch! Expected: azure, Got: %s", task.StorageProvider)
		return
	}
	
	log.Printf("VERIFICATION: All checks passed for azure")
}
