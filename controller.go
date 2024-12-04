package main

import (
	"errors"
	"net/http"
	"sync"
)

type ProxyController struct {
	mu      sync.Mutex
	running bool
	server  *http.Server
}

var proxyController = &ProxyController{}

func startProxyAsync() {
	proxyController.mu.Lock()
	defer proxyController.mu.Unlock()

	if proxyController.running {
		logger.Println("Proxy is already running...")
		return
	}

	proxyController.server = &http.Server{Addr: config.SelfPort}

	proxy1 := NewProxy(config.Host1)
	proxy2 := NewProxy(config.Host2)
	http.HandleFunc("/", HandleRequestAndRedirect(proxy1, proxy2))

	go func() {
		proxyController.running = true
		logger.Printf("Starting proxy server, forwarding requests to: %s and %s\n", config.Host1, config.Host2)
		err := proxyController.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Proxy server error: %v\n", err)
		}
		proxyController.mu.Lock()
		proxyController.running = false
		proxyController.mu.Unlock()
		logger.Println("Proxy server stopped.")
	}()
}
