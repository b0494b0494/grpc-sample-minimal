package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	provider := r.FormValue("storageProvider")
	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Error retrieving the file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate filename
	if header.Filename == "" {
		WriteJSONError(w, "Please enter a filename", http.StatusBadRequest)
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
	if provider != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "storage-provider", provider)
	}

	stream, err := c.UploadFile(ctx)
	if err != nil {
		WriteJSONError(w, "Failed to open upload stream", http.StatusInternalServerError)
		return
	}

	// Log filename for debugging
	log.Printf("Uploading file: filename=%s, size=%d, provider=%s", header.Filename, header.Size, provider)
	
	filename := header.Filename
	if filename == "" {
		WriteJSONError(w, "File has no filename", http.StatusBadRequest)
		return
	}
	
	// For empty files, send at least one chunk with filename to ensure it's received
	if header.Size == 0 {
		log.Printf("Empty file detected, sending filename-only chunk")
		if err := stream.Send(&pb.FileChunk{
			Content:  []byte{},
			Filename: filename,
			Filesize: 0,
		}); err != nil {
			WriteJSONError(w, fmt.Sprintf("Error sending empty file chunk: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Sent empty file chunk for: %s", filename)
	} else {
		// For non-empty files, read and send chunks
		buf := make([]byte, 1024)
		chunkCount := 0
		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				WriteJSONError(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
				return
			}
			// Always send filename and filesize in every chunk (required for proper handling)
			if err := stream.Send(&pb.FileChunk{
				Content:  buf[:n],
				Filename: filename,
				Filesize: header.Size,
			}); err != nil {
				WriteJSONError(w, fmt.Sprintf("Error sending file chunk: %v", err), http.StatusInternalServerError)
				return
			}
			chunkCount++
		}
		log.Printf("Sent %d chunks for file: %s (size: %d)", chunkCount, filename, header.Size)
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Failed to get upload status: %v", err), http.StatusInternalServerError)
		return
	}
	if !reply.GetSuccess() {
		WriteJSONError(w, reply.GetMessage(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	WriteJSON(w, map[string]interface{}{
		"message":        reply.GetMessage(),
		"filename":      reply.GetFilename(),
		"bytesWritten":   fmt.Sprintf("%d", reply.GetBytesWritten()),
		"success":        reply.GetSuccess(),
		"storageProvider": reply.GetStorageProvider(),
	})
}
