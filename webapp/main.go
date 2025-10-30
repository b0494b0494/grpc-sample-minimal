package main

import (
    "log"
    "net/http"

    "grpc-sample-minimal/webapp/handlers"
)

const (
    webPort = ":8080"
)

func main() {
    http.Handle("/", http.FileServer(http.Dir("webapp/build")))

    http.HandleFunc("/api/greet", handlers.GreetHandler)
    http.HandleFunc("/api/stream-counter", handlers.StreamCounterHandler)
    http.HandleFunc("/api/chat-stream", handlers.ChatStreamHandler)
    http.HandleFunc("/api/send-chat", handlers.SendChatHandler)
    http.HandleFunc("/api/upload-file", handlers.UploadFileHandler)
    http.HandleFunc("/api/download-file", handlers.DownloadFileHandler)

    log.Printf("Web server listening on port %s", webPort)
    log.Fatal(http.ListenAndServe(webPort, nil))
}
