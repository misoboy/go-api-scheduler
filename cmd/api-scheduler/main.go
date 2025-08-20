// cmd/api-scheduler/main.go
package main

import (
	"log"
	"net/http"

	"go-api-scheduler/internal/handler"
	"go-api-scheduler/internal/logger"
)

func main() {
	// Initialize the logger.
	logger.Init()
	// Initialize the handler.
	handler.Init()

	// Serve static files from the 'web/static' directory.
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/", fs)

	// Register API endpoints.
	http.HandleFunc("/start", handler.StartHandler)
	http.HandleFunc("/stop", handler.StopHandler)
	http.HandleFunc("/logs", handler.LogsHandler)

	// Add a new endpoint for the fake server.
	http.HandleFunc("/fake-server", handler.FakeServerHandler)

	port := ":8080"
	log.Printf("웹 서버가 http://localhost%s 에서 실행 중입니다.", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
