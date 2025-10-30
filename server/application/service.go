package application

import (
	"context"
	"bytes"
	"io"

	"grpc-sample-minimal/proto"
	"grpc-sample-minimal/server/domain"
)

type ApplicationService struct {
	greeterService domain.GreeterService
	storageService domain.StorageService
}

func NewApplicationService(greeterService domain.GreeterService, storageService domain.StorageService) *ApplicationService {
	return &ApplicationService{
		greeterService: greeterService,
		storageService: storageService,
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

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filename == "" {
			filename = chunk.GetFilename()
		}
		bytesWritten += int64(len(chunk.GetContent()))
		fileContent.Write(chunk.GetContent())
	}

	status, err := s.storageService.UploadFile(stream.Context(), filename, &fileContent)
	if err != nil {
		return err
	}
	status.BytesWritten = bytesWritten

	return stream.SendAndClose(status)
}

func (s *ApplicationService) DownloadFile(req *proto.FileDownloadRequest, stream proto.Greeter_DownloadFileServer) error {
	reader, err := s.storageService.DownloadFile(stream.Context(), req.GetFilename())
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
