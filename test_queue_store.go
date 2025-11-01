package main

import (
	"context"
	"fmt"
	"log"

	"grpc-sample-minimal/server/domain"
)

func main() {
	ctx := context.Background()

	// ????????????
	store, err := domain.GetOrCreateQueueTaskStore(ctx)
	if err != nil {
		log.Fatalf("Failed to get QueueTaskStore: %v", err)
	}

	// Azure??????
	stats, err := store.GetQueueStats(ctx, "azure")
	if err != nil {
		log.Fatalf("Failed to get queue stats: %v", err)
	}

	fmt.Printf("=== Azure Queue Statistics ===\n")
	fmt.Printf("Enqueued:   %d\n", stats.Enqueued)
	fmt.Printf("Dequeued:   %d\n", stats.Dequeued)
	fmt.Printf("Processing: %d\n", stats.Processing)
	fmt.Printf("Completed:  %d\n", stats.Completed)
	fmt.Printf("Failed:     %d\n", stats.Failed)
}
