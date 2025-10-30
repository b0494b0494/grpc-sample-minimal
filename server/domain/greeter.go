package domain

import (
	"context"
	"fmt"
	"io"
	"time"

	pb "grpc-sample-minimal/proto"
)

type GreeterService interface {
	SayHello(ctx context.Context, name string) (string, error)
	StreamCounter(ctx context.Context, limit int32, stream pb.Greeter_StreamCounterServer) error
	Chat(stream pb.Greeter_ChatServer) error
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
