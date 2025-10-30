package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func GreetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := req.Name
	if name == "" {
		name = "world"
	}

	c, conn, err := GetGrpcClient(r.Context())
	if err != nil {
		WriteJSONError(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	reply, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		WriteJSONError(w, "Failed to get greeting from gRPC server", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"greeting": reply.GetMessage()})
}
