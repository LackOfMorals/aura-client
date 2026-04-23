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
	"sync/atomic"
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
	service := NewHTTPService(30*time.Second, 3, testLogger())

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Get(ctx, server.URL+"/test", nil)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Post(ctx, server.URL+"/test", nil, expectedBody)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Put(ctx, server.URL+"/test", nil, expectedBody)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Patch(ctx, server.URL+"/test", nil, expectedBody)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Delete(ctx, server.URL+"/test", nil)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, server.URL+"/test", expectedHeaders)

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

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, server.URL+"/test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestHTTPService_NonSuccessStatus verifies that non-2xx responses are returned
// without error — status code interpretation is the responsibility of the api layer.
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

			service := NewHTTPService(10*time.Second, 3, testLogger())

			ctx := context.Background()
			response, err := service.Get(ctx, server.URL+"/test", nil)

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

// TestHTTPService_ContextCancellation verifies context cancellation is respected
func TestHTTPService_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.Get(ctx, server.URL+"/test", nil)

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

// TestHTTPService_Timeout verifies the client-level timeout is enforced
func TestHTTPService_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(50*time.Millisecond, 0, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, server.URL+"/test", nil)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// TestHTTPService_LargeResponse verifies the response size limit is enforced
func TestHTTPService_LargeResponse(t *testing.T) {
	largeBody := strings.Repeat("x", DefaultMaxResponseSize+1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer server.Close()

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	response, err := service.Get(ctx, server.URL+"/test", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(response.Body) != DefaultMaxResponseSize {
		t.Errorf("expected response truncated to %d bytes, got %d", DefaultMaxResponseSize, len(response.Body))
	}
}

// TestHTTPService_InvalidURL verifies that a malformed URL returns an error
func TestHTTPService_InvalidURL(t *testing.T) {
	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	_, err := service.Get(ctx, "http://[invalid-url", nil)

	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

// TestHTTPResponse_Structure verifies the HTTPResponse struct fields
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
		t.Errorf("expected Content-Type 'application/json', got '%s'", httpResp.Headers.Get("Content-Type"))
	}
}

// TestHTTPService_ConcurrentRequests verifies thread safety under concurrent load
func TestHTTPService_ConcurrentRequests(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("response-%d", count)))
	}))
	defer server.Close()

	service := NewHTTPService(10*time.Second, 3, testLogger())

	ctx := context.Background()
	done := make(chan bool)
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.Get(ctx, server.URL+"/test", nil)
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

// TestHTTPService_CompleteURLRequired verifies that the httpClient layer expects
// fully-formed URLs — URL construction is the responsibility of the api layer above.
func TestHTTPService_CompleteURLRequired(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewHTTPService(10*time.Second, 3, testLogger())
	ctx := context.Background()

	tests := []struct {
		name         string
		url          string
		expectedPath string
		wantErr      bool
	}{
		{
			name:         "full URL with path",
			url:          server.URL + "/v1/instances",
			expectedPath: "/v1/instances",
			wantErr:      false,
		},
		{
			name:         "full URL with nested path",
			url:          server.URL + "/v1/instances/abc123/snapshots",
			expectedPath: "/v1/instances/abc123/snapshots",
			wantErr:      false,
		},
		{
			name:    "relative path is not a valid URL",
			url:     server.URL + "instances/abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Get(ctx, tt.url, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error for relative path, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if receivedPath != tt.expectedPath {
				t.Errorf("expected path '%s', got '%s'", tt.expectedPath, receivedPath)
			}
		})
	}
}


