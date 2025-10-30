package domain

import (
	"context"
	"io"

	pb "grpc-sample-minimal/proto"
)

type StorageService interface {
	UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error)
	DownloadFile(ctx context.Context, filename string) (io.Reader, error)
}
