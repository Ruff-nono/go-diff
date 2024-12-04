package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logger  *log.Logger
	logFile = "diff.log"
)

func init() {
	file, err := os.Create(filepath.Clean(logFile))
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}

	logger = log.New(file, "time: ", log.LstdFlags)
	logger.Printf("start at %s", time.Now().Format(time.DateTime))
}
