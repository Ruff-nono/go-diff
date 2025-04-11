package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ErrorType int

const (
	StatusCodeMismatch ErrorType = iota
	HeaderMismatch
	BodyMismatch
)

func (e ErrorType) String() string {
	switch e {
	case StatusCodeMismatch:
		return "status code mismatch"
	case HeaderMismatch:
		return "header mismatch"
	case BodyMismatch:
		return "body mismatch"
	default:
		return "unknown"
	}
}

type CompareError struct {
	Path    string // 请求路径
	Type    ErrorType
	Message string
	Details map[string]interface{}
	CurlCmd string // 重放命令
}

func (e *CompareError) Error() string {
	return e.Message
}

func newCompareError(t ErrorType, msg string, details map[string]interface{}) *CompareError {
	return &CompareError{
		Type:    t,
		Message: msg,
		Details: details,
	}
}

type CurlLog struct {
	Path      string    `json:"path"`      // 请求路径
	ErrorType string    `json:"errorType"` // 错误类型
	Timestamp time.Time `json:"timestamp"` // 时间戳
	CurlCmd   string    `json:"curlCmd"`   // 重放命令
	Details   string
}

var (
	errorQueue = make(map[string][]CurlLog)
	logMutex   sync.Mutex
)

func recordCurlLog(r *http.Request, api, errorType string, details string) {
	// 生成CURL命令
	curlCmd := generateCurl(r)

	logMutex.Lock()
	defer logMutex.Unlock()

	key := fmt.Sprintf("%s_%s", api, errorType)

	maxEntries := config.ErrorMaxQueue.Int()
	if _, ok := errorQueue[key]; !ok {
		errorQueue[key] = make([]CurlLog, 0, maxEntries)
	}
	queue := errorQueue[key]
	if len(queue) >= maxEntries {
		queue = queue[1:]
	}
	queue = append(queue, CurlLog{
		Path:      api,
		ErrorType: errorType,
		Timestamp: time.Now().UTC(),
		CurlCmd:   curlCmd,
		Details:   details,
	})

	errorQueue[key] = queue
}

func generateCurl(r *http.Request) string {
	u := *r.URL
	if u.Scheme == "" {
		// 根据TLS状态推断协议
		if r.TLS != nil {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
	}

	cmd := strings.Builder{}
	cmd.WriteString("curl -X " + r.Method)
	cmd.WriteString(" '" + u.String() + "'")

	// 添加Header
	for k, v := range r.Header {
		cmd.WriteString(fmt.Sprintf(" -H '%s: %s'", k, strings.Join(v, ",")))
	}

	// GET方法不需要Body
	if r.Method != http.MethodGet && r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置Body读取位置
		cmd.WriteString(fmt.Sprintf(" --data '%s'", string(body)))
	}

	return cmd.String()
}

func queryCurlLog(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	queryApi := r.URL.Query().Get("api")
	queryState := r.URL.Query().Get("state")

	logMutex.Lock()
	defer logMutex.Unlock()

	key := fmt.Sprintf("%s_%s", queryApi, queryState)

	w.Header().Set("Content-Type", "application/json")
	if curlLogs, ok := errorQueue[key]; ok {
		json.NewEncoder(w).Encode(curlLogs)
		return
	}
	json.NewEncoder(w).Encode("")
}
