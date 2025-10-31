package application

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "log"
    "strings"
    "time"

    "google.golang.org/grpc/metadata"
    "grpc-sample-minimal/proto"
    "grpc-sample-minimal/server/domain"
)

type ApplicationService struct {
	greeterService domain.GreeterService
	storageService domain.StorageService
	fileRepo       domain.FileMetadataRepository
	ocrClient      domain.OCRClient // OCR??????????????????????
	ocrResultRepo  domain.OCRResultRepository // OCR?????
}

func NewApplicationService(
	greeterService domain.GreeterService, 
	storageService domain.StorageService, 
	fileRepo domain.FileMetadataRepository,
	ocrClient domain.OCRClient,
	ocrResultRepo domain.OCRResultRepository,
) *ApplicationService {
	return &ApplicationService{
		greeterService: greeterService,
		storageService: storageService,
		fileRepo:       fileRepo,
		ocrClient:      ocrClient,
		ocrResultRepo:  ocrResultRepo,
	}
}

func (s *ApplicationService) SayHello(ctx context.Context, name string) (string, error) {
	return s.greeterService.SayHello(ctx, name)
}

func (s *ApplicationService) StreamCounter(ctx context.Context, limit int32, stream proto.Greeter_StreamCounterServer) error {
	return s.greeterService.StreamCounter(ctx, limit, stream)
}

func (s *ApplicationService) Chat(stream proto.Greeter_ChatServer) error {
	return s.greeterService.Chat(stream)
}

func (s *ApplicationService) UploadFile(stream proto.Greeter_UploadFileServer) error {
	// The domain service expects an io.Reader, so we need to adapt the gRPC stream
	// For simplicity, we'll read the entire stream into a buffer first.
	// In a real application, you might want to stream directly to S3.

	var filename string
	var fileContent bytes.Buffer
	var bytesWritten int64
	chunkCount := 0

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			return err
		}
		chunkCount++
		
		// Get filename from any chunk that has it (not just the first)
		if chunk.GetFilename() != "" && filename == "" {
			filename = chunk.GetFilename()
			log.Printf("Received filename in chunk %d: %s", chunkCount, filename)
		}
		
		bytesWritten += int64(len(chunk.GetContent()))
		fileContent.Write(chunk.GetContent())
	}

	log.Printf("Received %d chunks, filename='%s', totalSize=%d", chunkCount, filename, bytesWritten)

	// Validate filename before proceeding
	if filename == "" {
		log.Printf("ERROR: filename is empty after receiving %d chunks", chunkCount)
		return fmt.Errorf("filename is required but was not provided")
	}

    // Choose storage provider from metadata (default: s3)
    storage := s.storageService
    if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
        if vals := md.Get("storage-provider"); len(vals) > 0 {
            switch vals[0] {
            case "gcs":
                if gcs, err := domain.NewGCSStorageService(stream.Context()); err == nil {
                    storage = gcs
                }
            case "azure":
                if azure, err := domain.NewAzureStorageService(stream.Context()); err == nil {
                    storage = azure
                }
            }
        }
    }

    // Validate that we have content before uploading
    if bytesWritten == 0 {
    	return fmt.Errorf("file content is empty")
    }

    status, err := storage.UploadFile(stream.Context(), filename, &fileContent)
	if err != nil {
		return err
	}
	status.BytesWritten = bytesWritten
	
	// Ensure status has the correct filename
	if status.Filename == "" {
		status.Filename = filename
	}

	// Save file metadata to database
	storagePath := domain.BuildStoragePath(filename)
	namespace := domain.GetFileNamespace(filename)
	provider := "s3"
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		if vals := md.Get("storage-provider"); len(vals) > 0 {
			provider = vals[0]
		}
	}
	
	fileMetadata := &domain.FileMetadata{
		Filename:        filename,
		Namespace:       strings.TrimSuffix(namespace, "/"),
		Size:            bytesWritten,
		StorageProvider: provider,
		StoragePath:     storagePath,
		UploadedAt:      time.Now(),
	}
	
	log.Printf("Saving file metadata: filename=%s, namespace=%s, size=%d, provider=%s", 
		fileMetadata.Filename, fileMetadata.Namespace, fileMetadata.Size, fileMetadata.StorageProvider)
	
	if err := s.fileRepo.Create(stream.Context(), fileMetadata); err != nil {
		log.Printf("Warning: Failed to save file metadata to database: %v", err)
		// Continue even if DB save fails
	} else {
		log.Printf("Successfully saved file metadata to database")
	}

	// documents/???images/??????????????????OCR?????
	// namespace?"documents/"???"images/"???????"/"???
	if (namespace == "documents/" || namespace == "images/") && s.ocrClient != nil {
		// ????????????OCR??????
		queueService, err := domain.NewQueueService(provider)
		if err != nil {
			log.Printf("Warning: Failed to create queue service: %v", err)
		} else {
			// ????OCR??????????
			go func() {
				if err := queueService.EnqueueOCRTask(context.Background(), filename, provider); err != nil {
					log.Printf("Warning: Failed to enqueue OCR task: %v", err)
				} else {
					log.Printf("OCR task queued for file: %s (provider: %s)", filename, provider)
				}
			}()
		}
	}

	return stream.SendAndClose(status)
}

