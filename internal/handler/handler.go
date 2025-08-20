// internal/handler/handler.go
package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"go-api-scheduler/internal/logger"
	"go-api-scheduler/internal/scheduler"
)

// Config holds the user's scheduler configuration.
type Config struct {
	ID          string `json:"id"`
	StartTime   string `json:"startTime"`
	RepeatValue int    `json:"repeatValue"`
	RepeatUnit  string `json:"repeatUnit"`
	APIURL      string `json:"apiURL"`
	HTTPMethod  string `json:"httpMethod"`
	Payload     string `json:"payload"`
}

// Init initializes the handler package.
func Init() {
	// Empty for now, but good practice for future initialization.
}

// StartHandler handles the request to start a scheduler.
func StartHandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, "잘못된 요청 본문입니다.", http.StatusBadRequest)
		return
	}

	scheduler.StartScheduler(config.ID, scheduler.SchedulerConfig{
		StartTime:   config.StartTime,
		RepeatValue: config.RepeatValue,
		RepeatUnit:  config.RepeatUnit,
		APIURL:      config.APIURL,
		HTTPMethod:  config.HTTPMethod,
		Payload:     config.Payload,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("스케줄러가 시작되었습니다."))
}

// StopHandler handles the request to stop a scheduler.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "잘못된 요청 본문입니다.", http.StatusBadRequest)
		return
	}

	id := reqBody["id"]
	scheduler.StopScheduler(id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("스케줄러가 중지되었습니다."))
}

// LogsHandler returns the current log entries.
func LogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logger.GetLogs())
}

// FakeServerHandler handles the request for the fake server.
func FakeServerHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "요청 본문을 읽을 수 없습니다.", http.StatusInternalServerError)
		return
	}

	// Set the content type to JSON.
	w.Header().Set("Content-Type", "application/json")

	// Write the received body back to the response.
	w.Write(body)
}
