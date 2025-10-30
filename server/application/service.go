package application

import (
	"context"

	"grpc-sample-minimal/proto"
	"grpc-sample-minimal/server/domain"
)

type ApplicationService struct {
	greeterService domain.GreeterService
}

func NewApplicationService(greeterService domain.GreeterService) *ApplicationService {
	return &ApplicationService{
		greeterService: greeterService,
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
	return s.greeterService.UploadFile(stream)
}

func (s *ApplicationService) DownloadFile(req *proto.FileDownloadRequest, stream proto.Greeter_DownloadFileServer) error {
	return s.greeterService.DownloadFile(req, stream)
}
