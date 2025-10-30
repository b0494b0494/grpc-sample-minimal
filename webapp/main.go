package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

const (
	grpcAddress = "server:50051"
	webPort     = ":8080"
	authToken = "my-secret-token"
)

var chatClients = make(map[chan string]bool)
var chatClientsMutex = &sync.Mutex{}

func main() {
	// Serve static React files
	http.Handle("/", http.FileServer(http.Dir("webapp/build")))

	// API endpoints for gRPC calls
	http.HandleFunc("/api/greet", greetHandler)
	http.HandleFunc("/api/stream-counter", streamCounterHandler)
	http.HandleFunc("/api/chat-stream", chatStreamHandler)
	http.HandleFunc("/api/send-chat", sendChatHandler)
	http.HandleFunc("/api/upload-file", uploadFileHandler)
	http.HandleFunc("/api/download-file", downloadFileHandler)

	log.Printf("Web server listening on port %s", webPort)
	log.Fatal(http.ListenAndServe(webPort, nil))
}

func getGrpcClient(ctx context.Context) (pb.GreeterClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("did not connect to gRPC server: %w", err)
	}
	c := pb.NewGreeterClient(conn)
	return c, conn, nil
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	} 
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := req.Name
	if name == "" {
		name = "world"
	}

	c, conn, err := getGrpcClient(r.Context())
	if err != nil {
		log.Printf("Error getting gRPC client: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	reply, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		http.Error(w, "Failed to get greeting from gRPC server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"greeting": reply.GetMessage()})
}

func streamCounterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	c, conn, err := getGrpcClient(r.Context())
	if err != nil {
		log.Printf("Error getting gRPC client: %v", err)
		fmt.Fprintf(w, "event: error\ndata: Failed to connect to gRPC server\n\n")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second) // Increased timeout for streaming
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	stream, err := c.StreamCounter(ctx, &pb.CounterRequest{Limit: 10})
	if err != nil {
		log.Printf("could not call StreamCounter: %v", err)
		fmt.Fprintf(w, "event: error\ndata: Failed to start counter stream\n\n")
		return
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			fmt.Fprintf(w, "event: end\ndata: Counter stream finished\n\n")
			break
		}
		if err != nil {
			log.Printf("error while receiving stream: %v", err)
			fmt.Fprintf(w, "event: error\ndata: Error receiving counter: %v\n\n", err)
			break
		}
		fmt.Fprintf(w, "event: message\ndata: %d\n\n", reply.GetCount())
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}

func chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string)

	chatClientsMutex.Lock()
	chatClients[messageChan] = true
	chatClientsMutex.Unlock()

	defer func() {
		chatClientsMutex.Lock()
		delete(chatClients, messageChan)
		chatClientsMutex.Unlock()
		close(messageChan)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func sendChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		User string `json:"user"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := req.User
	message := req.Message

	if user == "" || message == "" {
		http.Error(w, "User and message cannot be empty", http.StatusBadRequest)
		return
	}

	c, conn, err := getGrpcClient(r.Context())
	if err != nil {
		log.Printf("Error getting gRPC client: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	stream, err := c.Chat(ctx)
	if err != nil {
		log.Printf("could not open chat stream: %v", err)
		http.Error(w, "Failed to open chat stream", http.StatusInternalServerError)
		return
	}

	// Send message to gRPC server
	if err := stream.Send(&pb.ChatMessage{User: user, Message: message}); err != nil {
		log.Printf("failed to send chat message to gRPC server: %v", err)
		http.Error(w, "Failed to send chat message", http.StatusInternalServerError)
		return
	}

	// Receive echo from gRPC server and broadcast to web clients
	reply, err := stream.Recv()
	if err != nil {
		log.Printf("failed to receive chat message from gRPC server: %v", err)
		http.Error(w, "Failed to receive chat message", http.StatusInternalServerError)
		return
	}

	fullMessage := fmt.Sprintf("%s: %s (echoed by server)", reply.GetUser(), reply.GetMessage())
	chatClientsMutex.Lock()
	for clientChan := range chatClients {
		clientChan <- fullMessage
	}
	chatClientsMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	c, conn, err := getGrpcClient(r.Context())
	if err != nil {
		log.Printf("Error getting gRPC client: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Long timeout for large files
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	stream, err := c.UploadFile(ctx)
	if err != nil {
		log.Printf("could not open upload stream: %v", err)
		http.Error(w, "Failed to open upload stream", http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
	
		if err := stream.Send(&pb.FileChunk{Content: buf[:n], Filename: header.Filename, Filesize: header.Size}); err != nil {
			http.Error(w, "Error sending file chunk", http.StatusInternalServerError)
			return
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("failed to receive upload status: %v", err)
		http.Error(w, "Failed to get upload status", http.StatusInternalServerError)
		return
	}

	if !reply.GetSuccess() {
		http.Error(w, reply.GetMessage(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": reply.GetMessage(), "filename": reply.GetFilename(), "bytesWritten": fmt.Sprintf("%d", reply.GetBytesWritten())})
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	c, conn, err := getGrpcClient(r.Context())
	if err != nil {
		log.Printf("Error getting gRPC client: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Long timeout for large files
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

	stream, err := c.DownloadFile(ctx, &pb.FileDownloadRequest{Filename: filename})
	if err != nil {
		log.Printf("could not open download stream: %v", err)
		http.Error(w, "Failed to open download stream", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error while receiving file chunk: %v", err)
			http.Error(w, "Error receiving file chunk", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(chunk.GetContent()); err != nil {
			log.Printf("error writing file chunk to http response: %v", err)
			return
		}
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
	w.WriteHeader(http.StatusOK)
}
