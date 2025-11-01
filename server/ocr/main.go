package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "grpc-sample-minimal/proto"
	"grpc-sample-minimal/server/domain"
)

const (
	defaultPort = ":50052"
)

var (
	authToken = os.Getenv("AUTH_TOKEN")
)

// ocrServer ?OCR?????????gRPC????
type ocrServer struct {
	pb.UnimplementedGreeterServer
	ocrService       domain.OCRService
	ocrResultRepo    domain.OCRResultRepository
	fileMetadataRepo domain.FileMetadataRepository
	getStorageService func(ctx context.Context, provider string) (domain.StorageService, error)
}

// ProcessOCR ?OCR???????????
func (s *ocrServer) ProcessOCR(ctx context.Context, req *pb.OCRRequest) (*pb.OCRResponse, error) {
	// ???ID???
	taskID := req.Filename + "_" + req.StorageProvider + "_" + fmt.Sprintf("%d", time.Now().Unix())
	
	// ????OCR?????
	go s.processOCRAsync(context.Background(), req.Filename, req.StorageProvider)
	
	return &pb.OCRResponse{
		TaskId:  taskID,
		Success: true,
		Message: "OCR processing started",
	}, nil
}

// processOCRAsync ?????OCR???????
func (s *ocrServer) processOCRAsync(ctx context.Context, filename string, storageProvider string) {
	// panic??: defer + recover
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in OCR processing: %v", r)
			log.Printf("Panic occurred during OCR processing for file %s (provider: %s): %v", filename, storageProvider, r)
			s.saveFailedResult(ctx, filename, storageProvider, err)
		}
	}()
	
	log.Printf("Starting OCR processing for file: %s (provider: %s)", filename, storageProvider)
	
	// 1. ??????????????????
	storageService, err := s.getStorageService(ctx, storageProvider)
	if err != nil {
		log.Printf("Failed to get storage service: %v", err)
		s.saveFailedResult(ctx, filename, storageProvider, err)
		return
	}
	
	// ???????????storage_path???
	var contentReader io.Reader
	if s.fileMetadataRepo != nil {
		log.Printf("Looking up file metadata: filename=%s, provider=%s", filename, storageProvider)
		metadata, err := s.fileMetadataRepo.FindByFilename(ctx, filename, storageProvider)
		if err != nil {
			log.Printf("Error finding file metadata: %v, using filename to build path", err)
			contentReader, err = storageService.DownloadFile(ctx, filename)
		} else if metadata != nil && metadata.StoragePath != "" {
			// ???????????????storage_path???
			log.Printf("Using storage_path from DB: %s", metadata.StoragePath)
			contentReader, err = storageService.DownloadFileByPath(ctx, metadata.StoragePath)
		if err != nil {
			log.Printf("Failed to download file by path: %v, trying with filename", err)
			// ???????: filename???
			var fallbackErr error
			contentReader, fallbackErr = storageService.DownloadFile(ctx, filename)
			if fallbackErr != nil {
				log.Printf("Failed to download file with filename fallback: %v", fallbackErr)
				err = fallbackErr // ????????????
			} else {
				log.Printf("Successfully downloaded file with filename fallback")
				err = nil // ??????????????????
			}
		} else {
			log.Printf("Successfully downloaded file using storage_path from DB")
		}
		} else {
			// ???????????????filename???
			log.Printf("File metadata not found in DB (metadata=%v), using filename to build path", metadata)
			contentReader, err = storageService.DownloadFile(ctx, filename)
		}
	} else {
		// fileMetadataRepo??????filename???
		log.Printf("fileMetadataRepo is nil, using filename to build path")
		contentReader, err = storageService.DownloadFile(ctx, filename)
	}
	
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		s.saveFailedResult(ctx, filename, storageProvider, err)
		return
	}
	
	if contentReader == nil {
		log.Printf("contentReader is nil after download attempt")
		s.saveFailedResult(ctx, filename, storageProvider, fmt.Errorf("contentReader is nil"))
		return
	}
	
	// 2. OCR?????????????????
	engineNames := getEngineNames()
	results, err := s.ocrService.ProcessDocument(ctx, filename, contentReader, engineNames)
	if err != nil {
		log.Printf("Failed to process OCR: %v", err)
		s.saveFailedResult(ctx, filename, storageProvider, err)
		return
	}
	
	// 3. ??????????????
	if len(results) == 0 {
		log.Printf("No OCR results returned for file: %s", filename)
		s.saveFailedResult(ctx, filename, storageProvider, fmt.Errorf("no OCR results returned"))
		return
	}
	
	// ???????????
	for engineName, result := range results {
		if result == nil {
			log.Printf("Warning: %s engine result is nil for file: %s", engineName, filename)
			continue
		}
		result.StorageProvider = storageProvider
		result.Status = "completed"
		if result.ProcessedAt.IsZero() {
			result.ProcessedAt = time.Now()
		}
		if err := s.ocrResultRepo.SaveOCRResult(ctx, result); err != nil {
			log.Printf("Failed to save OCR result for engine %s: %v", engineName, err)
		} else {
			log.Printf("OCR result saved for engine %s: %s", engineName, filename)
		}
	}
	
	log.Printf("OCR processing completed for file: %s with %d engine(s)", filename, len(results))
}

