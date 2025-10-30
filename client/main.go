package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-sample-minimal/proto"
)

const (
	address     = "server:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	// Call StreamCounter
	log.Printf("Calling StreamCounter with limit 5")
	stream, err := c.StreamCounter(ctx, &pb.CounterRequest{Limit: 5})
	if err != nil {
		log.Fatalf("could not call StreamCounter: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while receiving stream: %v", err)
		}
		log.Printf("Stream Counter: %d", reply.GetCount())
	}
	log.Printf("StreamCounter finished")

	// Call Chat (bidirectional streaming)
	log.Printf("Calling Chat (bidirectional streaming)")
	chatStream, err := c.Chat(ctx)
	if err != nil {
		log.Fatalf("could not open chat stream: %v", err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := chatStream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a chat message: %v", err)
			}
			log.Printf("Received chat message from %s: %s", in.GetUser(), in.GetMessage())
		}
	}()

	for i := 0; i < 3; i++ {
		msg := &pb.ChatMessage{User: name, Message: fmt.Sprintf("Hello from client %d", i)}
		if err := chatStream.Send(msg); err != nil {
			log.Fatalf("Failed to send chat message: %v", err)
		}
		time.Sleep(time.Second)
	}
	chatStream.CloseSend()
	<-waitc
	log.Printf("Chat finished")
}
