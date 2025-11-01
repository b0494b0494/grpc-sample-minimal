package domain

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// MultiOCRClient ???OCR??????????????????
type MultiOCRClient struct {
	clients map[string]OCRClient // engineName -> OCRClient
	mu      sync.RWMutex
}

// NewMultiOCRClient ???OCR?????????
// ??????????????????:
//   - OCR_TESSERACT_ENDPOINT
//   - OCR_EASYOCR_ENDPOINT
//   - OCR_PADDLEOCR_ENDPOINT (????)
func NewMultiOCRClient(authToken string) (*MultiOCRClient, error) {
	multi := &MultiOCRClient{
		clients: make(map[string]OCRClient),
	}

	// Tesseract???????
	if endpoint := os.Getenv("OCR_TESSERACT_ENDPOINT"); endpoint != "" {
		client, err := NewOCRClient(endpoint, authToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create Tesseract OCR client: %w", err)
		}
		multi.clients["tesseract"] = client
	}

	// EasyOCR???????
	if endpoint := os.Getenv("OCR_EASYOCR_ENDPOINT"); endpoint != "" {
		client, err := NewOCRClient(endpoint, authToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create EasyOCR client: %w", err)
		}
		multi.clients["easyocr"] = client
	}

	return multi, nil
}

// ProcessDocument ????OCR?????????????
func (m *MultiOCRClient) ProcessDocument(ctx context.Context, filename string, content io.Reader, storageProvider string) (map[string]*OCRResult, error) {
	m.mu.RLock()
	clients := make(map[string]OCRClient)
	for name, client := range m.clients {
		clients[name] = client
	}
	m.mu.RUnlock()

	if len(clients) == 0 {
		return nil, fmt.Errorf("no OCR clients available")
	}

	// ?????????????????????????????????
	// ????????????????
	results := make(map[string]*OCRResult)
	for name := range clients {
		results[name] = &OCRResult{
			Filename:        filename,
			StorageProvider: storageProvider,
			EngineName:      name,
			Status:          "queued",
		}
	}

	return results, nil
}

// GetClient ??????????OCR?????????
func (m *MultiOCRClient) GetClient(engineName string) OCRClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[engineName]
}

// GetAvailableEngines ?????????????????
func (m *MultiOCRClient) GetAvailableEngines() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	engines := make([]string, 0, len(m.clients))
	for name := range m.clients {
		engines = append(engines, name)
	}
	return engines
}

// Close ????OCR??????????
func (m *MultiOCRClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []string
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing OCR clients: %s", strings.Join(errs, "; "))
	}

	return nil
}