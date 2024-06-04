package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	logger  *log.Logger
	logFile = "diff.log"
)

func initLog() {
	file, err := os.Create(filepath.Clean(logFile))
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}

	logger = log.New(file, "time: ", log.LstdFlags)
	logger.Printf("start at %s", time.Now().Format(time.DateTime))
}

func main() {
	initLog()

	proxy1 := NewProxy(config.Host1)
	proxy2 := NewProxy(config.Host2)
	http.HandleFunc("/", HandleRequestAndRedirect(proxy1, proxy2))
	logger.Printf("Starting proxy server, forwarding requests to: %s and %s\n", config.Host1, config.Host2)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
