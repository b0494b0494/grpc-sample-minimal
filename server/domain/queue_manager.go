package domain

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// QueueManager ?????????????????????????????
// ?????????????????????????????????????
type QueueManager struct {
	queues    map[string]QueueService // ??????????? -> ???????
	mutex     sync.RWMutex
	enabled   bool // ????????????????
}

var (
	globalQueueManager *QueueManager
	queueManagerOnce   sync.Once
)

// NewQueueManager ??????????????????
func NewQueueManager() *QueueManager {
	queueManagerOnce.Do(func() {
		globalQueueManager = &QueueManager{
			queues:  make(map[string]QueueService),
			enabled: true,
		}
	})
	return globalQueueManager
}

// GetQueueManager ????????????????????
func GetQueueManager() *QueueManager {
	if globalQueueManager == nil {
		return NewQueueManager()
	}
	return globalQueueManager
}

// GetOrCreateQueue ??????????????????????????????????
func (qm *QueueManager) GetOrCreateQueue(storageProvider string) (QueueService, error) {
	qm.mutex.RLock()
	if queue, exists := qm.queues[storageProvider]; exists {
		qm.mutex.RUnlock()
		return queue, nil
	}
	qm.mutex.RUnlock()

	// ??????????????
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// ??????????goroutine??????????????
	if queue, exists := qm.queues[storageProvider]; exists {
		return queue, nil
	}

	// ??????????
	queueService, err := NewQueueService(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue service for %s: %w", storageProvider, err)
	}

	qm.queues[storageProvider] = queueService
	log.Printf("Queue service created for storage provider: %s", storageProvider)
	return queueService, nil
}

// EnqueueOCRTask ?OCR?????????????????????????
func (qm *QueueManager) EnqueueOCRTask(ctx context.Context, filename string, storageProvider string) error {
	if !qm.enabled {
		return fmt.Errorf("queue manager is disabled")
	}

	queue, err := qm.GetOrCreateQueue(storageProvider)
	if err != nil {
		return fmt.Errorf("failed to get queue for %s: %w", storageProvider, err)
	}

	if err := queue.EnqueueOCRTask(ctx, filename, storageProvider); err != nil {
		return fmt.Errorf("failed to enqueue OCR task: %w", err)
	}

	log.Printf("OCR task enqueued via QueueManager: file=%s, provider=%s", filename, storageProvider)
	
	// ????????????
	if store, err := GetOrCreateQueueTaskStore(ctx); err == nil {
		if err := store.LogEnqueue(ctx, filename, storageProvider); err != nil {
			log.Printf("Warning: Failed to log enqueue to store: %v", err)
		}
	}
	
	return nil
}

// DequeueOCRTask ???????????????????OCR????????
func (qm *QueueManager) DequeueOCRTask(ctx context.Context, storageProvider string) (*OCRTask, error) {
	if !qm.enabled {
		return nil, fmt.Errorf("queue manager is disabled")
	}

	queue, err := qm.GetOrCreateQueue(storageProvider)
	if err != nil {
		log.Printf("ERROR: QueueManager failed to get queue for %s: %v", storageProvider, err)
		return nil, fmt.Errorf("failed to get queue for %s: %w", storageProvider, err)
	}

	log.Printf("DEBUG: QueueManager calling queue.DequeueOCRTask for provider=%s, queue type=%T", storageProvider, queue)
	task, err := queue.DequeueOCRTask(ctx)
	if err != nil {
		log.Printf("ERROR: QueueManager queue.DequeueOCRTask failed for provider=%s: %v", storageProvider, err)
		return nil, fmt.Errorf("failed to dequeue OCR task: %w", err)
	}

	if task != nil {
		log.Printf("OCR task dequeued via QueueManager: file=%s, provider=%s", task.Filename, task.StorageProvider)
		
		// ????????????
		if store, err := GetOrCreateQueueTaskStore(ctx); err == nil {
			if err := store.LogDequeue(ctx, task.Filename, task.StorageProvider); err != nil {
				log.Printf("Warning: Failed to log dequeue to store: %v", err)
			}
		}
	}
	return task, nil
}

// GetSupportedProviders ????????????????????????????
func (qm *QueueManager) GetSupportedProviders() []string {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	providers := make([]string, 0, len(qm.queues))
	for provider := range qm.queues {
		providers = append(providers, provider)
	}
	return providers
}

// Enable ????????????????
func (qm *QueueManager) Enable() {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	qm.enabled = true
	log.Println("QueueManager enabled")
}

// Disable ????????????????
func (qm *QueueManager) Disable() {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	qm.enabled = false
	log.Println("QueueManager disabled")
}

// IsEnabled ????????????????????
func (qm *QueueManager) IsEnabled() bool {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()
	return qm.enabled
}
