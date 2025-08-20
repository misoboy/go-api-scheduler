// internal/logger/logger.go
package logger

import (
	"sync"
	"time"
)

// LogEntry represents a single log message.
type LogEntry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
}

var (
	// logs stores the console output.
	logs []LogEntry
	// mu protects concurrent access to the logs.
	mu sync.Mutex
)

// Init initializes the logger.
func Init() {
	// You can add initialization logic here if needed.
}

// AddLog adds a new log message to the log list.
func AddLog(message string) {
	mu.Lock()
	defer mu.Unlock()
	entry := LogEntry{
		Time:    time.Now().Format("15:04:05"),
		Message: message,
	}
	logs = append(logs, entry)
	// Keep the log list from growing too large.
	if len(logs) > 100 {
		logs = logs[1:]
	}
}

// GetLogs returns the current log entries.
func GetLogs() []LogEntry {
	mu.Lock()
	defer mu.Unlock()
	return logs
}
