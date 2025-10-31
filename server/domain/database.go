package domain

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// OCRResultRepository ?OCR?????????????????????
type OCRResultRepository interface {
	SaveOCRResult(ctx context.Context, result *OCRResult) error
	GetOCRResult(ctx context.Context, filename string, provider string, engineName string) (*OCRResult, error)
	ListOCRResults(ctx context.Context, provider string) ([]*OCRResult, error)
	GetOCRComparison(ctx context.Context, filename string, provider string) ([]*OCRResult, error)
	DeleteOCRResult(ctx context.Context, filename string, provider string, engineName string) error
	// LogError ??????????????????????????????
	LogError(ctx context.Context, filename string, provider string, engineName string, errorType string, errorMsg string) error
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
	
	-- OCR??????
	CREATE TABLE IF NOT EXISTS ocr_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL,
		storage_provider TEXT NOT NULL,
		engine_name TEXT NOT NULL,  -- 'tesseract', 'easyocr', 'paddleocr'
		status TEXT NOT NULL,  -- 'processing', 'completed', 'failed'
		extracted_text TEXT,
		error_message TEXT,
		average_confidence REAL,  -- ?????
		processed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(filename, storage_provider, engine_name)  -- ???????????????????
	);

	CREATE INDEX IF NOT EXISTS idx_ocr_filename_provider ON ocr_results(filename, storage_provider);
	CREATE INDEX IF NOT EXISTS idx_ocr_engine ON ocr_results(engine_name);
	CREATE INDEX IF NOT EXISTS idx_ocr_status ON ocr_results(status);
	CREATE INDEX IF NOT EXISTS idx_ocr_failed ON ocr_results(status, processed_at) WHERE status = 'failed';
	
	-- OCR????????????????????????????????
	CREATE TABLE IF NOT EXISTS ocr_error_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL,
		storage_provider TEXT NOT NULL,
		engine_name TEXT NOT NULL,
		error_type TEXT,  -- 'storage_error', 'ocr_error', 'db_error', 'panic', 'save_error'
		error_message TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_ocr_error_filename ON ocr_error_logs(filename, storage_provider);
	CREATE INDEX IF NOT EXISTS idx_ocr_error_type ON ocr_error_logs(error_type);
	CREATE INDEX IF NOT EXISTS idx_ocr_error_created ON ocr_error_logs(created_at);
	
	-- OCR???????
	CREATE TABLE IF NOT EXISTS ocr_pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ocr_result_id INTEGER NOT NULL,
		page_number INTEGER NOT NULL,
		text TEXT,
		confidence REAL,
		FOREIGN KEY (ocr_result_id) REFERENCES ocr_results(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_ocr_pages_result_id ON ocr_pages(ocr_result_id);
	
	-- ??OCR????????????????
	CREATE VIEW IF NOT EXISTS ocr_comparison AS
	SELECT 
		filename,
		storage_provider,
		GROUP_CONCAT(engine_name) as engines,
		COUNT(*) as engine_count,
		MAX(processed_at) as last_processed_at
	FROM ocr_results
	WHERE status = 'completed'
	GROUP BY filename, storage_provider;
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
		SELECT filename, namespace, size, uploaded_at
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
		var uploadedAt time.Time

		if err := rows.Scan(&filename, &namespace, &size, &uploadedAt); err != nil {
			log.Printf("Error scanning file row: %v", err)
			continue
		}

		// Skip entries with empty filename (invalid data)
		if filename == "" {
			log.Printf("Skipping entry with empty filename")
			continue
		}

		log.Printf("DB row: filename=%s, namespace=%s, size=%d, uploaded_at=%v", filename, namespace, size, uploadedAt)
		
		var uploadedAtUnix int64
		if !uploadedAt.IsZero() {
			uploadedAtUnix = uploadedAt.Unix()
		}
		
		files = append(files, &pb.FileInfo{
			Filename:   filename,
			Namespace:  namespace,
			Size:       size,
			UploadedAt: uploadedAtUnix,
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

// sqliteOCRResultRepository ?OCRResultRepository?SQLite??
type sqliteOCRResultRepository struct {
	db *sql.DB
}

// NewOCRResultRepository ????OCRResultRepository?????
func NewOCRResultRepository(ctx context.Context) (OCRResultRepository, error) {
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

	repo := &sqliteOCRResultRepository{db: db}
	
	// ???????initSchema???????????????????
	// ?????????????????
	
	return repo, nil
}

// SaveOCRResult ?OCR???????
func (r *sqliteOCRResultRepository) SaveOCRResult(ctx context.Context, result *OCRResult) error {
	// ??????????
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// OCR?????
	query := `
		INSERT OR REPLACE INTO ocr_results 
		(filename, storage_provider, engine_name, status, extracted_text, error_message, average_confidence, processed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	processedAt := result.ProcessedAt
	if processedAt.IsZero() {
		processedAt = time.Now()
	}
	
	errorMsg := ""
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}
	
	// ?????????????extracted_text??????
	extractedText := result.ExtractedText
	if extractedText == "" && len(result.Pages) > 0 {
		for _, page := range result.Pages {
			extractedText += page.Text + "\n"
		}
		extractedText = strings.TrimSpace(extractedText)
	}
	
	// ????????
	avgConfidence := result.Confidence
	if avgConfidence == 0 && len(result.Pages) > 0 {
		sum := 0.0
		for _, page := range result.Pages {
			sum += page.Confidence
		}
		avgConfidence = sum / float64(len(result.Pages))
	}
	
	res, err := tx.ExecContext(ctx, query,
		result.Filename,
		result.StorageProvider, // storage_provider
		result.EngineName,
		result.Status,
		extractedText,
		errorMsg,
		avgConfidence,
		processedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save OCR result: %w", err)
	}
	
	ocrResultID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get OCR result ID: %w", err)
	}
	
	// ?????????
	_, err = tx.ExecContext(ctx, "DELETE FROM ocr_pages WHERE ocr_result_id = ?", ocrResultID)
	if err != nil {
		return fmt.Errorf("failed to delete existing pages: %w", err)
	}
	
	// OCR??????
	if len(result.Pages) > 0 {
		pageQuery := `
			INSERT INTO ocr_pages (ocr_result_id, page_number, text, confidence)
			VALUES (?, ?, ?, ?)
		`
		for _, page := range result.Pages {
			_, err = tx.ExecContext(ctx, pageQuery,
				ocrResultID,
				page.PageNumber,
				page.Text,
				page.Confidence,
			)
			if err != nil {
				return fmt.Errorf("failed to save OCR page: %w", err)
			}
		}
	}
	
	// ????????????
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetOCRResult ?OCR???????
func (r *sqliteOCRResultRepository) GetOCRResult(ctx context.Context, filename string, provider string, engineName string) (*OCRResult, error) {
	query := `
		SELECT id, filename, storage_provider, engine_name, status, extracted_text, error_message, average_confidence, processed_at
		FROM ocr_results
		WHERE filename = ? AND storage_provider = ? AND engine_name = ?
	`
	
	var result OCRResult
	var resultID int64
	var errorMsg sql.NullString
	var processedAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, filename, provider, engineName).Scan(
		&resultID,
		&result.Filename,
		&result.StorageProvider, // storage_provider
		&result.EngineName,
		&result.Status,
		&result.ExtractedText,
		&errorMsg,
		&result.Confidence,
		&processedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OCR result: %w", err)
	}
	
	if errorMsg.Valid {
		result.Error = fmt.Errorf(errorMsg.String)
	}
	if processedAt.Valid {
		result.ProcessedAt = processedAt.Time
	}
	
	// ????????
	pagesQuery := `
		SELECT page_number, text, confidence
		FROM ocr_pages
		WHERE ocr_result_id = ?
		ORDER BY page_number
	`
	
	rows, err := r.db.QueryContext(ctx, pagesQuery, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OCR pages: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var page OCRPage
		if err := rows.Scan(&page.PageNumber, &page.Text, &page.Confidence); err != nil {
			log.Printf("Error scanning OCR page row: %v", err)
			continue
		}
		result.Pages = append(result.Pages, page)
	}
	
	return &result, nil
}

// ListOCRResults ??????????OCR?????????
func (r *sqliteOCRResultRepository) ListOCRResults(ctx context.Context, provider string) ([]*OCRResult, error) {
	query := `
		SELECT id, filename, storage_provider, engine_name, status, extracted_text, error_message, average_confidence, processed_at
		FROM ocr_results
		WHERE storage_provider = ?
		ORDER BY processed_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query OCR results: %w", err)
	}
	defer rows.Close()
	
	var results []*OCRResult
	for rows.Next() {
		var result OCRResult
		var resultID int64
		var errorMsg sql.NullString
		var processedAt sql.NullTime
		
		if err := rows.Scan(
			&resultID,
			&result.Filename,
			&result.StorageProvider,
			&result.EngineName,
			&result.Status,
			&result.ExtractedText,
			&errorMsg,
			&result.Confidence,
			&processedAt,
		); err != nil {
			log.Printf("Error scanning OCR result row: %v", err)
			continue
		}
		
		if errorMsg.Valid {
			result.Error = fmt.Errorf(errorMsg.String)
		}
		if processedAt.Valid {
			result.ProcessedAt = processedAt.Time
		}
		
		results = append(results, &result)
	}
	
	return results, nil
}

// GetOCRComparison ???OCR????????????
func (r *sqliteOCRResultRepository) GetOCRComparison(ctx context.Context, filename string, provider string) ([]*OCRResult, error) {
	query := `
		SELECT id, filename, storage_provider, engine_name, status, extracted_text, error_message, average_confidence, processed_at
		FROM ocr_results
		WHERE filename = ? AND storage_provider = ?
		ORDER BY engine_name
	`
	
	rows, err := r.db.QueryContext(ctx, query, filename, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query OCR comparison: %w", err)
	}
	defer rows.Close()
	
	var results []*OCRResult
	for rows.Next() {
		var result OCRResult
		var resultID int64
		var errorMsg sql.NullString
		var processedAt sql.NullTime
		
		if err := rows.Scan(
			&resultID,
			&result.Filename,
			&result.StorageProvider,
			&result.EngineName,
			&result.Status,
			&result.ExtractedText,
			&errorMsg,
			&result.Confidence,
			&processedAt,
		); err != nil {
			log.Printf("Error scanning OCR result row: %v", err)
			continue
		}
		
		if errorMsg.Valid {
			result.Error = fmt.Errorf(errorMsg.String)
		}
		if processedAt.Valid {
			result.ProcessedAt = processedAt.Time
		}
		
		// ?????????
		pagesQuery := `
			SELECT page_number, text, confidence
			FROM ocr_pages
			WHERE ocr_result_id = ?
			ORDER BY page_number
		`
		
		pageRows, err := r.db.QueryContext(ctx, pagesQuery, resultID)
		if err == nil {
			defer pageRows.Close()
			for pageRows.Next() {
				var page OCRPage
				if err := pageRows.Scan(&page.PageNumber, &page.Text, &page.Confidence); err == nil {
					result.Pages = append(result.Pages, page)
				}
			}
		}
		
		results = append(results, &result)
	}
	
	return results, nil
}

// DeleteOCRResult ?OCR???????
func (r *sqliteOCRResultRepository) DeleteOCRResult(ctx context.Context, filename string, provider string, engineName string) error {
	query := `DELETE FROM ocr_results WHERE filename = ? AND storage_provider = ? AND engine_name = ?`
	_, err := r.db.ExecContext(ctx, query, filename, provider, engineName)
	return err
}

// LogError ???????????????????
func (r *sqliteOCRResultRepository) LogError(ctx context.Context, filename string, provider string, engineName string, errorType string, errorMsg string) error {
	query := `
		INSERT INTO ocr_error_logs (filename, storage_provider, engine_name, error_type, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := r.db.ExecContext(ctx, query, filename, provider, engineName, errorType, errorMsg)
	return err
}

// Close ?????????????
func (r *sqliteOCRResultRepository) Close() error {
	return r.db.Close()
}
