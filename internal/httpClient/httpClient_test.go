package httpClient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestNewHTTPRequestService verifies service creation with proper configuration
func TestNewHTTPRequestService(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 30 * time.Second
	maxRetry := 3
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	service := NewHTTPRequestService(baseURL, timeout, maxRetry, logger)

	if service == nil {
		t.Fatal("expected service to be non-nil")
	}

	concreteService, ok := service.(*HTTPRequestsService)
	if !ok {
		t.Fatal("expected service to be *HTTPRequestsService")
	}

	if concreteService.BaseURL != baseURL {
		t.Errorf("expected BaseURL '%s', got '%s'", baseURL, concreteService.BaseURL)
	}
	if concreteService.Timeout != timeout {
		t.Errorf("expected Timeout %v, got %v", timeout, concreteService.Timeout)
	}
	if concreteService.client == nil {
		t.Error("expected HTTP client to be initialized")
	}
	if concreteService.logger != logger {
		t.Error("expected logger to match provided logger")
	}
}

// TestToHTTPHeader verifies map to http.Header conversion
func TestToHTTPHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected http.Header
	}{
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: http.Header{},
		},
		{
			name: "single header",
			input: map[string]string{
				"Content-Type": "application/json",
			},
			expected: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name: "multiple headers",
			input: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
				"User-Agent":    "test-client",
			},
			expected: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer token123"},
				"User-Agent":    []string{"test-client"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toHTTPHeader(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d headers, got %d", len(tt.expected), len(result))
			}

			for key, expectedValues := range tt.expected {
				resultValues, exists := result[key]
				if !exists {
					t.Errorf("expected header '%s' not found", key)
					continue
				}
				if len(resultValues) != len(expectedValues) {
					t.Errorf("header '%s': expected %d values, got %d", key, len(expectedValues), len(resultValues))
					continue
				}
				if resultValues[0] != expectedValues[0] {
					t.Errorf("header '%s': expected value '%s', got '%s'", key, expectedValues[0], resultValues[0])
				}
			}
		})
	}
}

// TestMakeRequest_Success verifies successful HTTP request
func TestMakeRequest_Success(t *testing.T) {
	expectedResponse := `{"message":"success"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	response, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("expected response to be non-nil")
	}
	if response.ResponsePayload == nil {
		t.Fatal("expected ResponsePayload to be non-nil")
	}
	if string(*response.ResponsePayload) != expectedResponse {
		t.Errorf("expected payload '%s', got '%s'", expectedResponse, string(*response.ResponsePayload))
	}
	if response.RequestResponse.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.RequestResponse.StatusCode)
	}
}

// TestMakeRequest_WithHeaders verifies headers are properly set
func TestMakeRequest_WithHeaders(t *testing.T) {
	expectedHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer test-token",
		"Custom-Header": "custom-value",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers were received
		for key, expectedValue := range expectedHeaders {
			actualValue := r.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("header '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	_, err := service.MakeRequest(ctx, "/test", http.MethodGet, expectedHeaders, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestMakeRequest_WithBody verifies request body is sent
func TestMakeRequest_WithBody(t *testing.T) {
	expectedBody := `{"name":"test","value":123}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if string(body) != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	_, err := service.MakeRequest(ctx, "/test", http.MethodPost, nil, expectedBody)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestMakeRequest_EmptyBody verifies empty body handling
func TestMakeRequest_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if len(body) != 0 {
			t.Errorf("expected empty body, got %d bytes", len(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	_, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestMakeRequest_HTTPMethods verifies different HTTP methods
func TestMakeRequest_HTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected method %s, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

			ctx := context.Background()
			_, err := service.MakeRequest(ctx, "/test", method, nil, "")

			if err != nil {
				t.Fatalf("expected no error for %s, got %v", method, err)
			}
		})
	}
}

