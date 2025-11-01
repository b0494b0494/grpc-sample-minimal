package domain

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// QueueTaskStore ?????????????????????
type QueueTaskStore interface {
	// LogEnqueue ????????
	LogEnqueue(ctx context.Context, filename string, storageProvider string) error
	// LogDequeue ???????
	LogDequeue(ctx context.Context, filename string, storageProvider string) error
	// LogProcessing ???????
	LogProcessing(ctx context.Context, filename string, storageProvider string) error
	// LogCompleted ???????
	LogCompleted(ctx context.Context, filename string, storageProvider string) error
	// LogFailed ???????
	LogFailed(ctx context.Context, filename string, storageProvider string, err error) error
	// GetQueueStats ????????
	GetQueueStats(ctx context.Context, storageProvider string) (*QueueStats, error)
}

// QueueStats ???????
type QueueStats struct {
	Enqueued   int
	Dequeued   int
	Processing int
	Completed  int
	Failed     int
}

// sqliteQueueTaskStore SQLite?????????????
type sqliteQueueTaskStore struct {
	db *sql.DB
}

var (
	globalQueueTaskStore QueueTaskStore
	queueTaskStoreOnce   sync.Once
)

// NewQueueTaskStore ????????????
func NewQueueTaskStore(ctx context.Context) (QueueTaskStore, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/files.db"
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// ???????????????queue_tasks?????
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS queue_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			storage_provider TEXT NOT NULL,
			status TEXT NOT NULL,
			enqueued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			dequeued_at DATETIME,
			processed_at DATETIME,
			error_message TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_queue_filename_provider ON queue_tasks(filename, storage_provider);
		CREATE INDEX IF NOT EXISTS idx_queue_status ON queue_tasks(status);
		CREATE INDEX IF NOT EXISTS idx_queue_enqueued_at ON queue_tasks(enqueued_at);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue_tasks table: %w", err)
	}

	store := &sqliteQueueTaskStore{db: db}
	log.Printf("QueueTaskStore initialized successfully (db: %s)", dbPath)
	return store, nil
}

// GetOrCreateQueueTaskStore ??????????????????????
func GetOrCreateQueueTaskStore(ctx context.Context) (QueueTaskStore, error) {
	var err error
	queueTaskStoreOnce.Do(func() {
		globalQueueTaskStore, err = NewQueueTaskStore(ctx)
		if err != nil {
			log.Printf("Warning: Failed to create QueueTaskStore: %v", err)
			globalQueueTaskStore = nil
		}
	})

	if globalQueueTaskStore == nil {
		return nil, fmt.Errorf("QueueTaskStore is not available")
	}

	return globalQueueTaskStore, nil
}

func (s *sqliteQueueTaskStore) LogEnqueue(ctx context.Context, filename string, storageProvider string) error {
	query := `
		INSERT INTO queue_tasks (filename, storage_provider, status, enqueued_at)
		VALUES (?, ?, 'enqueued', CURRENT_TIMESTAMP)
	`
	_, err := s.db.ExecContext(ctx, query, filename, storageProvider)
	if err != nil {
		log.Printf("Warning: Failed to log enqueue: %v", err)
		return err
	}
	return nil
}

func (s *sqliteQueueTaskStore) LogDequeue(ctx context.Context, filename string, storageProvider string) error {
	query := `
		UPDATE queue_tasks
		SET status = 'dequeued', dequeued_at = CURRENT_TIMESTAMP
		WHERE filename = ? AND storage_provider = ? AND status = 'enqueued'
		ORDER BY enqueued_at ASC
		LIMIT 1
	`
	result, err := s.db.ExecContext(ctx, query, filename, storageProvider)
	if err != nil {
		log.Printf("Warning: Failed to log dequeue: %v", err)
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// ??????????????????
		query = `
			INSERT INTO queue_tasks (filename, storage_provider, status, enqueued_at, dequeued_at)
			VALUES (?, ?, 'dequeued', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = s.db.ExecContext(ctx, query, filename, storageProvider)
		return err
	}
	return nil
}

func (s *sqliteQueueTaskStore) LogProcessing(ctx context.Context, filename string, storageProvider string) error {
	query := `
		UPDATE queue_tasks
		SET status = 'processing'
		WHERE filename = ? AND storage_provider = ? AND status = 'dequeued'
		ORDER BY dequeued_at DESC
		LIMIT 1
	`
	_, err := s.db.ExecContext(ctx, query, filename, storageProvider)
	if err != nil {
		log.Printf("Warning: Failed to log processing: %v", err)
		return err
	}
	return nil
}

func (s *sqliteQueueTaskStore) LogCompleted(ctx context.Context, filename string, storageProvider string) error {
	query := `
		UPDATE queue_tasks
		SET status = 'completed', processed_at = CURRENT_TIMESTAMP
		WHERE filename = ? AND storage_provider = ? AND (status = 'processing' OR status = 'dequeued')
		ORDER BY dequeued_at DESC
		LIMIT 1
	`
	_, err := s.db.ExecContext(ctx, query, filename, storageProvider)
	if err != nil {
		log.Printf("Warning: Failed to log completed: %v", err)
		return err
	}
	return nil
}

func (s *sqliteQueueTaskStore) LogFailed(ctx context.Context, filename string, storageProvider string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	query := `
		UPDATE queue_tasks
		SET status = 'failed', processed_at = CURRENT_TIMESTAMP, error_message = ?
		WHERE filename = ? AND storage_provider = ? AND (status = 'processing' OR status = 'dequeued')
		ORDER BY dequeued_at DESC
		LIMIT 1
	`
	_, dbErr := s.db.ExecContext(ctx, query, errorMsg, filename, storageProvider)
	if dbErr != nil {
		log.Printf("Warning: Failed to log failed: %v", dbErr)
		return dbErr
	}
	return nil
}

func (s *sqliteQueueTaskStore) GetQueueStats(ctx context.Context, storageProvider string) (*QueueStats, error) {
	query := `
		SELECT 
			SUM(CASE WHEN status = 'enqueued' THEN 1 ELSE 0 END) as enqueued,
			SUM(CASE WHEN status = 'dequeued' THEN 1 ELSE 0 END) as dequeued,
			SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END) as processing,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM queue_tasks
		WHERE storage_provider = ?
	`
	
	var stats QueueStats
	err := s.db.QueryRowContext(ctx, query, storageProvider).Scan(
		&stats.Enqueued,
		&stats.Dequeued,
		&stats.Processing,
		&stats.Completed,
		&stats.Failed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}
	
	return &stats, nil
}

// GetNextTask ???????????????????????
// enqueued???????????????dequeued???
func (s *sqliteQueueTaskStore) GetNextTask(ctx context.Context, storageProvider string) (*OCRTask, error) {
	// ??????????enqueued???????????dequeued???
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ????enqueued?????????
	query := `
		SELECT filename, storage_provider, id
		FROM queue_tasks
		WHERE storage_provider = ? AND status = 'enqueued'
		ORDER BY enqueued_at ASC
		LIMIT 1
	`
	
	var filename, provider string
	var taskID int64
	err = tx.QueryRowContext(ctx, query, storageProvider).Scan(&filename, &provider, &taskID)
	if err == sql.ErrNoRows {
		return nil, nil // ?????
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next task: %w", err)
	}

	// ???dequeued???
	updateQuery := `
		UPDATE queue_tasks
		SET status = 'dequeued', dequeued_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, updateQuery, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &OCRTask{
		Filename:        filename,
		StorageProvider: provider,
	}, nil
}
