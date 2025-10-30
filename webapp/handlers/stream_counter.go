package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	pb "grpc-sample-minimal/proto"
)

func StreamCounterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	c, conn, err := GetGrpcClient(r.Context())
	if err != nil {
		WriteSSEError(w, "Failed to connect to gRPC server")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", GetAuthToken())

	stream, err := c.StreamCounter(ctx, &pb.CounterRequest{Limit: 10})
	if err != nil {
		WriteSSEError(w, "Failed to start counter stream")
		return
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Fprintf(w, "event: end\ndata: Counter stream finished\n\n")
				break
			}
			WriteSSEError(w, fmt.Sprintf("Error receiving counter: %v", err))
			break
		}
		fmt.Fprintf(w, "event: message\ndata: %d\n\n", reply.GetCount())
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}
