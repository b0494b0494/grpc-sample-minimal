package handlers

import (
	"context"
	"fmt"
	"io"
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
		WriteJSONError(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

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

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			WriteJSONError(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		if err := stream.Send(&pb.FileChunk{Content: buf[:n], Filename: header.Filename, Filesize: header.Size}); err != nil {
			WriteJSONError(w, "Error sending file chunk", http.StatusInternalServerError)
			return
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		WriteJSONError(w, "Failed to get upload status", http.StatusInternalServerError)
		return
	}
	if !reply.GetSuccess() {
		WriteJSONError(w, reply.GetMessage(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	WriteJSON(w, map[string]string{
		"message": reply.GetMessage(), "filename": reply.GetFilename(), "bytesWritten": fmt.Sprintf("%d", reply.GetBytesWritten()),
	})
}
