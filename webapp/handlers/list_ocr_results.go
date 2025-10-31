package handlers

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func ListOCRResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	storageProvider := r.URL.Query().Get("storageProvider")
	if storageProvider == "" {
		storageProvider = "azure" // Default for Phase 1
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

	ocrReq := &pb.OCRListRequest{
		StorageProvider: storageProvider,
	}

	resp, err := client.ListOCRResults(ctx, ocrReq)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, resp)
}
