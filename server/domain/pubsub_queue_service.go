package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	pubsubClient         *pubsub.Client
	pubsubClientOnce     sync.Once
	pubsubClientInitErr  error // ?????????
	pubsubTopicName      = "ocr-tasks"
	pubsubSubscriptionName = "ocr-tasks-subscription"
	pubsubProjectID      = "dev-project"
)

// pubsubQueueService ?GCP Pub/Sub????????????
type pubsubQueueService struct {
	client       *pubsub.Client
	topic        *pubsub.Topic
	sub          *pubsub.Subscription
	projectID    string
	topicName    string
	subName      string
	fallback     QueueService
	taskChan     chan *OCRTask // ??????????????
	receiveMutex sync.Mutex    // Receive???????
	receiveStarted bool        // Receive???????
	receiveCtx   context.Context    // Receive?????context
	receiveCancel context.CancelFunc // Receive?????cancel function
}

// NewGCPPubSubService ?GCP Pub/Sub???????????????
func NewGCPPubSubService() QueueService {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = pubsubProjectID
	}

	return &pubsubQueueService{
		projectID: projectID,
		topicName: pubsubTopicName,
		subName:  pubsubSubscriptionName,
		fallback: NewCommonQueueService(),
		taskChan: make(chan *OCRTask, 10), // ??????????
	}
}

// getPubsubClient ?Pub/Sub????????????????
func (q *pubsubQueueService) getPubsubClient(ctx context.Context) (*pubsub.Client, error) {
	pubsubClientOnce.Do(func() {
		emulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
		
		// PUBSUB_EMULATOR_HOST??????????SDK????????????????
		// ?????????SDK???????????
		if emulatorHost != "" {
			log.Printf("Using Pub/Sub emulator: %s (SDK will auto-configure)", emulatorHost)
		} else {
			log.Printf("PUBSUB_EMULATOR_HOST not set, using production Pub/Sub")
		}

		var err error
		// SDK?????PUBSUB_EMULATOR_HOST????????????????
		pubsubClient, err = pubsub.NewClient(ctx, q.projectID)
		if err != nil {
			pubsubClientInitErr = fmt.Errorf("failed to create Pub/Sub client: %w", err)
			pubsubClient = nil
			return
		}
		pubsubClientInitErr = nil
	})

	// ??????????????????????
	if pubsubClientInitErr != nil {
		return nil, pubsubClientInitErr
	}

	// ???????nil????????
	if pubsubClient == nil {
		return nil, fmt.Errorf("Pub/Sub client is nil (initialization failed)")
	}

	return pubsubClient, nil
}

// ensureTopicExists ??????????????????????????????
func (q *pubsubQueueService) ensureTopicExists(ctx context.Context) error {
	client, err := q.getPubsubClient(ctx)
	if err != nil {
		return err
	}

	// ???????nil???????????????
	if client == nil {
		return fmt.Errorf("Pub/Sub client is nil")
	}

	// recover?panic?????????????
	var topic *pubsub.Topic
	var topicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Warning: Panic in client.Topic(): %v", r)
				topicErr = fmt.Errorf("panic in client.Topic(): %v", r)
				topic = nil
			}
		}()
		topic = client.Topic(q.topicName)
	}()

	if topicErr != nil {
		return topicErr
	}
	if topic == nil {
		return fmt.Errorf("Failed to get Pub/Sub topic")
	}
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Printf("Warning: Failed to check topic existence: %v", err)
		return err
	}

	if !exists {
		topic, err = client.CreateTopic(ctx, q.topicName)
		if err != nil {
			log.Printf("Warning: Failed to create Pub/Sub topic: %v", err)
			return err
		}
		log.Printf("Pub/Sub topic created: %s", q.topicName)
	} else {
		log.Printf("Pub/Sub topic found: %s", q.topicName)
	}

	q.topic = topic
	return nil
}

// ensureSubscriptionExists ??????????????????????????????????
func (q *pubsubQueueService) ensureSubscriptionExists(ctx context.Context) error {
	if err := q.ensureTopicExists(ctx); err != nil {
		return err
	}

	client, err := q.getPubsubClient(ctx)
	if err != nil {
		return err
	}

	// ???????nil???????????????
	if client == nil {
		return fmt.Errorf("Pub/Sub client is nil")
	}

	sub := client.Subscription(q.subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		log.Printf("Warning: Failed to check subscription existence: %v", err)
		return err
	}

	if !exists {
		sub, err = client.CreateSubscription(ctx, q.subName, pubsub.SubscriptionConfig{
			Topic:       q.topic,
			AckDeadline: 30 * time.Second, // 30??time.Duration??????????
		})
		if err != nil {
			log.Printf("Warning: Failed to create Pub/Sub subscription: %v", err)
			return err
		}
		log.Printf("Pub/Sub subscription created: %s", q.subName)
		// ?????????????????????????????????????????????
		time.Sleep(500 * time.Millisecond)
	} else {
		log.Printf("Pub/Sub subscription found: %s", q.subName)
	}

	q.sub = sub
	return nil
}

