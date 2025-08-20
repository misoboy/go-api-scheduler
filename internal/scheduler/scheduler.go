// internal/scheduler/scheduler.go
package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-api-scheduler/internal/logger"
)

// SchedulerConfig holds the user's scheduler configuration.
type SchedulerConfig struct {
	StartTime   string `json:"startTime"`
	RepeatValue int    `json:"repeatValue"`
	RepeatUnit  string `json:"repeatUnit"`
	APIURL      string `json:"apiURL"`
	HTTPMethod  string `json:"httpMethod"`
	Payload     string `json:"payload"`
}

var (
	// schedulerStopChan is used to signal the scheduler to stop.
	schedulerStopChan chan struct{}
	// schedulerRunning indicates if the scheduler is currently active.
	schedulerRunning bool
	// mu protects concurrent access to the state variables.
	mu sync.Mutex
)

// Init initializes the scheduler package.
func Init() {
	// Empty for now, but good practice for future initialization.
}

// StartScheduler starts the scheduler goroutine.
func StartScheduler(config SchedulerConfig) {
	mu.Lock()
	if schedulerRunning {
		mu.Unlock()
		logger.AddLog("스케줄러가 이미 실행 중입니다. 새로운 요청을 무시합니다.")
		return
	}
	schedulerRunning = true
	schedulerStopChan = make(chan struct{})
	mu.Unlock()
	go runScheduler(config)
}

// StopScheduler stops the scheduler.
func StopScheduler() {
	mu.Lock()
	defer mu.Unlock()
	if !schedulerRunning {
		logger.AddLog("스케줄러가 실행 중이지 않습니다.")
		return
	}
	close(schedulerStopChan)
	schedulerRunning = false
}

// runScheduler is a goroutine that handles the scheduling and API calls.
func runScheduler(config SchedulerConfig) {
	logger.AddLog("스케줄러 시작 요청을 받았습니다.")
	logger.AddLog(fmt.Sprintf("설정: 시작 시각 %s, 반복 %d%s, URL %s", config.StartTime, config.RepeatValue, config.RepeatUnit, config.APIURL))

	// Get the local time zone.
	loc := time.Local

	// Parse the start time using the local time zone.
	startTimeStr := fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), config.StartTime)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	if err != nil {
		logger.AddLog(fmt.Sprintf("시작 시간 파싱 오류: %v", err))
		StopScheduler()
		return
	}

	// Calculate the duration until the start time.
	now := time.Now()
	if startTime.Before(now) {
		startTime = startTime.Add(24 * time.Hour)
	}
	waitDuration := startTime.Sub(now)

	logger.AddLog(fmt.Sprintf("스케줄 시작까지 대기 중입니다... 남은 시간: %s", waitDuration))

	select {
	case <-time.After(waitDuration):
		// Start time has been reached. Continue.
	case <-schedulerStopChan:
		logger.AddLog("스케줄러가 시작 전에 중지되었습니다.")
		return
	}

	// Determine the repeat interval.
	var repeatInterval time.Duration
	switch config.RepeatUnit {
	case "h":
		repeatInterval = time.Duration(config.RepeatValue) * time.Hour
	case "m":
		repeatInterval = time.Duration(config.RepeatValue) * time.Minute
	case "s":
		repeatInterval = time.Duration(config.RepeatValue) * time.Second
	default:
		logger.AddLog("유효하지 않은 반복 단위입니다. 스케줄러를 중지합니다.")
		StopScheduler()
		return
	}

	ticker := time.NewTicker(repeatInterval)
	defer ticker.Stop()

	logger.AddLog("스케줄러가 실행 중입니다.")

	// Start the main loop.
	for {
		select {
		case <-ticker.C:
			// Time to execute the API call.
			callAPI(config)
		case <-schedulerStopChan:
			// Stop signal received.
			logger.AddLog("스케줄러가 중지되었습니다.")
			return
		}
	}
}

// callAPI makes the HTTP request based on the configuration.
func callAPI(config SchedulerConfig) {
	logger.AddLog(fmt.Sprintf("API 호출 시작: URL %s, 메서드 %s", config.APIURL, config.HTTPMethod))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var req *http.Request
	var err error

	// Unmarshal the payload from JSON string to a map
	var payload map[string]string
	json.Unmarshal([]byte(config.Payload), &payload)

	// Handle POST and GET methods based on the payload.
	if strings.ToUpper(config.HTTPMethod) == "POST" {
		form := url.Values{}
		for key, value := range payload {
			form.Add(key, value)
		}
		req, err = http.NewRequest("POST", config.APIURL, strings.NewReader(form.Encode()))
		if err != nil {
			logger.AddLog(fmt.Sprintf("요청 생성 오류: %v", err))
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		// For GET, append payload to the URL as query parameters.
		baseURL, err := url.Parse(config.APIURL)
		if err != nil {
			logger.AddLog(fmt.Sprintf("URL 파싱 오류: %v", err))
			return
		}
		params := url.Values{}
		for key, value := range payload {
			params.Add(key, value)
		}
		baseURL.RawQuery = params.Encode()
		req, err = http.NewRequest("GET", baseURL.String(), nil)
		if err != nil {
			logger.AddLog(fmt.Sprintf("요청 생성 오류: %v", err))
			return
		}
	}

	// Make the request.
	resp, err := client.Do(req)
	if err != nil {
		logger.AddLog(fmt.Sprintf("API 호출 오류: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.AddLog(fmt.Sprintf("응답 본문 읽기 오류: %v", err))
		return
	}

	logger.AddLog(fmt.Sprintf("API 호출 성공 - HTTP 상태 코드: %d", resp.StatusCode))
	logger.AddLog(fmt.Sprintf("응답 본문: %s", string(body)))

	// If the response is successful (200 OK), stop the scheduler.
	if resp.StatusCode == http.StatusOK {
		logger.AddLog("응답 성공 (200 OK) - 스케줄러가 자동으로 중지됩니다.")
		StopScheduler()
	}
}
