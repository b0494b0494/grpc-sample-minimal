package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-sample-minimal/proto"
)

const (
	grpcAddress = "server:50051"
	webPort     = ":8080"
)

type PageData struct {
	Greeting string
	Error    string
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/greet", greetHandler)
	http.HandleFunc("/stream-counter", streamCounterHandler)
	http.HandleFunc("/chat-stream", chatStreamHandler)
	http.HandleFunc("/send-chat", sendChatHandler)

	log.Printf("Web server listening on port %s", webPort)
	log.Fatal(http.ListenAndServe(webPort, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("webapp/index.html"))
	tmpl.Execute(w, PageData{})
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = "world"
	}

	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect to gRPC server: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		http.Error(w, "Failed to get greeting from gRPC server", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("webapp/index.html"))
	tmpl.Execute(w, PageData{Greeting: reply.GetMessage()})
}

func streamCounterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect to gRPC server: %v", err)
		fmt.Fprintf(w, "event: error\ndata: Failed to connect to gRPC server\n\n")
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout for streaming
	defer cancel()

	stream, err := c.StreamCounter(ctx, &pb.CounterRequest{Limit: 10})
	if err != nil {
		log.Printf("could not call StreamCounter: %v", err)
		fmt.Fprintf(w, "event: error\ndata: Failed to start counter stream\n\n")
		return
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			fmt.Fprintf(w, "event: end\ndata: Counter stream finished\n\n")
			break
		}
		if err != nil {
			log.Printf("error while receiving stream: %v", err)
			fmt.Fprintf(w, "event: error\ndata: Error receiving counter: %v\n\n", err)
			break
		}
		fmt.Fprintf(w, "event: message\ndata: %d\n\n", reply.GetCount())
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}

var chatClients = make(map[chan string]bool)
var chatClientsMutex = &sync.Mutex{}

func chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string)

	chatClientsMutex.Lock()
	chatClients[messageChan] = true
	chatClientsMutex.Unlock()

	defer func() {
		chatClientsMutex.Lock()
		delete(chatClients, messageChan)
		chatClientsMutex.Unlock()
		close(messageChan)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func sendChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user := r.FormValue("user")
	message := r.FormValue("message")

	if user == "" || message == "" {
		http.Error(w, "User and message cannot be empty", http.StatusBadRequest)
		return
	}

	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect to gRPC server: %v", err)
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := c.Chat(ctx)
	if err != nil {
		log.Printf("could not open chat stream: %v", err)
		http.Error(w, "Failed to open chat stream", http.StatusInternalServerError)
		return
	}

	// Send message to gRPC server
	if err := stream.Send(&pb.ChatMessage{User: user, Message: message}); err != nil {
		log.Printf("failed to send chat message to gRPC server: %v", err)
		http.Error(w, "Failed to send chat message", http.StatusInternalServerError)
		return
	}

	// Receive echo from gRPC server and broadcast to web clients
	reply, err := stream.Recv()
	if err != nil {
		log.Printf("failed to receive chat message from gRPC server: %v", err)
		http.Error(w, "Failed to receive chat message", http.StatusInternalServerError)
		return
	}

	fullMessage := fmt.Sprintf("%s: %s (echoed by server)", reply.GetUser(), reply.GetMessage())
	chatClientsMutex.Lock()
	for clientChan := range chatClients {
		clientChan <- fullMessage
	}
	chatClientsMutex.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}