func (q *pubsubQueueService) EnqueueOCRTask(ctx context.Context, filename string, storageProvider string) error {
	// ????????????????
	client, err := q.getPubsubClient(ctx)
	if err != nil || client == nil {
		log.Printf("Warning: Pub/Sub client unavailable: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	if err := q.ensureTopicExists(ctx); err != nil {
		log.Printf("Warning: Failed to ensure topic exists: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	// topic?nil????????
	if q.topic == nil {
		log.Printf("Warning: Pub/Sub topic is nil, using fallback")
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

	// Pub/Sub?????????
	result := q.topic.Publish(ctx, &pubsub.Message{
		Data: taskJSON,
	})

	// ????????
	messageID, err := result.Get(ctx)
	if err != nil {
		log.Printf("Warning: Failed to publish message to Pub/Sub: %v, using fallback", err)
		return q.fallback.EnqueueOCRTask(ctx, filename, storageProvider)
	}

	log.Printf("OCR task enqueued to Pub/Sub: file=%s, provider=%s, messageID=%s", filename, storageProvider, messageID)
	return nil
}

// startReceiveLoop Pub/Sub?Receive?????????????????
// ????: Pub/Sub?Receive????????????????????????????????1?????????
func (q *pubsubQueueService) startReceiveLoop(ctx context.Context) error {
	q.receiveMutex.Lock()
	defer q.receiveMutex.Unlock()

	if q.receiveStarted {
		log.Printf("Pub/Sub Receive loop already started")
		return nil // ??????
	}

	if q.sub == nil {
		return fmt.Errorf("subscription is nil")
	}

	// Receive?????context?????????
	q.receiveCtx, q.receiveCancel = context.WithCancel(context.Background())
	
	q.receiveStarted = true
	log.Printf("Starting Pub/Sub Receive loop for subscription: %s", q.subName)

	// ?????????Receive???????
		go func() {
		defer func() {
			q.receiveMutex.Lock()
			q.receiveStarted = false
			q.receiveMutex.Unlock()
			log.Printf("Pub/Sub Receive loop stopped")
		}()

		for {
			// Receive?????????????????????????????
			err := q.sub.Receive(q.receiveCtx, func(ctx context.Context, msg *pubsub.Message) {
				log.Printf("Debug: Received message from Pub/Sub: messageID=%s", msg.ID)
				// JSON?OCRTask?????????
				var t OCRTask
				if err := json.Unmarshal(msg.Data, &t); err != nil {
					log.Printf("Warning: Failed to unmarshal OCR task: %v", err)
					msg.Nack()
					return
				}

				// ???????????????????select????
				select {
				case q.taskChan <- &t:
					msg.Ack()
					log.Printf("OCR task enqueued to internal channel: file=%s, provider=%s", t.Filename, t.StorageProvider)
				case <-q.receiveCtx.Done():
					msg.Nack()
					log.Printf("Warning: Context canceled while enqueuing task")
				case <-time.After(10 * time.Second):
					msg.Nack()
					log.Printf("Warning: Timeout while enqueuing task to channel")
				}
			})

			if err != nil {
			if err == context.Canceled {
				log.Printf("Pub/Sub Receive loop canceled")
				return
			}
				log.Printf("Error in Pub/Sub Receive loop: %v, restarting...", err)
				// ?????????????????
				select {
				case <-q.receiveCtx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			} else {
				// ????????????????????
				log.Printf("Pub/Sub Receive loop stopped")
				q.receiveMutex.Lock()
				q.receiveStarted = false
				q.receiveMutex.Unlock()
				return
			}
		}
	}()

	// Receive???????????????????
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (q *pubsubQueueService) DequeueOCRTask(ctx context.Context) (*OCRTask, error) {
	log.Printf("DEBUG: DequeueOCRTask called for Pub/Sub (provider: gcs)")
	
	if err := q.ensureSubscriptionExists(ctx); err != nil {
		log.Printf("Warning: Failed to ensure subscription exists: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	// ??????????nil???????
	if q.sub == nil {
		log.Printf("Warning: Pub/Sub subscription is nil, using fallback")
		return q.fallback.DequeueOCRTask(ctx)
	}

	log.Printf("DEBUG: Subscription exists, calling startReceiveLoop")
	// Receive????????????
	if err := q.startReceiveLoop(ctx); err != nil {
		log.Printf("Warning: Failed to start Receive loop: %v, using fallback", err)
		return q.fallback.DequeueOCRTask(ctx)
	}

	log.Printf("DEBUG: startReceiveLoop succeeded, waiting for message from channel (timeout: 30s)")

	// ????????????????????????????????????30???????????
	select {
	case task := <-q.taskChan:
		log.Printf("OCR task dequeued from Pub/Sub: file=%s, provider=%s", task.Filename, task.StorageProvider)
		return task, nil
	case <-time.After(30 * time.Second):
		// ??????????????????
		log.Printf("DEBUG: No message received from Pub/Sub within 30 seconds")
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
