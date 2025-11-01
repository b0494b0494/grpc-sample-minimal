package domain

import (
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// easyOCREngine EasyOCR???????
type easyOCREngine struct {
	languages []string // OCR??: ["ja", "en"]??
}

// NewEasyOCREngine EasyOCR????????????
func NewEasyOCREngine(languages []string) OCREngine {
	if len(languages) == 0 {
		languages = []string{"ja", "en"} // ?????: ???+??
	}
	return &easyOCREngine{
		languages: languages,
	}
}

// Name ????????
func (e *easyOCREngine) Name() string {
	return "easyocr"
}

// ProcessDocument ?????????OCR??
func (e *easyOCREngine) ProcessDocument(ctx context.Context, filename string, content io.Reader) (*OCRResult, error) {
	result := &OCRResult{
		Filename:    filename,
		EngineName:  e.Name(),
		Status:      "processing",
		ProcessedAt: time.Now(),
	}

	// ?????
	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	// PDF?????poppler-utils???
	if ext == "pdf" {
		return e.processPDF(ctx, filename, content, result)
	}

	// ??????
	if isImageFile(ext) {
		return e.processImageFile(ctx, filename, content, result)
	}

	// ?????????????????
	result.Status = "failed"
	result.Error = fmt.Errorf("unsupported file type: %s", ext)
	return result, result.Error
}

// ProcessImage ???OCR??
func (e *easyOCREngine) ProcessImage(ctx context.Context, img image.Image) (string, float64, error) {
	// ????????????
	tempFile, err := saveImageToTempFile(img)
	if err != nil {
		return "", 0.0, fmt.Errorf("failed to save image to temp file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// EasyOCR?Python????????????
	text, confidence, err := e.callEasyOCR(tempFile.Name())
	if err != nil {
		return "", 0.0, fmt.Errorf("EasyOCR processing failed: %w", err)
	}

	return text, confidence, nil
}

// callEasyOCR Python????????EasyOCR?????
// ????: HTTP/gRPC API??????????????????????
func (e *easyOCREngine) callEasyOCR(imagePath string) (string, float64, error) {
	// TODO: Python?????????????????????????
	// ??????????????????????
	// ?????Python????????????HTTP/gRPC API?????????
	
	// ????: ???????????
	return "", 0.0, fmt.Errorf("EasyOCR implementation pending: Python integration required")
}

// processPDF PDF??
func (e *easyOCREngine) processPDF(ctx context.Context, filename string, content io.Reader, result *OCRResult) (*OCRResult, error) {
	// PDF??
	pdfConverter := NewPDFConverter()
	images, err := pdfConverter.ConvertPDFToImages(ctx, content)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to convert PDF to images: %w", err)
		return result, result.Error
	}

	log.Printf("PDF converted to %d images", len(images))

	// ????OCR??
	var allText strings.Builder
	var totalConfidence float64
	pages := make([]OCRPage, 0, len(images))

	for pageNum, img := range images {
		// OCR??
		text, confidence, err := e.ProcessImage(ctx, img)
		if err != nil {
			log.Printf("Failed to process OCR for page %d: %v", pageNum+1, err)
			// ?????????
			pages = append(pages, OCRPage{
				PageNumber: pageNum + 1,
				Text:       "",
				Confidence: 0.0,
			})
			continue
		}

		// ?????
		allText.WriteString(fmt.Sprintf("\n--- Page %d ---\n", pageNum+1))
		allText.WriteString(text)
		totalConfidence += confidence
		pages = append(pages, OCRPage{
			PageNumber: pageNum + 1,
			Text:       text,
			Confidence: confidence,
		})

		log.Printf("Processed page %d/%d: confidence=%.2f", pageNum+1, len(images), confidence)
	}

	// ?????
	if len(pages) == 0 {
		result.Status = "failed"
		result.Error = fmt.Errorf("no pages were successfully processed")
		return result, result.Error
	}

	result.ExtractedText = strings.TrimSpace(allText.String())
	if len(pages) > 0 {
		result.Confidence = totalConfidence / float64(len(pages))
	}
	result.Status = "completed"
	result.Pages = pages

	log.Printf("PDF processing completed: %d pages processed", len(pages))
	return result, nil
}

// processImageFile ????????
func (e *easyOCREngine) processImageFile(ctx context.Context, filename string, content io.Reader, result *OCRResult) (*OCRResult, error) {
	// io.Reader?????????
	img, format, err := image.Decode(content)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to decode image: %w", err)
		return result, result.Error
	}

	log.Printf("Decoded image format: %s", format)

	// OCR??
	text, confidence, err := e.ProcessImage(ctx, img)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("OCR processing failed: %w", err)
		return result, result.Error
	}

	// ?????
	result.ExtractedText = text
	result.Confidence = confidence
	result.Status = "completed"
	result.Pages = []OCRPage{
		{
			PageNumber: 1,
			Text:       text,
			Confidence: confidence,
		},
	}

	return result, nil
}
