package domain

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// PDFConverter PDF???????????????????
type PDFConverter interface {
	// ConvertPDFToImages PDF??????????
	ConvertPDFToImages(ctx context.Context, pdfContent io.Reader) ([]image.Image, error)
}

// popplerPDFConverter poppler-utils?????PDF????
type popplerPDFConverter struct{}

// NewPDFConverter ???PDFConverter?????
func NewPDFConverter() PDFConverter {
	return &popplerPDFConverter{}
}

// ConvertPDFToImages PDF???????????poppler-utils???
func (c *popplerPDFConverter) ConvertPDFToImages(ctx context.Context, pdfContent io.Reader) ([]image.Image, error) {
	// 1. PDF??????????
	tempPDF, err := savePDFToTempFile(pdfContent)
	if err != nil {
		return nil, fmt.Errorf("failed to save PDF to temp file: %w", err)
	}
	defer func() {
		tempPDF.Close()
		os.Remove(tempPDF.Name())
	}()

	// 2. ??????????????????
	tempDir := filepath.Dir(tempPDF.Name())
	outputPrefix := filepath.Join(tempDir, "pdf_page")
	
	// 3. pdftoppm?PDF?PNG???
	// pdftoppm -png -r 150 input.pdf output_prefix
	// ??: output_prefix-01.png, output_prefix-02.png, ...
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", "150", tempPDF.Name(), outputPrefix)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// 4. ????????????????
	images, err := loadPNGFiles(tempDir, "pdf_page")
	if err != nil {
		return nil, fmt.Errorf("failed to load PNG files: %w", err)
	}

	// 5. ?????PNG???????????????????????????????
	if err := cleanupPNGFiles(tempDir, "pdf_page"); err != nil {
		// ???????????
		fmt.Printf("Warning: Failed to cleanup PNG files: %v\n", err)
	}

	return images, nil
}

// savePDFToTempFile PDF????????????
func savePDFToTempFile(content io.Reader) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "pdf_converter_*.pdf")
	if err != nil {
		return nil, err
	}

	// PDF?????
	if _, err := io.Copy(tempFile, content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	// ????????
	tempFile.Close()

	// ??????????
	readFile, err := os.Open(tempFile.Name())
	if err != nil {
		os.Remove(tempFile.Name())
		return nil, err
	}

	return readFile, nil
}

// loadPNGFiles ????????????????PNG?????????
func loadPNGFiles(dir, prefix string) ([]image.Image, error) {
	// ???????????????
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// ????????????PNG????????????
	var pngFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), ".png") {
			pngFiles = append(pngFiles, filepath.Join(dir, file.Name()))
		}
	}

	// ?????????????????
	sort.Slice(pngFiles, func(i, j int) bool {
		// ????????????????
		return extractPageNumber(pngFiles[i]) < extractPageNumber(pngFiles[j])
	})

	// ?PNG?????????
	images := make([]image.Image, 0, len(pngFiles))
	for _, pngPath := range pngFiles {
		file, err := os.Open(pngPath)
		if err != nil {
			continue // ??????????
		}

		img, _, err := image.Decode(file)
		file.Close()
		
		if err != nil {
			continue // ??????????????
		}

		images = append(images, img)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no PNG files were loaded")
	}

	return images, nil
}

// extractPageNumber ?????????????????
// ?: "pdf_page-01.png" -> 1
func extractPageNumber(filePath string) int {
	base := filepath.Base(filePath)
	// "pdf_page-01.png" -> "01"
	withoutExt := strings.TrimSuffix(base, ".png")
	parts := strings.Split(withoutExt, "-")
	if len(parts) < 2 {
		return 0
	}
	
	// ???????????
	pageNum, _ := strconv.Atoi(parts[len(parts)-1])
	return pageNum
}

// cleanupPNGFiles ?????PNG?????????
func cleanupPNGFiles(dir, prefix string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), ".png") {
			os.Remove(filepath.Join(dir, file.Name()))
		}
	}

	return nil
}