// saveFailedResult ?????OCR???????
func (s *ocrServer) saveFailedResult(ctx context.Context, filename string, storageProvider string, err error) {
	result := &domain.OCRResult{
		Filename:        filename,
		StorageProvider: storageProvider,
		EngineName:     "tesseract",
		Status:         "failed",
		Error:          err,
		ProcessedAt:    time.Now(),
	}
	
	if saveErr := s.ocrResultRepo.SaveOCRResult(ctx, result); saveErr != nil {
		log.Printf("Failed to save failed OCR result: %v", saveErr)
	}
}

// GetOCRResult ?OCR???????
func (s *ocrServer) GetOCRResult(ctx context.Context, req *pb.OCRResultRequest) (*pb.OCRResultResponse, error) {
	engineName := req.EngineName
	if engineName == "" {
		engineName = "tesseract" // ?????
	}
	
	result, err := s.ocrResultRepo.GetOCRResult(ctx, req.Filename, req.StorageProvider, engineName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get OCR result: %v", err)
	}
	
	if result == nil {
		return &pb.OCRResultResponse{
			Filename:  req.Filename,
			EngineName: engineName,
			Status:    "not_found",
		}, nil
	}
	
	// OCRPage???
	pages := make([]*pb.OCRPage, len(result.Pages))
	for i, page := range result.Pages {
		pages[i] = &pb.OCRPage{
			PageNumber: int32(page.PageNumber),
			Text:       page.Text,
			Confidence: page.Confidence,
		}
	}
	
	errorMsg := ""
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}
	
	return &pb.OCRResultResponse{
		Filename:     result.Filename,
		EngineName:   result.EngineName,
		ExtractedText: result.ExtractedText,
		Pages:        pages,
		Status:       result.Status,
		ErrorMessage: errorMsg,
		Confidence:   result.Confidence,
		ProcessedAt: result.ProcessedAt.Unix(),
	}, nil
}

// ListOCRResults ?OCR?????????
func (s *ocrServer) ListOCRResults(ctx context.Context, req *pb.OCRListRequest) (*pb.OCRListResponse, error) {
	results, err := s.ocrResultRepo.ListOCRResults(ctx, req.StorageProvider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list OCR results: %v", err)
	}
	
	summaries := make([]*pb.OCRResultSummary, len(results))
	for i, result := range results {
		summaries[i] = &pb.OCRResultSummary{
			Filename:    result.Filename,
			EngineName:  result.EngineName,
			Status:      result.Status,
			ProcessedAt: result.ProcessedAt.Unix(),
		}
	}
	
	return &pb.OCRListResponse{
		Results: summaries,
	}, nil
}

