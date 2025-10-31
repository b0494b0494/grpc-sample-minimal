package domain

import (
    "cloud.google.com/go/storage"
    "bytes"
    "context"
    "fmt"
    pb "grpc-sample-minimal/proto"
    "io"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

var (
    gcsBucketName = os.Getenv("GCS_BUCKET_NAME")
)

type gcsStorageService struct {
    client *storage.Client
}

func NewGCSStorageService(ctx context.Context) (StorageService, error) {
    emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST")
    log.Printf("NewGCSStorageService: STORAGE_EMULATOR_HOST=%s, GCS_BUCKET_NAME=%s", emulatorHost, gcsBucketName)
    
    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCS client: %w", err)
    }

    if gcsBucketName == "" {
        return nil, fmt.Errorf("GCS_BUCKET_NAME is not set")
    }

    // Ensure bucket exists (works against emulator)
    bucket := client.Bucket(gcsBucketName)
    attrs, err := bucket.Attrs(ctx)
    if err != nil {
        log.Printf("NewGCSStorageService: Bucket %s does not exist, creating...", gcsBucketName)
        if err := bucket.Create(ctx, "dev-project", nil); err != nil {
            log.Printf("Warning: failed to create GCS bucket %s: %v", gcsBucketName, err)
        } else {
            log.Printf("NewGCSStorageService: Successfully created bucket %s", gcsBucketName)
        }
    } else {
        log.Printf("NewGCSStorageService: Bucket %s exists (location: %s)", gcsBucketName, attrs.Location)
    }

    return &gcsStorageService{client: client}, nil
}

func (s *gcsStorageService) UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error) {
	// Build storage path with namespace prefix (documents/, media/, or others/)
	storagePath := BuildStoragePath(filename)

    wc := s.client.Bucket(gcsBucketName).Object(storagePath).NewWriter(ctx)
    if _, err := io.Copy(wc, content); err != nil {
        _ = wc.Close()
        return nil, fmt.Errorf("failed to write object to GCS: %w", err)
    }
    if err := wc.Close(); err != nil {
        return nil, fmt.Errorf("failed to close GCS writer: %w", err)
    }

    return &pb.FileUploadStatus{
        Filename:        filename,
        Success:         true,
        Message:         fmt.Sprintf("File %s uploaded to GCS at %s", filename, storagePath),
        StorageProvider: "gcs",
    }, nil
}

func (s *gcsStorageService) DownloadFile(ctx context.Context, filename string) (io.Reader, error) {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)
	log.Printf("GCS DownloadFile: filename=%s, storagePath=%s", filename, storagePath)

    rc, err := s.client.Bucket(gcsBucketName).Object(storagePath).NewReader(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to open GCS object: %w", err)
    }
    defer rc.Close()

    data, err := ioutil.ReadAll(rc)
    if err != nil {
        return nil, fmt.Errorf("failed to read GCS object: %w", err)
    }
    return bytes.NewReader(data), nil
}

func (s *gcsStorageService) DownloadFileByPath(ctx context.Context, storagePath string) (io.Reader, error) {
	log.Printf("GCS DownloadFileByPath: storagePath=%s, bucket=%s, emulator=%s", storagePath, gcsBucketName, os.Getenv("STORAGE_EMULATOR_HOST"))

    rc, err := s.client.Bucket(gcsBucketName).Object(storagePath).NewReader(ctx)
    if err != nil {
        log.Printf("GCS DownloadFileByPath error details: bucket=%s, object=%s, error=%v", gcsBucketName, storagePath, err)
        return nil, fmt.Errorf("failed to open GCS object: %w", err)
    }
    defer rc.Close()

    data, err := ioutil.ReadAll(rc)
    if err != nil {
        return nil, fmt.Errorf("failed to read GCS object: %w", err)
    }
    log.Printf("GCS DownloadFileByPath: Successfully read %d bytes from %s", len(data), storagePath)
    return bytes.NewReader(data), nil
}

func (s *gcsStorageService) DeleteFile(ctx context.Context, filename string) error {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)

	obj := s.client.Bucket(gcsBucketName).Object(storagePath)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file from GCS: %w", err)
	}
	return nil
}

func (s *gcsStorageService) ListFiles(ctx context.Context) ([]*pb.FileInfo, error) {
    var files []*pb.FileInfo
    
    bucket := s.client.Bucket(gcsBucketName)
    it := bucket.Objects(ctx, nil)
    
    for {
        attrs, err := it.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("failed to list GCS objects: %w", err)
        }
        
        // Skip directory prefixes (names ending with "/")
        name := attrs.Name
        if strings.HasSuffix(name, "/") {
            continue
        }
        
        // Extract namespace and filename from name (e.g., "documents/file.pdf")
        var namespace, filename string
        if strings.HasPrefix(name, "documents/") {
            namespace = "documents"
            filename = strings.TrimPrefix(name, "documents/")
        } else if strings.HasPrefix(name, "images/") {
            namespace = "images"
            filename = strings.TrimPrefix(name, "images/")
        } else if strings.HasPrefix(name, "media/") {
            namespace = "media"
            filename = strings.TrimPrefix(name, "media/")
        } else if strings.HasPrefix(name, "others/") {
            namespace = "others"
            filename = strings.TrimPrefix(name, "others/")
        } else {
            // Legacy files without namespace
            namespace = "others"
            filename = name
        }
        
        files = append(files, &pb.FileInfo{
            Filename:  filename,
            Namespace: namespace,
            Size:      attrs.Size,
        })
    }
    
    return files, nil
}
