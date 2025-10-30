package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

const (
	address     = "server:50051"
	defaultName = "world"
	authToken = "my-secret-token"
)

// loggingClientInterceptor is a unary interceptor that logs RPC calls.
func loggingClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Printf("Outgoing RPC: %s, Request: %v", method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("Incoming RPC: %s, Response: %v, Error: %v", method, reply, err)
	return err
}

// loggingClientStreamInterceptor is a stream interceptor that logs RPC calls.
func loggingClientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Printf("Outgoing stream RPC: %s", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	log.Printf("Incoming stream RPC: %s, Error: %v", method, err)
	return s, err
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(loggingClientInterceptor),
		grpc.WithChainStreamInterceptor(loggingClientStreamInterceptor),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Add auth token to context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	// Call StreamCounter
	log.Printf("Calling StreamCounter with limit 5")
	stream, err := c.StreamCounter(ctx, &pb.CounterRequest{Limit: 5})
	if err != nil {
		log.Fatalf("could not call StreamCounter: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while receiving stream: %v", err)
		}
		log.Printf("Stream Counter: %d", reply.GetCount())
	}
	log.Printf("StreamCounter finished")

	// Call Chat (bidirectional streaming)
	log.Printf("Calling Chat (bidirectional streaming)")
	chatStream, err := c.Chat(ctx)
	if err != nil {
		log.Fatalf("could not open chat stream: %v", err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := chatStream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a chat message: %v", err)
			}
			log.Printf("Received chat message from %s: %s", in.GetUser(), in.GetMessage())
		}
	}()

	for i := 0; i < 3; i++ {
		msg := &pb.ChatMessage{User: name, Message: fmt.Sprintf("Hello from client %d", i)}
		if err := chatStream.Send(msg); err != nil {
			log.Fatalf("Failed to send chat message: %v", err)
		}
		time.Sleep(time.Second)
	}
	chatStream.CloseSend()
	<-waitc
	log.Printf("Chat finished")

	// File Upload
	log.Printf("Uploading file: test.txt")
	uploadFile(ctx, c, "test.txt", []byte("This is a test file content."))

	// File Download
	log.Printf("Downloading file: test.txt")
	downloadFile(ctx, c, "test.txt")
}

func uploadFile(ctx context.Context, c pb.GreeterClient, filename string, content []byte) {
	stream, err := c.UploadFile(ctx)
	if err != nil {
		log.Fatalf("could not open upload stream: %v", err)
	}

	fileSize := int64(len(content))
	chunkSize := 1024 // 1KB

	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]

		if err := stream.Send(&pb.FileChunk{Content: chunk, Filename: filename, Filesize: fileSize}); err != nil {
			log.Fatalf("failed to send file chunk: %v", err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive upload status: %v", err)
	}
	log.Printf("File Upload Status: %v", reply)
}

func downloadFile(ctx context.Context, c pb.GreeterClient, filename string) {
	stream, err := c.DownloadFile(ctx, &pb.FileDownloadRequest{Filename: filename})
	if err != nil {
		log.Fatalf("could not open download stream: %v", err)
	}

	var downloadedContent []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while receiving file chunk: %v", err)
		}
		downloadedContent = append(downloadedContent, chunk.GetContent()...)
	}
	log.Printf("File Downloaded: %s, Size: %d, Content: %s", filename, len(downloadedContent), string(downloadedContent))
}
