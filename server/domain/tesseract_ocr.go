package domain

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
)

// tesseractEngine ?Tesseract OCR???????
type tesseractEngine struct {
	language string // OCR??: "jpn+eng"??
}

// NewTesseractEngine ????Tesseract OCR?????????
func NewTesseractEngine(language string) OCREngine {
	if language == "" {
		language = "jpn+eng" // ?????: ???+??
	}
	return &tesseractEngine{
		language: language,
	}
}

// Name ?????????
func (e *tesseractEngine) Name() string {
	return "tesseract"
}

// ProcessDocument ???????????????OCR?????
func (e *tesseractEngine) ProcessDocument(ctx context.Context, filename string, content io.Reader) (*OCRResult, error) {
	result := &OCRResult{
		Filename:    filename,
		EngineName:  e.Name(),
		Status:      "processing",
		ProcessedAt: time.Now(),
	}

	// ??????
	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	// PDF??????poppler-utils???
	if ext == "pdf" {
		return e.processPDF(ctx, filename, content, result)
	}

	// ????????
	if isImageFile(ext) {
		return e.processImageFile(ctx, filename, content, result)
	}

	// ???????????????????
	// TODO: ????????????????????PDF???????
	result.Status = "failed"
	result.Error = fmt.Errorf("unsupported file type: %s", ext)
	return result, result.Error
}

// ProcessImage ????OCR?????
func (e *tesseractEngine) ProcessImage(ctx context.Context, img image.Image) (string, float64, error) {
	// gosseract?OCR??
	client := gosseract.NewClient()
	defer client.Close()
	
	// ????
	if err := client.SetLanguage(e.language); err != nil {
		return "", 0.0, fmt.Errorf("failed to set language: %w", err)
	}
	
	// ?????????????gosseract?????????????
	tempFile, err := saveImageToTempFile(img)
	if err != nil {
		return "", 0.0, fmt.Errorf("failed to save image to temp file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()
	
	// OCR??
	client.SetImage(tempFile.Name())
	text, err := client.Text()
	if err != nil {
		return "", 0.0, fmt.Errorf("failed to extract text: %w", err)
	}
	
	// ?????????gosseract v2??MeanConfidence?????????????0.0????
	// TODO: HOCR?????????????????
	confidence := 0.0
	
	return text, confidence, nil
}

// saveImageToTempFile ???????????????
func saveImageToTempFile(img image.Image) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "tesseract_ocr_*.png")
	if err != nil {
		return nil, err
	}
	
	// PNG?????
	if err := png.Encode(tempFile, img); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}
	
	// ??????????gosseract?????????
	tempFile.Close()
	
	// ??????????
	readFile, err := os.Open(tempFile.Name())
	if err != nil {
		os.Remove(tempFile.Name())
		return nil, err
	}
	
	return readFile, nil
}

// processPDF PDF??????poppler-utils???
func (e *tesseractEngine) processPDF(ctx context.Context, filename string, content io.Reader, result *OCRResult) (*OCRResult, error) {
	// PDF??????
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
			// ??????????????
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

// processImageFile ????????????
func (e *tesseractEngine) processImageFile(ctx context.Context, filename string, content io.Reader, result *OCRResult) (*OCRResult, error) {
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

// isImageFile ????????????????????
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true,
		"bmp": true, "svg": true, "ico": true,
	}
	return imageExts[ext]
}