// CompareOCRResults ????????OCR????????Phase 2B?
func (s *ocrServer) CompareOCRResults(ctx context.Context, req *pb.OCRComparisonRequest) (*pb.OCRComparisonResponse, error) {
	results, err := s.ocrResultRepo.GetOCRComparison(ctx, req.Filename, req.StorageProvider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get OCR comparison: %v", err)
	}
	
	pbResults := make([]*pb.OCRResultResponse, len(results))
	for i, result := range results {
		pages := make([]*pb.OCRPage, len(result.Pages))
		for j, page := range result.Pages {
			pages[j] = &pb.OCRPage{
				PageNumber: int32(page.PageNumber),
				Text:       page.Text,
				Confidence: page.Confidence,
			}
		}
		
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		
		pbResults[i] = &pb.OCRResultResponse{
			Filename:     result.Filename,
			EngineName:   result.EngineName,
			ExtractedText: result.ExtractedText,
			Pages:        pages,
			Status:       result.Status,
			ErrorMessage: errorMsg,
			Confidence:   result.Confidence,
			ProcessedAt:  result.ProcessedAt.Unix(),
		}
	}
	
	return &pb.OCRComparisonResponse{
		Filename:      req.Filename,
		StorageProvider: req.StorageProvider,
		Results:       pbResults,
	}, nil
}

// getEngineNames ??????OCR????????????????: tesseract?
func getEngineNames() []string {
	enginesEnv := os.Getenv("OCR_ENGINES")
	if enginesEnv == "" {
		return []string{"tesseract"}
	}
	
	engines := strings.Split(enginesEnv, ",")
	result := make([]string, 0, len(engines))
	for _, e := range engines {
		e = strings.TrimSpace(e)
		if e != "" {
			result = append(result, e)
		}
	}
	
	if len(result) == 0 {
		return []string{"tesseract"}
	}
	return result
}

// containsString ????????????????????????
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// authInterceptor ???????????
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 || values[0] != authToken {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	log.Printf("OCR Service: Auth successful for method: %s", info.FullMethod)
	return handler(ctx, req)
}

func main() {
	// OCR????????
	ocrService := domain.NewOCRService()
	
	// Tesseract???????
	tesseractEngine := domain.NewTesseractEngine("jpn+eng")
	ocrService.RegisterEngine(tesseractEngine)
	
	// OCR???????????
	ocrResultRepo, err := domain.NewOCRResultRepository(context.Background())
	if err != nil {
		log.Fatalf("failed to create OCR result repository: %v", err)
	}
	defer func() {
		if closer, ok := ocrResultRepo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("Error closing OCR result repository: %v", err)
			}
		}
	}()
	
	port := os.Getenv("OCR_SERVICE_PORT")
	if port == "" {
		port = defaultPort
	}
	port = ":" + port

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)
	
	// ????????????????
	getStorageService := func(ctx context.Context, provider string) (domain.StorageService, error) {
		switch provider {
		case "azure":
			return domain.NewAzureStorageService(ctx)
		case "s3":
			return domain.NewS3StorageService()
		case "gcs":
			return domain.NewGCSStorageService(ctx)
		default:
			return nil, fmt.Errorf("unsupported storage provider: %s", provider)
		}
	}
	
	// ????????????????????????
	fileMetadataRepo, err := domain.NewFileMetadataRepository(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to create file metadata repository: %v", err)
		fileMetadataRepo = nil // nil?????????????
	}
	defer func() {
		if closer, ok := fileMetadataRepo.(interface{ Close() error }); ok && closer != nil {
			if err := closer.Close(); err != nil {
				log.Printf("Error closing file metadata repository: %v", err)
			}
		}
	}()
	
	ocrSrv := &ocrServer{
		ocrService:       ocrService,
		ocrResultRepo:    ocrResultRepo,
		fileMetadataRepo: fileMetadataRepo,
		getStorageService: getStorageService,
	}
	
	// ????????????????????????????????????????????
	ctx := context.Background()
	go startOCRWorker(ctx, "azure", ocrService, ocrResultRepo, fileMetadataRepo, getStorageService)
	go startOCRWorker(ctx, "s3", ocrService, ocrResultRepo, fileMetadataRepo, getStorageService)
	go startOCRWorker(ctx, "gcs", ocrService, ocrResultRepo, fileMetadataRepo, getStorageService)
	
	log.Printf("OCR workers started for all storage providers")
	
	// OCR?????RPC????????
	// ??: ?????proto.GreeterServer??????????????????
	// RegisterGreeterServer?OCR???RPC?????
	pb.RegisterGreeterServer(s, ocrSrv)
	log.Printf("OCR service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startOCRWorker ??????OCR???????????????????
func startOCRWorker(
	ctx context.Context,
	storageProvider string,
	ocrService domain.OCRService,
	ocrResultRepo domain.OCRResultRepository,
	fileMetadataRepo domain.FileMetadataRepository,
	getStorageService func(ctx context.Context, provider string) (domain.StorageService, error),
) {
	queueManager := domain.GetQueueManager()
	if !queueManager.IsEnabled() {
		log.Printf("Warning: QueueManager is disabled for %s", storageProvider)
		return
	}

	log.Printf("OCR worker started for storage provider: %s (via QueueManager)", storageProvider)

	log.Printf("Debug: %s worker loop started, will call DequeueOCRTask repeatedly", storageProvider)
	loopCount := 0
	for {
		loopCount++
		if loopCount%10 == 0 {
			log.Printf("Debug: %s worker loop iteration %d, calling DequeueOCRTask", storageProvider, loopCount)
		}
		// ??????????????????
		task, err := queueManager.DequeueOCRTask(ctx, storageProvider)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				log.Printf("Debug: %s worker context canceled or deadline exceeded", storageProvider)
				break
			}
			log.Printf("Error dequeuing OCR task via QueueManager (provider: %s): %v", storageProvider, err)
			// ???????????????
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				continue
			}
		}

		if task == nil {
			// ???????????????????
			if loopCount%20 == 0 {
				log.Printf("Debug: %s worker received nil task (no messages available), will retry", storageProvider)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}

		log.Printf("Processing OCR task: file=%s, provider=%s", task.Filename, task.StorageProvider)

		// OCR??????defer + recover?panic????
		func() {
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic in OCR processing: %v", r)
					log.Printf("Panic occurred during OCR processing for file %s: %v", task.Filename, r)
					saveFailedResult(ctx, task.Filename, task.StorageProvider, ocrResultRepo, err)
				}
			}()
			processOCRTask(ctx, task.Filename, task.StorageProvider, ocrService, ocrResultRepo, fileMetadataRepo, getStorageService)
		}()
	}
}

