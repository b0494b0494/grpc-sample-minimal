package main

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	pb "grpc-sample-minimal/proto"
)

const (
	port = ":50051"
)

// server is used to implement proto.GreeterServer.
type server struct{
	pb.UnimplementedGreeterServer
}

// SayHello implements proto.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()},
	 nil
}

func (s *server) StreamCounter(in *pb.CounterRequest, stream pb.Greeter_StreamCounterServer) error {
	log.Printf("Received StreamCounter request with limit: %d", in.GetLimit())
	for i := 0; i < int(in.GetLimit()); i++ {
		if err := stream.Send(&pb.CounterReply{Count: int32(i + 1)}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (s *server) Chat(stream pb.Greeter_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("Chat message from %s: %s", in.GetUser(), in.GetMessage())
		if err := stream.Send(&pb.ChatMessage{User: "Server", Message: "Echo: " + in.GetMessage()}); err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
