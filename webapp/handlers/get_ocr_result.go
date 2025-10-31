package handlers

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func GetOCRResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	storageProvider := r.URL.Query().Get("storageProvider")
	engineName := r.URL.Query().Get("engineName")

	if filename == "" {
		WriteJSONError(w, "filename is required", http.StatusBadRequest)
		return
	}

	if storageProvider == "" {
		storageProvider = "azure" // Default for Phase 1
	}

	if engineName == "" {
		engineName = "tesseract" // Default
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

	ocrReq := &pb.OCRResultRequest{
		Filename:        filename,
		StorageProvider: storageProvider,
		EngineName:      engineName,
	}

	resp, err := client.GetOCRResult(ctx, ocrReq)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, resp)
}
