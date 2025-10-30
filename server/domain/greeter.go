package domain

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pb "grpc-sample-minimal/proto"
)

type GreeterService interface {
	SayHello(ctx context.Context, name string) (string, error)
	StreamCounter(ctx context.Context, limit int32, stream pb.Greeter_StreamCounterServer) error
	Chat(stream pb.Greeter_ChatServer) error
	UploadFile(stream pb.Greeter_UploadFileServer) error
	DownloadFile(req *pb.FileDownloadRequest, stream pb.Greeter_DownloadFileServer) error
}

type greeterService struct{}

func NewGreeterService() GreeterService {
	return &greeterService{}
}

func (s *greeterService) SayHello(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello %s", name), nil
}

func (s *greeterService) StreamCounter(ctx context.Context, limit int32, stream pb.Greeter_StreamCounterServer) error {
	for i := 0; i < int(limit); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := stream.Send(&pb.CounterReply{Count: int32(i + 1)}); err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
	}
	return nil
}

func (s *greeterService) Chat(stream pb.Greeter_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// For simplicity, just echo back the message
		if err := stream.Send(&pb.ChatMessage{User: "Server", Message: "Echo: " + in.GetMessage()}); err != nil {
			return err
		}
	}
}

func (s *greeterService) UploadFile(stream pb.Greeter_UploadFileServer) error {
	var file *os.File
	var filename string
	var bytesWritten int64

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if file == nil {
			filename = chunk.GetFilename()
			file, err = os.Create("uploads/" + filename)
			if err != nil {
				return err
			}
			defer file.Close()
		}

		n, err := file.Write(chunk.GetContent())
		if err != nil {
			return err
		}
		bytesWritten += int64(n)
	}

	return stream.SendAndClose(&pb.FileUploadStatus{
		Filename:    filename,
		BytesWritten: bytesWritten,
		Success:     true,
		Message:     "File uploaded successfully",
	})
}

func (s *greeterService) DownloadFile(req *pb.FileDownloadRequest, stream pb.Greeter_DownloadFileServer) error {
	filename := req.GetFilename()
	file, err := os.Open("uploads/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.FileChunk{Content: buf[:n]}); err != nil {
			return err
		}
	}
	return nil
}
