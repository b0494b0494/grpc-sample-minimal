package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	provider := r.URL.Query().Get("storageProvider")
	if filename == "" {
		WriteJSONError(w, "Filename is required", http.StatusBadRequest)
		return
	}
	if provider == "" {
		provider = "s3"
	}

	c, conn, err := GetGrpcClient(r.Context())
	if err != nil {
		WriteJSONError(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	req := &pb.DeleteFileRequest{
		Filename:       filename,
		StorageProvider: provider,
	}

	resp, err := c.DeleteFile(ctx, req)
	if err != nil {
		WriteJSONError(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
