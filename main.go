package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
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

func init() {
	file, err := os.Create(filepath.Clean(logFile))
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}

	multiWriter := io.MultiWriter(
		file,      // 写入文件
		os.Stdout, // 同时输出到控制台
	)

	logger = log.New(multiWriter, "time: ", log.LstdFlags)

	logger.Printf("start at %s", time.Now().Format(time.DateTime))
}

func main() {
	go func() {
		// 监控服务配置
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsMux.HandleFunc("/chart", combinedChartHandler)
		metricsMux.HandleFunc("/debug/errors", queryCurlLog)

		logger.Printf("Starting metrics server[%s]\n", ":19090")
		log.Fatal(http.ListenAndServe(":19090", metricsMux))
	}()

	host1, host2 := config.Host1.String(), config.Host2.String()
	proxy1 := NewProxy(host1)
	proxy2 := NewProxy(host2)
	http.HandleFunc("/", HandleRequestAndRedirect(proxy1, proxy2))
	logger.Printf("Starting proxy server[%s], forwarding requests to: %s and %s\n", config.SelfPort.String(), host1, host2)
	err := http.ListenAndServe(config.SelfPort.String(), nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}

}
