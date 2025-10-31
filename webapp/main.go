package main

import (
    "log"
    "net/http"
    "path/filepath"
    "strings"

    "grpc-sample-minimal/webapp/handlers"
)

const (
    webPort = ":8080"
)

func main() {
    fs := http.FileServer(http.Dir("webapp/build"))
    http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ext := strings.ToLower(filepath.Ext(r.URL.Path))
        switch ext {
        case ".html":
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
        case ".js":
            w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
        case ".css":
            w.Header().Set("Content-Type", "text/css; charset=utf-8")
        case ".json":
            w.Header().Set("Content-Type", "application/json; charset=utf-8")
        case ".txt", ".svg":
            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        case "":
            // SPA routes without extension
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
        }
        fs.ServeHTTP(w, r)
    }))

    http.HandleFunc("/api/greet", handlers.GreetHandler)
    http.HandleFunc("/api/stream-counter", handlers.StreamCounterHandler)
    http.HandleFunc("/api/chat-stream", handlers.ChatStreamHandler)
    http.HandleFunc("/api/send-chat", handlers.SendChatHandler)
    http.HandleFunc("/api/upload-file", handlers.UploadFileHandler)
    http.HandleFunc("/api/download-file", handlers.DownloadFileHandler)
	http.HandleFunc("/api/list-files", handlers.ListFilesHandler)
	http.HandleFunc("/api/delete-file", handlers.DeleteFileHandler)

    log.Printf("Web server listening on port %s", webPort)
    log.Fatal(http.ListenAndServe(webPort, nil))
}
