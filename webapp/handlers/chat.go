package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func ChatStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string)
	RegisterChatClient(messageChan)
	defer UnregisterChatClient(messageChan)

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteSSEError(w, "Streaming unsupported!")
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

func SendChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ User, Message string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.User == "" || req.Message == "" {
		WriteJSONError(w, "User and message cannot be empty", http.StatusBadRequest)
		return
	}

	c, conn, err := GetGrpcClient(r.Context())
	if err != nil {
		WriteJSONError(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	stream, err := c.Chat(ctx)
	if err != nil {
		WriteJSONError(w, "Failed to open chat stream", http.StatusInternalServerError)
		return
	}
	if err := stream.Send(&pb.ChatMessage{User: req.User, Message: req.Message}); err != nil {
		WriteJSONError(w, "Failed to send chat message", http.StatusInternalServerError)
		return
	}
	reply, err := stream.Recv()
	if err != nil {
		WriteJSONError(w, "Failed to receive chat message", http.StatusInternalServerError)
		return
	}
	BroadcastChat(fmt.Sprintf("%s: %s (echoed by server)", reply.GetUser(), reply.GetMessage()))
	w.WriteHeader(http.StatusOK)
}
