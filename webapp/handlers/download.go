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
	preview := r.URL.Query().Get("preview") == "true" // Check if this is a preview request
	
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

	// Set appropriate Content-Type based on file extension
	contentType := "application/octet-stream"
	if preview {
		// Determine content type for preview
		if len(filename) > 4 {
			ext := filename[len(filename)-4:]
			switch ext {
			case ".pdf":
				contentType = "application/pdf"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".gif":
				contentType = "image/gif"
			case ".txt":
				contentType = "text/plain"
			case ".json":
				contentType = "application/json"
			case ".xml":
				contentType = "application/xml"
			case ".html":
				contentType = "text/html"
			}
		}
	}
	
	w.Header().Set("Content-Type", contentType)
	
	// Set Content-Disposition: inline for preview, attachment for download
	if preview {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

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