// processOCRTask ?OCR?????????processOCRAsync????????
func processOCRTask(
	ctx context.Context,
	filename string,
	storageProvider string,
	ocrService domain.OCRService,
	ocrResultRepo domain.OCRResultRepository,
	fileMetadataRepo domain.FileMetadataRepository,
	getStorageService func(ctx context.Context, provider string) (domain.StorageService, error),
) {
	log.Printf("Starting OCR processing for file: %s (provider: %s)", filename, storageProvider)
	
	// 1. ??????????????????
	storageService, err := getStorageService(ctx, storageProvider)
	if err != nil {
		log.Printf("Failed to get storage service: %v", err)
		saveFailedResult(ctx, filename, storageProvider, ocrResultRepo, err)
		return
	}
	
	// ???????????storage_path???
	var contentReader io.Reader
	if fileMetadataRepo != nil {
		log.Printf("Looking up file metadata: filename=%s, provider=%s", filename, storageProvider)
		metadata, err := fileMetadataRepo.FindByFilename(ctx, filename, storageProvider)
		if err != nil {
			log.Printf("Error finding file metadata: %v, using filename to build path", err)
			contentReader, err = storageService.DownloadFile(ctx, filename)
		} else if metadata != nil && metadata.StoragePath != "" {
			// ???????????????storage_path???
			log.Printf("Using storage_path from DB: %s", metadata.StoragePath)
			contentReader, err = storageService.DownloadFileByPath(ctx, metadata.StoragePath)
		if err != nil {
			log.Printf("Failed to download file by path: %v, trying with filename", err)
			// ???????: filename???
			var fallbackErr error
			contentReader, fallbackErr = storageService.DownloadFile(ctx, filename)
			if fallbackErr != nil {
				log.Printf("Failed to download file with filename fallback: %v", fallbackErr)
				err = fallbackErr // ????????????
			} else {
				log.Printf("Successfully downloaded file with filename fallback")
				err = nil // ??????????????????
			}
		} else {
			log.Printf("Successfully downloaded file using storage_path from DB")
		}
		} else {
			// ???????????????filename???
			log.Printf("File metadata not found in DB (metadata=%v), using filename to build path", metadata)
			contentReader, err = storageService.DownloadFile(ctx, filename)
		}
	} else {
		// fileMetadataRepo??????filename???
		log.Printf("fileMetadataRepo is nil, using filename to build path")
		contentReader, err = storageService.DownloadFile(ctx, filename)
	}
	
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		saveFailedResult(ctx, filename, storageProvider, ocrResultRepo, err)
		return
	}
	
	// 2. OCR?????????????????
	engineNames := getEngineNames()
	results, err := ocrService.ProcessDocument(ctx, filename, contentReader, engineNames)
	if err != nil {
		log.Printf("Failed to process OCR: %v", err)
		saveFailedResult(ctx, filename, storageProvider, ocrResultRepo, err)
		return
	}
	
	// 3. ??????????????
	if len(results) == 0 {
		log.Printf("No OCR results returned for file: %s", filename)
		saveFailedResult(ctx, filename, storageProvider, ocrResultRepo, fmt.Errorf("no OCR results returned"))
		return
	}
	
	// ???????????
	for engineName, result := range results {
		if result == nil {
			log.Printf("Warning: %s engine result is nil for file: %s", engineName, filename)
			continue
		}
		result.StorageProvider = storageProvider
		result.Status = "completed"
		if result.ProcessedAt.IsZero() {
			result.ProcessedAt = time.Now()
		}
		if err := ocrResultRepo.SaveOCRResult(ctx, result); err != nil {
			log.Printf("Failed to save OCR result for engine %s: %v", engineName, err)
		} else {
			log.Printf("OCR result saved for engine %s: %s", engineName, filename)
		}
	}
	
	log.Printf("OCR processing completed for file: %s with %d engine(s)", filename, len(results))
}

