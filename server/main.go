package main

import (
	"context"
	"fmt"
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

func (s *server) SayHelloServerStream(in *pb.HelloRequest, stream pb.Greeter_SayHelloServerStreamServer) error {
	log.Printf("Received: %v for server stream", in.GetName())
	for i := 0; i < 5; i++ {
		message := fmt.Sprintf("Hello %s, count %d", in.GetName(), i)
		if err := stream.Send(&pb.HelloReply{Message: message}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
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
