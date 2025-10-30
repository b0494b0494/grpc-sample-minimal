package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "grpc-sample-minimal/proto"
	"grpc-sample-minimal/server/application"
	"grpc-sample-minimal/server/domain"
)

const (
	port = ":50051"
	authToken = "my-secret-token"
)

// server is used to implement proto.GreeterServer.
type server struct{
	pb.UnimplementedGreeterServer
	appService *application.ApplicationService
}

// authInterceptor is a unary interceptor that checks for a valid auth token.
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 || values[0] != authToken {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	log.Printf("Auth successful for method: %s", info.FullMethod)
	return handler(ctx, req)
}

// loggingInterceptor is a unary interceptor that logs RPC calls.
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("Incoming RPC: %s, Request: %v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	log.Printf("Outgoing RPC: %s, Response: %v, Error: %v", info.FullMethod, resp, err)
	return resp, err
}

// authStreamInterceptor is a stream interceptor that checks for a valid auth token.
func authStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 || values[0] != authToken {
		return status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	log.Printf("Auth successful for stream method: %s", info.FullMethod)
	return handler(srv, ss)
}

// loggingStreamInterceptor is a stream interceptor that logs RPC calls.
func loggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("Incoming stream RPC: %s", info.FullMethod)
	err := handler(srv, ss)
	log.Printf("Outgoing stream RPC: %s, Error: %v", info.FullMethod, err)
	return err
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	msg, err := s.appService.SayHello(ctx, in.GetName())
	if err != nil {
		return nil, err
	}
	return &pb.HelloReply{Message: msg}, nil
}

func (s *server) StreamCounter(in *pb.CounterRequest, stream pb.Greeter_StreamCounterServer) error {
	return s.appService.StreamCounter(stream.Context(), in.GetLimit(), stream)
}

func (s *server) Chat(stream pb.Greeter_ChatServer) error {
	return s.appService.Chat(stream)
}

func main() {
	domainService := domain.NewGreeterService()
	appService := application.NewApplicationService(domainService)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor, authInterceptor),
		grpc.ChainStreamInterceptor(loggingStreamInterceptor, authStreamInterceptor),
	)
	pb.RegisterGreeterServer(s, &server{appService: appService})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}