// saveFailedResult ?????OCR???????
// ?????????????????????????????????????
func saveFailedResult(ctx context.Context, filename string, storageProvider string, ocrResultRepo domain.OCRResultRepository, err error) {
	errorType := "ocr_error"
	if err != nil {
		errStr := err.Error()
		if contains(errStr, "storage") || contains(errStr, "download") {
			errorType = "storage_error"
		} else if contains(errStr, "database") || contains(errStr, "save") {
			errorType = "db_error"
		} else if contains(errStr, "panic") {
			errorType = "panic"
		}
	}
	
	result := &domain.OCRResult{
		Filename:        filename,
		StorageProvider: storageProvider,
		EngineName:      "tesseract",
		Status:          "failed",
		Error:           err,
		ProcessedAt:     time.Now(),
	}
	
	// ??????????????3??
	maxRetries := 3
	var lastSaveErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if saveErr := ocrResultRepo.SaveOCRResult(ctx, result); saveErr != nil {
			lastSaveErr = saveErr
			log.Printf("Attempt %d/%d: Failed to save failed OCR result: %v", attempt, maxRetries, saveErr)
			if attempt < maxRetries {
				// ??????????
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
		} else {
			log.Printf("Successfully saved failed OCR result: filename=%s, provider=%s, error=%v", filename, storageProvider, err)
			// ????????????????????
			if logErr := ocrResultRepo.LogError(ctx, filename, storageProvider, "tesseract", errorType, err.Error()); logErr != nil {
				log.Printf("Warning: Failed to log error to error_logs table: %v", logErr)
			}
			return
		}
	}
	
	// ???????????????????????????
	if lastSaveErr != nil {
		log.Printf("CRITICAL: Failed to save failed OCR result after %d attempts. Filename: %s, Provider: %s, Original Error: %v, Save Error: %v", 
			maxRetries, filename, storageProvider, err, lastSaveErr)
		
		// ????????????????????????
		if logErr := ocrResultRepo.LogError(ctx, filename, storageProvider, "tesseract", "db_error", 
			fmt.Sprintf("Failed to save OCR result: %v (Original: %v)", lastSaveErr, err)); logErr != nil {
			log.Printf("CRITICAL: Failed to log error to error_logs table: %v", logErr)
		}
	}
}

// contains ????????????????????????????????????
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