// TestMakeRequest_NonSuccessStatus verifies error handling for non-2xx status codes
func TestMakeRequest_NonSuccessStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			body:       `{"error":"bad request"}`,
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"unauthorized"}`,
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			body:       `{"error":"not found"}`,
		},
		// Note: 500 status codes are not tested here because retryablehttp
		// retries 5xx errors. After retries are exhausted, it returns both
		// a response AND an error, which causes issues with the test assertions.
		// Use non-retryable status codes (4xx) for error handling tests.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

			ctx := context.Background()
			response, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

			if err == nil {
				t.Fatal("expected error for non-2xx status code")
			}
			if response == nil {
				t.Fatal("expected response even on error")
			}
			if response.RequestResponse.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, response.RequestResponse.StatusCode)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", tt.statusCode)) {
				t.Errorf("expected error to contain status code, got: %v", err)
			}
		})
	}
}

// TestMakeRequest_ContextCancellation verifies context cancellation handling
func TestMakeRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

// TestMakeRequest_Timeout verifies timeout handling
func TestMakeRequest_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 50*time.Millisecond, 3, logger)

	ctx := context.Background()
	_, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// TestMakeRequest_LargeResponse verifies size limit enforcement
func TestMakeRequest_LargeResponse(t *testing.T) {
	// Create response larger than DefaultMaxResponseSize
	largeBody := strings.Repeat("x", DefaultMaxResponseSize+1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	response, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Response should be truncated to DefaultMaxResponseSize
	if len(*response.ResponsePayload) > DefaultMaxResponseSize {
		t.Errorf("expected response size <= %d, got %d", DefaultMaxResponseSize, len(*response.ResponsePayload))
	}
	if len(*response.ResponsePayload) != DefaultMaxResponseSize {
		t.Errorf("expected response to be truncated to %d, got %d", DefaultMaxResponseSize, len(*response.ResponsePayload))
	}
}

// TestMakeRequest_InvalidURL verifies handling of malformed URLs
func TestMakeRequest_InvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService("http://[invalid-url", 10*time.Second, 3, logger)

	ctx := context.Background()
	_, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")

	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

// TestCheckResponse verifies response validation
func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        []byte
		shouldError bool
	}{
		{
			name:        "200 OK",
			statusCode:  200,
			body:        []byte("success"),
			shouldError: false,
		},
		{
			name:        "201 Created",
			statusCode:  201,
			body:        []byte("created"),
			shouldError: false,
		},
		{
			name:        "204 No Content",
			statusCode:  204,
			body:        []byte(""),
			shouldError: false,
		},
		{
			name:        "299 Success",
			statusCode:  299,
			body:        []byte("success"),
			shouldError: false,
		},
		{
			name:        "300 Multiple Choices",
			statusCode:  300,
			body:        []byte("redirect"),
			shouldError: true,
		},
		{
			name:        "400 Bad Request",
			statusCode:  400,
			body:        []byte("bad request"),
			shouldError: true,
		},
		{
			name:        "404 Not Found",
			statusCode:  404,
			body:        []byte("not found"),
			shouldError: true,
		},
		{
			name:        "500 Internal Server Error",
			statusCode:  500,
			body:        []byte("server error"),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Status:     http.StatusText(tt.statusCode),
				Request:    req,
			}

			err := checkResponse(resp, tt.body)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if tt.shouldError && err != nil {
				if !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", tt.statusCode)) {
					t.Errorf("expected error to contain status code %d, got: %v", tt.statusCode, err)
				}
				if !strings.Contains(err.Error(), string(tt.body)) {
					t.Errorf("expected error to contain body, got: %v", err)
				}
			}
		})
	}
}

// TestHTTPResponse_Structure verifies HTTPResponse struct
func TestHTTPResponse_Structure(t *testing.T) {
	payload := []byte("test payload")
	resp := &http.Response{
		StatusCode: 200,
	}

	httpResp := &HTTPResponse{
		ResponsePayload: &payload,
		RequestResponse: resp,
	}

	if httpResp.ResponsePayload == nil {
		t.Error("expected ResponsePayload to be non-nil")
	}
	if httpResp.RequestResponse == nil {
		t.Error("expected RequestResponse to be non-nil")
	}
	if string(*httpResp.ResponsePayload) != "test payload" {
		t.Errorf("expected payload 'test payload', got '%s'", string(*httpResp.ResponsePayload))
	}
}

// TestMakeRequest_ConcurrentRequests verifies thread safety
func TestMakeRequest_ConcurrentRequests(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("response-%d", requestCount)))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewHTTPRequestService(server.URL, 10*time.Second, 3, logger)

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.MakeRequest(ctx, "/test", http.MethodGet, nil, "")
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	close(errors)

	for err := range errors {
		t.Errorf("concurrent request error: %v", err)
	}
}
