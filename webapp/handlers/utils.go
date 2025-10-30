package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-sample-minimal/proto"
)

const (
	grpcAddress = "server:50051"
)

var (
	authToken = os.Getenv("AUTH_TOKEN")
)

func GetAuthToken() string { return authToken }

func GetGrpcClient(ctx context.Context) (pb.GreeterClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("did not connect to gRPC server: %w", err)
	}
	c := pb.NewGreeterClient(conn)
	return c, conn, nil
}

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func WriteSSEError(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
    fmt.Fprintf(w, "event: error\ndata: %s\n\n", message)
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func WriteJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(v)
}

// Chat broadcast helpers
var chatClients = make(map[chan string]bool)
var chatClientsMutex = &sync.Mutex{}

func RegisterChatClient(ch chan string) {
    chatClientsMutex.Lock()
    chatClients[ch] = true
    chatClientsMutex.Unlock()
}

func UnregisterChatClient(ch chan string) {
    chatClientsMutex.Lock()
    delete(chatClients, ch)
    chatClientsMutex.Unlock()
}

func BroadcastChat(msg string) {
    chatClientsMutex.Lock()
    for c := range chatClients {
        c <- msg
    }
    chatClientsMutex.Unlock()
}
