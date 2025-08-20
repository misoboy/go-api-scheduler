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

// Scheduler represents a single scheduler instance.
type Scheduler struct {
	id       string
	stopChan chan struct{}
	running  bool
	config   SchedulerConfig
}

var (
	// schedulers stores active scheduler instances by their ID.
	schedulers map[string]*Scheduler
	// mu protects concurrent access to the schedulers map.
	mu sync.Mutex
)

// Init initializes the scheduler package.
func Init() {
	schedulers = make(map[string]*Scheduler)
}

// StartScheduler starts a new scheduler instance.
func StartScheduler(id string, config SchedulerConfig) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := schedulers[id]; ok {
		logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 이미 실행 중입니다. 새로운 요청을 무시합니다.", id))
		return
	}

	s := &Scheduler{
		id:       id,
		stopChan: make(chan struct{}),
		running:  true,
		config:   config,
	}
	schedulers[id] = s
	go s.run()
}

// StopScheduler stops a scheduler instance by its ID.
func StopScheduler(id string) {
	mu.Lock()
	defer mu.Unlock()

	if s, ok := schedulers[id]; ok {
		if s.running {
			close(s.stopChan)
			s.running = false
			delete(schedulers, id)
			logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 중지되었습니다.", id))
		} else {
			logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 실행 중이지 않습니다.", id))
		}
	} else {
		logger.AddLog(fmt.Sprintf("[%s] 존재하지 않는 스케줄러 ID입니다.", id))
	}
}

// run is a goroutine that handles the scheduling and API calls for a single scheduler.
func (s *Scheduler) run() {
	logger.AddLog(fmt.Sprintf("[%s] 스케줄러 시작 요청을 받았습니다.", s.id))
	logger.AddLog(fmt.Sprintf("[%s] 설정: 시작 시각 %s, 반복 %d%s, URL %s", s.id, s.config.StartTime, s.config.RepeatValue, s.config.RepeatUnit, s.config.APIURL))

	loc := time.Local
	startTimeStr := fmt.Sprintf("%s %s", time.Now().Format("2006-01-02"), s.config.StartTime)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	if err != nil {
		logger.AddLog(fmt.Sprintf("[%s] 시작 시간 파싱 오류: %v", s.id, err))
		StopScheduler(s.id)
		return
	}

	now := time.Now()
	if startTime.Before(now) {
		startTime = startTime.Add(24 * time.Hour)
	}
	waitDuration := startTime.Sub(now)

	logger.AddLog(fmt.Sprintf("[%s] 스케줄 시작까지 대기 중입니다... 남은 시간: %s", s.id, waitDuration))

	select {
	case <-time.After(waitDuration):
		// Start time has been reached. Continue.
	case <-s.stopChan:
		logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 시작 전에 중지되었습니다.", s.id))
		return
	}

	var repeatInterval time.Duration
	switch s.config.RepeatUnit {
	case "h":
		repeatInterval = time.Duration(s.config.RepeatValue) * time.Hour
	case "m":
		repeatInterval = time.Duration(s.config.RepeatValue) * time.Minute
	case "s":
		repeatInterval = time.Duration(s.config.RepeatValue) * time.Second
	default:
		logger.AddLog(fmt.Sprintf("[%s] 유효하지 않은 반복 단위입니다. 스케줄러를 중지합니다.", s.id))
		StopScheduler(s.id)
		return
	}

	ticker := time.NewTicker(repeatInterval)
	defer ticker.Stop()

	logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 실행 중입니다.", s.id))

	for {
		select {
		case <-ticker.C:
			s.callAPI()
		case <-s.stopChan:
			logger.AddLog(fmt.Sprintf("[%s] 스케줄러가 중지되었습니다.", s.id))
			return
		}
	}
}

// callAPI makes the HTTP request based on the scheduler's configuration.
func (s *Scheduler) callAPI() {
	logger.AddLog(fmt.Sprintf("[%s] API 호출 시작: URL %s, 메서드 %s", s.id, s.config.APIURL, s.config.HTTPMethod))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var req *http.Request
	var err error

	var payload map[string]string
	json.Unmarshal([]byte(s.config.Payload), &payload)

	if strings.ToUpper(s.config.HTTPMethod) == "POST" {
		form := url.Values{}
		for key, value := range payload {
			form.Add(key, value)
		}
		req, err = http.NewRequest("POST", s.config.APIURL, strings.NewReader(form.Encode()))
		if err != nil {
			logger.AddLog(fmt.Sprintf("[%s] 요청 생성 오류: %v", s.id, err))
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		baseURL, err := url.Parse(s.config.APIURL)
		if err != nil {
			logger.AddLog(fmt.Sprintf("[%s] URL 파싱 오류: %v", s.id, err))
			return
		}
		params := url.Values{}
		for key, value := range payload {
			params.Add(key, value)
		}
		baseURL.RawQuery = params.Encode()
		req, err = http.NewRequest("GET", baseURL.String(), nil)
		if err != nil {
			logger.AddLog(fmt.Sprintf("[%s] 요청 생성 오류: %v", s.id, err))
			return
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.AddLog(fmt.Sprintf("[%s] API 호출 오류: %v", s.id, err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.AddLog(fmt.Sprintf("[%s] 응답 본문 읽기 오류: %v", s.id, err))
		return
	}

	logger.AddLog(fmt.Sprintf("[%s] API 호출 성공 - HTTP 상태 코드: %d", s.id, resp.StatusCode))
	logger.AddLog(fmt.Sprintf("[%s] 응답 본문: %s", s.id, string(body)))

	if resp.StatusCode == http.StatusOK {
		logger.AddLog(fmt.Sprintf("[%s] 응답 성공 (200 OK) - 스케줄러가 자동으로 중지됩니다.", s.id))
		StopScheduler(s.id)
	}
}