func (s *ApplicationService) DownloadFile(req *proto.FileDownloadRequest, stream proto.Greeter_DownloadFileServer) error {
    storage := s.storageService
    provider := req.GetStorageProvider()
    switch provider {
    case "gcs":
        if gcs, err := domain.NewGCSStorageService(stream.Context()); err == nil {
            storage = gcs
        }
    case "azure":
        if azure, err := domain.NewAzureStorageService(stream.Context()); err == nil {
            storage = azure
        }
    }

    reader, err := storage.DownloadFile(stream.Context(), req.GetFilename())
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&proto.FileChunk{Content: buf[:n]}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ApplicationService) ListFiles(ctx context.Context, req *proto.FileListRequest) (*proto.FileListResponse, error) {
	provider := req.GetStorageProvider()
	if provider == "" {
		provider = "s3"
	}

	// Get files from database instead of directly from storage
	// This is more reliable and faster
	files, err := s.fileRepo.ListByProvider(ctx, provider)
	if err != nil {
		log.Printf("Error listing files from database: %v, falling back to storage", err)
		
		// Fallback to storage listing if DB fails
		storage := s.storageService
		switch provider {
		case "gcs":
			if gcs, err := domain.NewGCSStorageService(ctx); err == nil {
				storage = gcs
			}
		case "azure":
			if azure, err := domain.NewAzureStorageService(ctx); err == nil {
				storage = azure
			}
		}
		
		storageFiles, err := storage.ListFiles(ctx)
		if err != nil {
			return nil, err
		}
		files = storageFiles
	}
	
	return &proto.FileListResponse{
		Files: files,
	}, nil
}

// ProcessOCR ?OCR???????????????????????
func (s *ApplicationService) ProcessOCR(ctx context.Context, req *proto.OCRRequest) (*proto.OCRResponse, error) {
	if s.ocrClient == nil {
		return nil, fmt.Errorf("OCR client is not available")
	}
	
	// ???????????????????OCR???????????????
	// ??: ????????OCR?????????????????????????
	// ????OCR?????????????????
	
	// OCR??????????????????????
	// ?????????????OCR??????????????
	_, err := s.ocrClient.ProcessDocument(ctx, req.Filename, nil, req.StorageProvider)
	if err != nil {
		return &proto.OCRResponse{
			TaskId:  "",
			Success: false,
			Message: fmt.Sprintf("Failed to start OCR processing: %v", err),
		}, nil // ?????????????????????
	}
	
	// taskID????????OCR????????????????????
	taskID := req.Filename + "_" + req.StorageProvider + "_" + fmt.Sprintf("%d", time.Now().Unix())
	
	return &proto.OCRResponse{
		TaskId:  taskID,
		Success: true,
		Message: "OCR processing started",
	}, nil
}

