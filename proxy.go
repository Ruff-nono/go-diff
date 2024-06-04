package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"sync"
)

// NewProxy creates a new reverse proxy to the target host
func NewProxy(targetHost string) *httputil.ReverseProxy {
	target, _ := url.Parse(targetHost)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}

// CompareResponses compares the responses from two servers
func CompareResponses(resp1, resp2 *http.Response) error {
	var wg sync.WaitGroup
	wg.Add(2)
	var body1, body2 []byte
	var err error
	// Forward the request to the first proxy
	go func() {
		defer wg.Done()
		// Read the body of the first response
		body1, err = io.ReadAll(resp1.Body)
		if err != nil {
			logger.Printf("Error reading body from first response: %v", err)
			//return errors.New("reading body from first error")
		}
	}()

	// Forward the request to the second proxy
	go func() {
		defer wg.Done()
		// Read the body of the second response
		body2, err = io.ReadAll(resp2.Body)
		if err != nil {
			logger.Printf("Error reading body from second response: %v", err)
		}
	}()

	wg.Wait() // Wait for both goroutines to finish

	// Compare status code if configured
	if config.CompareStatusCode && !isStatusEquivalent(resp1.StatusCode, resp2.StatusCode) {
		body := fmt.Sprintf("Response bodies.\nFirst response: %s\nSecond response: %s", string(body1), string(body2))
		return errors.New(fmt.Sprintf("Status codes are different. First: %d, Second: %d, %s", resp1.StatusCode, resp2.StatusCode, body))
	}

	// Compare headers
	for _, key := range config.HeadersToCompare {
		val1, ok1 := resp1.Header[key]
		val2, ok2 := resp2.Header[key]
		if ok1 != ok2 || !stringSliceEqual(val1, val2) {
			return errors.New(fmt.Sprintf("Header '%s' values are different. First: %v, Second: %v", key, val1, val2))
		}
	}

	// Compare the bodies
	if config.CompareBody && !bytes.Equal(body1, body2) {
		return errors.New(fmt.Sprintf("Response bodies are different.\nFirst response: %s\nSecond response: %s", string(body1), string(body2)))
	}
	return nil
}

// HandleRequestAndRedirect handles incoming requests and redirects them to two target hosts
func HandleRequestAndRedirect(proxy1, proxy2 *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		req1 := r.Clone(context.Background())
		req2 := r.Clone(context.Background())

		defer func(req1, req2 *http.Request) {
			req1.Body.Close()
			req2.Body.Close()
		}(req1, req2)

		// Create a copy of the request body for the second proxy
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		req1.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Copy of request body for the first proxy
		req2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Copy of request body for the second proxy

		// Create response recorders
		rec1 := httptest.NewRecorder()
		rec2 := httptest.NewRecorder()

		// Forward the request to the two proxies concurrently using goroutines
		var wg sync.WaitGroup
		wg.Add(2)

		// Forward the request to the first proxy
		go func() {
			defer wg.Done()
			proxy1.ServeHTTP(rec1, req1)
		}()

		// Forward the request to the second proxy
		go func() {
			defer wg.Done()
			proxy2.ServeHTTP(rec2, req2)
		}()

		wg.Wait() // Wait for both goroutines to finish

		//Compare the responses
		err = CompareResponses(rec1.Result(), rec2.Result())
		if err != nil {
			logger.Printf("[diff] Request: %s\n, error: %v", r.URL.RequestURI(), err)
		}

		w.WriteHeader(rec1.Result().StatusCode)
		_, _ = w.Write(rec1.Body.Bytes())
	}
}
