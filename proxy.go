package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
)

// NewProxy creates a new reverse proxy to the target host
func NewProxy(targetHost string) *httputil.ReverseProxy {
	target, _ := url.Parse(targetHost)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}

// CompareResponses compares the responses from two servers
func CompareResponses(resp1, resp2 *http.Response) error {
	// Read the body of the first response
	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		logger.Printf("Error reading body from first response: %v", err)
		return errors.New("reading body from first error")
	}
	// Read the body of the second response
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		logger.Printf("Error reading body from second response: %v", err)
		return errors.New("reading body from second error")
	}

	// Compare status code if configured
	if config.CompareStatusCode && !isStatusEquivalent(resp1.StatusCode, resp2.StatusCode) {
		return errors.New(fmt.Sprintf("Status codes are different. First: %d, Second: %d", resp1.StatusCode, resp2.StatusCode))
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
	if !bytes.Equal(body1, body2) {
		return errors.New(fmt.Sprintf("Response bodies are different.\nFirst response: %s\nSecond response: %s", string(body1), string(body2)))
	}
	return nil
}

// HandleRequestAndRedirect handles incoming requests and redirects them to two target hosts
func HandleRequestAndRedirect(proxy1, proxy2 *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a copy of the request body for the second proxy
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Printf("Failed to read request body")
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		reqBodyCopy := io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Forward the request to the first proxy
		rec1 := httptest.NewRecorder()
		r.Body = reqBodyCopy
		proxy1.ServeHTTP(rec1, r)

		// Forward the request to the second proxy
		rec2 := httptest.NewRecorder()
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body for the second request
		proxy2.ServeHTTP(rec2, r)

		// Compare the responses
		err = CompareResponses(rec1.Result(), rec2.Result())
		if err != nil {
			logger.Printf("[diff] Request: %s\n, error: %v", r.URL.RequestURI(), err)
		}
	}
}
