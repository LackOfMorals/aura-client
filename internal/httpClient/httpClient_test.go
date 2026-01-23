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

// testLogger creates a logger for testing
func testLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

// TestNewHTTPService verifies service creation with proper configuration
func TestNewHTTPService(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 30 * time.Second
	maxRetry := 3
	logger := testLogger()

	service := NewHTTPService(baseURL, timeout, maxRetry, logger)

	if service == nil {
		t.Fatal("expected service to be non-nil")
	}
}

// TestHTTPService_Get_Success verifies successful GET request
func TestHTTPService_Get_Success(t *testing.T) {
	expectedResponse := `{"message":"success"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Get(ctx, "test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("expected response to be non-nil")
	}
	if string(response.Body) != expectedResponse {
		t.Errorf("expected body '%s', got '%s'", expectedResponse, string(response.Body))
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.StatusCode)
	}
}

// TestHTTPService_Post_Success verifies successful POST request
func TestHTTPService_Post_Success(t *testing.T) {
	expectedBody := `{"name":"test","value":123}`
	expectedResponse := `{"id":"created-123"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Post(ctx, "test", nil, expectedBody)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", response.StatusCode)
	}
	if string(response.Body) != expectedResponse {
		t.Errorf("expected body '%s', got '%s'", expectedResponse, string(response.Body))
	}
}

// TestHTTPService_Put_Success verifies successful PUT request
func TestHTTPService_Put_Success(t *testing.T) {
	expectedBody := `{"name":"updated"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Put(ctx, "test", nil, expectedBody)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.StatusCode)
	}
}

// TestHTTPService_Patch_Success verifies successful PATCH request
func TestHTTPService_Patch_Success(t *testing.T) {
	expectedBody := `{"name":"patched"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Patch(ctx, "test", nil, expectedBody)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.StatusCode)
	}
}

// TestHTTPService_Delete_Success verifies successful DELETE request
func TestHTTPService_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Delete(ctx, "test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", response.StatusCode)
	}
}

// TestHTTPService_WithHeaders verifies headers are properly set
func TestHTTPService_WithHeaders(t *testing.T) {
	expectedHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer test-token",
		"Custom-Header": "custom-value",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, expectedValue := range expectedHeaders {
			actualValue := r.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("header '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, "test", expectedHeaders)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestHTTPService_EmptyBody verifies empty body handling
func TestHTTPService_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("expected empty body, got %d bytes", len(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, "test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestHTTPService_NonSuccessStatus verifies handling of non-2xx status codes
func TestHTTPService_NonSuccessStatus(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

			ctx := context.Background()
			response, err := service.Get(ctx, "test", nil)

			// The new implementation doesn't return an error for non-2xx status codes
			// It just returns the response with the status code
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if response == nil {
				t.Fatal("expected response to be non-nil")
			}
			if response.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, response.StatusCode)
			}
			if string(response.Body) != tt.body {
				t.Errorf("expected body '%s', got '%s'", tt.body, string(response.Body))
			}
		})
	}
}

// TestHTTPService_ContextCancellation verifies context cancellation handling
func TestHTTPService_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.Get(ctx, "test", nil)

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

// TestHTTPService_Timeout verifies timeout handling
func TestHTTPService_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 50*time.Millisecond, 0, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, "test", nil)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// TestHTTPService_LargeResponse verifies size limit enforcement
func TestHTTPService_LargeResponse(t *testing.T) {
	// Create response larger than DefaultMaxResponseSize
	largeBody := strings.Repeat("x", DefaultMaxResponseSize+1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Get(ctx, "test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Response should be truncated to DefaultMaxResponseSize
	if len(response.Body) > DefaultMaxResponseSize {
		t.Errorf("expected response size <= %d, got %d", DefaultMaxResponseSize, len(response.Body))
	}
	if len(response.Body) != DefaultMaxResponseSize {
		t.Errorf("expected response to be truncated to %d, got %d", DefaultMaxResponseSize, len(response.Body))
	}
}

// TestHTTPService_InvalidURL verifies handling of malformed URLs
func TestHTTPService_InvalidURL(t *testing.T) {
	service := NewHTTPService("http://[invalid-url", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, "test", nil)

	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

// TestHTTPResponse_Structure verifies HTTPResponse struct
func TestHTTPResponse_Structure(t *testing.T) {
	body := []byte("test payload")
	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	httpResp := &HTTPResponse{
		StatusCode: 200,
		Body:       body,
		Headers:    headers,
	}

	if httpResp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", httpResp.StatusCode)
	}
	if string(httpResp.Body) != "test payload" {
		t.Errorf("expected body 'test payload', got '%s'", string(httpResp.Body))
	}
	if httpResp.Headers.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type header 'application/json', got '%s'", httpResp.Headers.Get("Content-Type"))
	}
}

// TestHTTPService_ConcurrentRequests verifies thread safety
func TestHTTPService_ConcurrentRequests(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("response-%d", requestCount)))
	}))
	defer server.Close()

	service := NewHTTPService(server.URL+"/", 10*time.Second, 3, testLogger())

	ctx := context.Background()
	done := make(chan bool)
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.Get(ctx, "test", nil)
			if err != nil {
				errs <- err
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	close(errs)

	for err := range errs {
		t.Errorf("concurrent request error: %v", err)
	}
}

