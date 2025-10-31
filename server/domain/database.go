package domain

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	pb "grpc-sample-minimal/proto"
)

var (
	dbPath = os.Getenv("DB_PATH")
)

type FileMetadata struct {
	ID          int64
	Filename    string
	Namespace   string
	Size        int64
	StorageProvider string
	StoragePath string
	UploadedAt  time.Time
}

type FileMetadataRepository interface {
	Create(ctx context.Context, metadata *FileMetadata) error
	ListByProvider(ctx context.Context, provider string) ([]*pb.FileInfo, error)
	FindByFilename(ctx context.Context, filename string, provider string) (*FileMetadata, error)
	Delete(ctx context.Context, filename string, provider string) error
}

type sqliteFileMetadataRepository struct {
	db *sql.DB
}

func NewFileMetadataRepository(ctx context.Context) (FileMetadataRepository, error) {
	if dbPath == "" {
		dbPath = "/app/data/files.db"
	}
	
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &sqliteFileMetadataRepository{db: db}
	if err := repo.initSchema(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (r *sqliteFileMetadataRepository) initSchema(ctx context.Context) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL,
		namespace TEXT NOT NULL,
		size INTEGER NOT NULL,
		storage_provider TEXT NOT NULL,
		storage_path TEXT NOT NULL,
		uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(filename, storage_provider)
	);

	CREATE INDEX IF NOT EXISTS idx_storage_provider ON file_metadata(storage_provider);
	CREATE INDEX IF NOT EXISTS idx_namespace ON file_metadata(namespace);
	`

	_, err := r.db.ExecContext(ctx, createTableSQL)
	return err
}

func (r *sqliteFileMetadataRepository) Create(ctx context.Context, metadata *FileMetadata) error {
	query := `
		INSERT OR REPLACE INTO file_metadata (filename, namespace, size, storage_provider, storage_path, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	uploadedAt := metadata.UploadedAt
	if uploadedAt.IsZero() {
		uploadedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		metadata.Filename,
		metadata.Namespace,
		metadata.Size,
		metadata.StorageProvider,
		metadata.StoragePath,
		uploadedAt,
	)
	return err
}

func (r *sqliteFileMetadataRepository) ListByProvider(ctx context.Context, provider string) ([]*pb.FileInfo, error) {
	query := `
		SELECT filename, namespace, size
		FROM file_metadata
		WHERE storage_provider = ?
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

		var files []*pb.FileInfo
	for rows.Next() {
		var filename, namespace string
		var size int64

		if err := rows.Scan(&filename, &namespace, &size); err != nil {
			log.Printf("Error scanning file row: %v", err)
			continue
		}

		// Skip entries with empty filename (invalid data)
		if filename == "" {
			log.Printf("Skipping entry with empty filename")
			continue
		}

		log.Printf("DB row: filename=%s, namespace=%s, size=%d", filename, namespace, size)
		
		files = append(files, &pb.FileInfo{
			Filename:  filename,
			Namespace: namespace,
			Size:      size,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return files, nil
}

func (r *sqliteFileMetadataRepository) FindByFilename(ctx context.Context, filename string, provider string) (*FileMetadata, error) {
	query := `
		SELECT id, filename, namespace, size, storage_provider, storage_path, uploaded_at
		FROM file_metadata
		WHERE filename = ? AND storage_provider = ?
	`

	var metadata FileMetadata
	err := r.db.QueryRowContext(ctx, query, filename, provider).Scan(
		&metadata.ID,
		&metadata.Filename,
		&metadata.Namespace,
		&metadata.Size,
		&metadata.StorageProvider,
		&metadata.StoragePath,
		&metadata.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	return &metadata, nil
}

func (r *sqliteFileMetadataRepository) Delete(ctx context.Context, filename string, provider string) error {
	query := `DELETE FROM file_metadata WHERE filename = ? AND storage_provider = ?`
	_, err := r.db.ExecContext(ctx, query, filename, provider)
	return err
}

func (r *sqliteFileMetadataRepository) Close() error {
	return r.db.Close()
}
