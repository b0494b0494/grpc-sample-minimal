package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func ProcessOCRHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename        string `json:"filename"`
		StorageProvider string `json:"storage_provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		WriteJSONError(w, "filename is required", http.StatusBadRequest)
		return
	}

	if req.StorageProvider == "" {
		req.StorageProvider = "azure" // Default for Phase 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, conn, err := GetGrpcClient(ctx)
	if err != nil {
		WriteJSONError(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Add auth token to metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	ocrReq := &pb.OCRRequest{
		Filename:        req.Filename,
		StorageProvider: req.StorageProvider,
	}

	resp, err := client.ProcessOCR(ctx, ocrReq)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, resp)
}
