package domain

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	pb "grpc-sample-minimal/proto"
)

type StorageService interface {
	UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error)
	DownloadFile(ctx context.Context, filename string) (io.Reader, error)
	ListFiles(ctx context.Context) ([]*pb.FileInfo, error)
	DeleteFile(ctx context.Context, filename string) error
}

// GetFileNamespace returns the namespace prefix based on file extension
// Returns: "documents/", "media/", or "others/"
func GetFileNamespace(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	
	// Document file extensions
	documentExts := map[string]bool{
		"pdf": true, "doc": true, "docx": true, "txt": true, "md": true,
		"xls": true, "xlsx": true, "ppt": true, "pptx": true, "csv": true,
		"rtf": true, "odt": true, "ods": true, "odp": true,
	}
	
	// Media file extensions (images, videos, audio)
	mediaExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true,
		"bmp": true, "svg": true, "ico": true,
		"mp4": true, "avi": true, "mov": true, "mkv": true, "webm": true,
		"flv": true, "wmv": true, "m4v": true,
		"mp3": true, "wav": true, "flac": true, "aac": true, "ogg": true,
		"m4a": true, "wma": true,
	}
	
	if documentExts[ext] {
		return "documents/"
	}
	if mediaExts[ext] {
		return "media/"
	}
	return "others/"
}

// BuildStoragePath constructs the full storage path with namespace prefix
func BuildStoragePath(filename string) string {
	namespace := GetFileNamespace(filename)
	// Remove any existing namespace prefixes to avoid duplication
	filename = strings.TrimPrefix(filename, "documents/")
	filename = strings.TrimPrefix(filename, "media/")
	filename = strings.TrimPrefix(filename, "others/")
	return namespace + filename
}
