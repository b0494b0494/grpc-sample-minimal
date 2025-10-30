package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	provider := r.URL.Query().Get("storageProvider")
	if filename == "" {
		WriteJSONError(w, "Filename is required", http.StatusBadRequest)
		return
	}

	c, conn, err := GetGrpcClient(r.Context())
	if err != nil {
		WriteJSONError(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	stream, err := c.DownloadFile(ctx, &pb.FileDownloadRequest{Filename: filename, StorageProvider: provider})
	if err != nil {
		WriteJSONError(w, "Failed to open download stream", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if _, err := w.Write(chunk.GetContent()); err != nil {
			return
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	w.WriteHeader(http.StatusOK)
}
