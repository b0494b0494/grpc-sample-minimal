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
)

var (
    gcsBucketName = os.Getenv("GCS_BUCKET_NAME")
)

type gcsStorageService struct {
    client *storage.Client
}

func NewGCSStorageService(ctx context.Context) (StorageService, error) {
    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCS client: %w", err)
    }

    if gcsBucketName == "" {
        return nil, fmt.Errorf("GCS_BUCKET_NAME is not set")
    }

    // Ensure bucket exists (works against emulator)
    bucket := client.Bucket(gcsBucketName)
    if _, err := bucket.Attrs(ctx); err != nil {
        if err := bucket.Create(ctx, "dev-project", nil); err != nil {
            log.Printf("Warning: failed to create GCS bucket %s: %v", gcsBucketName, err)
        }
    }

    return &gcsStorageService{client: client}, nil
}

func (s *gcsStorageService) UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error) {
	// Build storage path with namespace prefix (documents/, media/, or others/)
	storagePath := buildStoragePath(filename)

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
	storagePath := buildStoragePath(filename)

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