// TestMockHTTPService verifies the mock implementation
func TestMockHTTPService_Get(t *testing.T) {
	mock := NewMockHTTPService()
	mock.WithResponse(200, `{"status":"ok"}`)

	ctx := context.Background()
	resp, err := mock.Get(ctx, "/test", map[string]string{"X-Test": "value"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if mock.LastMethod != "GET" {
		t.Errorf("expected method GET, got %s", mock.LastMethod)
	}
	if mock.LastURL != "/test" {
		t.Errorf("expected URL /test, got %s", mock.LastURL)
	}
	if mock.CallCount != 1 {
		t.Errorf("expected call count 1, got %d", mock.CallCount)
	}
}

// TestMockHTTPService_Post verifies the mock POST implementation
func TestMockHTTPService_Post(t *testing.T) {
	mock := NewMockHTTPService()
	mock.WithPostResponse(201, `{"id":"123"}`)

	ctx := context.Background()
	resp, err := mock.Post(ctx, "/test", nil, `{"name":"test"}`)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
	if mock.LastBody != `{"name":"test"}` {
		t.Errorf("expected body '{\"name\":\"test\"}', got '%s'", mock.LastBody)
	}
}

// TestMockHTTPService_Error verifies error handling
func TestMockHTTPService_Error(t *testing.T) {
	mock := NewMockHTTPService()
	expectedErr := fmt.Errorf("network error")
	mock.WithError(expectedErr)

	ctx := context.Background()
	_, err := mock.Get(ctx, "/test", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

// TestMockHTTPService_Reset verifies reset functionality
func TestMockHTTPService_Reset(t *testing.T) {
	mock := NewMockHTTPService()
	mock.WithResponse(200, "test")

	ctx := context.Background()
	mock.Get(ctx, "/test", nil)
	mock.Reset()

	if mock.CallCount != 0 {
		t.Errorf("expected call count 0 after reset, got %d", mock.CallCount)
	}
	if mock.Response != nil {
		t.Error("expected nil response after reset")
	}
	if len(mock.CallHistory) != 0 {
		t.Errorf("expected empty call history after reset, got %d", len(mock.CallHistory))
	}
}

// TestMockHTTPService_CallHistory verifies call history tracking
func TestMockHTTPService_CallHistory(t *testing.T) {
	mock := NewMockHTTPService()
	mock.WithResponse(200, "ok")

	ctx := context.Background()
	mock.Get(ctx, "/test1", map[string]string{"X-Test": "1"})
	mock.Post(ctx, "/test2", map[string]string{"X-Test": "2"}, "body")
	mock.Delete(ctx, "/test3", nil)

	if len(mock.CallHistory) != 3 {
		t.Fatalf("expected 3 calls in history, got %d", len(mock.CallHistory))
	}

	if mock.CallHistory[0].Method != "GET" {
		t.Errorf("expected first call to be GET, got %s", mock.CallHistory[0].Method)
	}
	if mock.CallHistory[1].Method != "POST" {
		t.Errorf("expected second call to be POST, got %s", mock.CallHistory[1].Method)
	}
	if mock.CallHistory[1].Body != "body" {
		t.Errorf("expected second call body 'body', got '%s'", mock.CallHistory[1].Body)
	}
	if mock.CallHistory[2].Method != "DELETE" {
		t.Errorf("expected third call to be DELETE, got %s", mock.CallHistory[2].Method)
	}
}
