package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "grpc-sample-minimal/proto"
	"grpc-sample-minimal/server/application"
	"grpc-sample-minimal/server/domain"
)

const (
	defaultPort = ":50051"
)

var (
	authToken = os.Getenv("AUTH_TOKEN")
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

func (s *server) UploadFile(stream pb.Greeter_UploadFileServer) error {
	return s.appService.UploadFile(stream)
}

func (s *server) DownloadFile(req *pb.FileDownloadRequest, stream pb.Greeter_DownloadFileServer) error {
	return s.appService.DownloadFile(req, stream)
}

func (s *server) ListFiles(ctx context.Context, req *pb.FileListRequest) (*pb.FileListResponse, error) {
	return s.appService.ListFiles(ctx, req)
}

func (s *server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	return s.appService.DeleteFile(ctx, req)
}

func (s *server) ProcessOCR(ctx context.Context, req *pb.OCRRequest) (*pb.OCRResponse, error) {
	return s.appService.ProcessOCR(ctx, req)
}

func (s *server) GetOCRResult(ctx context.Context, req *pb.OCRResultRequest) (*pb.OCRResultResponse, error) {
	return s.appService.GetOCRResult(ctx, req)
}

func (s *server) ListOCRResults(ctx context.Context, req *pb.OCRListRequest) (*pb.OCRListResponse, error) {
	return s.appService.ListOCRResults(ctx, req)
}

func (s *server) CompareOCRResults(ctx context.Context, req *pb.OCRComparisonRequest) (*pb.OCRComparisonResponse, error) {
	return s.appService.CompareOCRResults(ctx, req)
}

func main() {
	domainService := domain.NewGreeterService()
	storageService, err := domain.NewS3StorageService()
	if err != nil {
		log.Fatalf("failed to create S3 storage service: %v", err)
	}
	
	// Initialize file metadata repository (SQLite)
	fileRepo, err := domain.NewFileMetadataRepository(context.Background())
	if err != nil {
		log.Fatalf("failed to create file metadata repository: %v", err)
	}
	defer func() {
		if closer, ok := fileRepo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("Error closing file repository: %v", err)
			}
		}
	}()
	
	// OCR?????????????????gRPC????????
	// ??OCR?????????: MultiOCRClient???
	var ocrClient domain.OCRClient
	var multiOCRClient *domain.MultiOCRClient
	
	// ????????????????????MultiOCRClient???
	tesseractEndpoint := os.Getenv("OCR_TESSERACT_ENDPOINT")
	easyOCREndpoint := os.Getenv("OCR_EASYOCR_ENDPOINT")
	
	if tesseractEndpoint != "" || easyOCREndpoint != "" {
		// ????????
		multi, err := domain.NewMultiOCRClient(authToken)
		if err != nil {
			log.Printf("Warning: Failed to create MultiOCR client: %v (OCR features will be unavailable)", err)
		} else {
			multiOCRClient = multi
			log.Printf("MultiOCR client initialized with engines: %v", multi.GetAvailableEngines())
			// ???????????????????ocrClient?????
			engines := multi.GetAvailableEngines()
			if len(engines) > 0 {
				ocrClient = multi.GetClient(engines[0])
			}
		}
		defer func() {
			if multiOCRClient != nil {
				if err := multiOCRClient.Close(); err != nil {
					log.Printf("Error closing MultiOCR client: %v", err)
				}
			}
		}()
	} else {
		// ????????????????
		ocrEndpoint := os.Getenv("OCR_SERVICE_ENDPOINT")
		if ocrEndpoint == "" {
			ocrEndpoint = "ocr-service:50052" // ?????
		}
		
		client, err := domain.NewOCRClient(ocrEndpoint, authToken)
		if err != nil {
			log.Printf("Warning: Failed to create OCR client: %v (OCR features will be unavailable)", err)
		} else {
			ocrClient = client
		}
	}
	
	// OCR??????????????DB????
	ocrResultRepo, err := domain.NewOCRResultRepository(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to create OCR result repository: %v", err)
	}
	defer func() {
		if ocrResultRepo != nil {
			if closer, ok := ocrResultRepo.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					log.Printf("Error closing OCR result repository: %v", err)
				}
			}
		}
	}()
	
	appService := application.NewApplicationService(
		domainService, 
		storageService, 
		fileRepo,
		ocrClient,
		ocrResultRepo,
	)

	port := os.Getenv("GRPC_SERVER_PORT")
	if port == "" {
		port = defaultPort
	}
	port = ":" + port

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