// GetOCRResult ?OCR???????
func (s *ApplicationService) GetOCRResult(ctx context.Context, req *proto.OCRResultRequest) (*proto.OCRResultResponse, error) {
	if s.ocrClient == nil {
		return nil, fmt.Errorf("OCR client is not available")
	}
	
	engineName := req.EngineName
	if engineName == "" {
		engineName = "tesseract" // ?????
	}
	
	result, err := s.ocrClient.GetOCRResult(ctx, req.Filename, req.StorageProvider, engineName)
	if err != nil {
		return nil, fmt.Errorf("failed to get OCR result: %w", err)
	}
	
	if result == nil {
		return &proto.OCRResultResponse{
			Filename:  req.Filename,
			EngineName: engineName,
			Status:    "not_found",
		}, nil
	}
	
	// OCRPage???
	pages := make([]*proto.OCRPage, len(result.Pages))
	for i, page := range result.Pages {
		pages[i] = &proto.OCRPage{
			PageNumber: int32(page.PageNumber),
			Text:       page.Text,
			Confidence: page.Confidence,
		}
	}
	
	errorMsg := ""
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}
	
	return &proto.OCRResultResponse{
		Filename:     result.Filename,
		EngineName:   result.EngineName,
		ExtractedText: result.ExtractedText,
		Pages:        pages,
		Status:       result.Status,
		ErrorMessage: errorMsg,
		Confidence:   result.Confidence,
		ProcessedAt:  result.ProcessedAt.Unix(),
	}, nil
}

// ListOCRResults ?OCR?????????
func (s *ApplicationService) ListOCRResults(ctx context.Context, req *proto.OCRListRequest) (*proto.OCRListResponse, error) {
	if s.ocrResultRepo == nil {
		return nil, fmt.Errorf("OCR result repository is not available")
	}
	
	results, err := s.ocrResultRepo.ListOCRResults(ctx, req.StorageProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to list OCR results: %w", err)
	}
	
	summaries := make([]*proto.OCRResultSummary, len(results))
	for i, result := range results {
		summaries[i] = &proto.OCRResultSummary{
			Filename:    result.Filename,
			EngineName:  result.EngineName,
			Status:      result.Status,
			ProcessedAt: result.ProcessedAt.Unix(),
		}
	}
	
	return &proto.OCRListResponse{
		Results: summaries,
	}, nil
}

// CompareOCRResults ????????OCR????????Phase 2B?
func (s *ApplicationService) CompareOCRResults(ctx context.Context, req *proto.OCRComparisonRequest) (*proto.OCRComparisonResponse, error) {
	if s.ocrResultRepo == nil {
		return nil, fmt.Errorf("OCR result repository is not available")
	}
	
	results, err := s.ocrResultRepo.GetOCRComparison(ctx, req.Filename, req.StorageProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get OCR comparison: %w", err)
	}
	
	pbResults := make([]*proto.OCRResultResponse, len(results))
	for i, result := range results {
		pages := make([]*proto.OCRPage, len(result.Pages))
		for j, page := range result.Pages {
			pages[j] = &proto.OCRPage{
				PageNumber: int32(page.PageNumber),
				Text:       page.Text,
				Confidence: page.Confidence,
			}
		}
		
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		
		pbResults[i] = &proto.OCRResultResponse{
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
	
	return &proto.OCRComparisonResponse{
		Filename:      req.Filename,
		StorageProvider: req.StorageProvider,
		Results:       pbResults,
	}, nil
}

func (s *ApplicationService) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	filename := req.GetFilename()
	provider := req.GetStorageProvider()
	if provider == "" {
		provider = "s3"
	}

	// Select storage provider
	storage := s.storageService
	switch provider {
	case "gcs":
		if gcs, err := domain.NewGCSStorageService(ctx); err == nil {
			storage = gcs
		}
	case "azure":
		if azure, err := domain.NewAzureStorageService(ctx); err == nil {
			storage = azure
		}
	}

	// Delete from storage
	if err := storage.DeleteFile(ctx, filename); err != nil {
		log.Printf("Error deleting file from storage: %v", err)
		return &proto.DeleteFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete file from storage: %v", err),
		}, nil
	}

	// Delete from database
	if err := s.fileRepo.Delete(ctx, filename, provider); err != nil {
		log.Printf("Warning: Failed to delete file metadata from database: %v", err)
		// Continue even if DB delete fails, as file is already deleted from storage
	}

	return &proto.DeleteFileResponse{
		Success: true,
		Message: fmt.Sprintf("File %s deleted successfully from %s", filename, provider),
	}, nil
}
