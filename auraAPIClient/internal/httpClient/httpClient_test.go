package httpClient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPRequestService(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 30 * time.Second

	service := NewHTTPRequestService(baseURL, timeout)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	concreteService, ok := service.(*HTTPRequestsService)
	if !ok {
		t.Fatal("Expected service to be of type *HTTPRequestsService")
	}

	if concreteService.BaseURL != baseURL {
		t.Errorf("Expected BaseURL to be %s, got %s", baseURL, concreteService.BaseURL)
	}

	if concreteService.Timeout != timeout {
		t.Errorf("Expected Timeout to be %v, got %v", timeout, concreteService.Timeout)
	}
}

func TestMakeRequest_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		// Verify endpoint
		if r.URL.Path != "/test" {
			t.Errorf("Expected path /test, got %s", r.URL.Path)
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/test", http.MethodGet, nil, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.ResponsePayload == nil {
		t.Fatal("Expected response payload, got nil")
	}

	expected := `{"message": "success"}`
	if string(*resp.ResponsePayload) != expected {
		t.Errorf("Expected payload %s, got %s", expected, string(*resp.ResponsePayload))
	}

	if resp.RequestResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.RequestResponse.StatusCode)
	}
}

func TestMakeRequest_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
	}

	resp, err := service.MakeRequest(context.Background(), "/test", http.MethodGet, header, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestMakeRequest_WithBody(t *testing.T) {
	expectedBody := `{"name": "test", "value": 123}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify body
		body := make([]byte, len(expectedBody))
		_, err := r.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			t.Errorf("Error reading body: %v", err)
		}

		if string(body) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/create", http.MethodPost, nil, expectedBody)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.RequestResponse.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", resp.RequestResponse.StatusCode)
	}
}

func TestMakeRequest_4xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/notfound", http.MethodGet, nil, "")
	if err == nil {
		t.Fatal("Expected error for 404 status, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response on error, got %v", resp)
	}
}

func TestMakeRequest_5xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/error", http.MethodGet, nil, "")
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response on error, got %v", resp)
	}
}

func TestMakeRequest_InvalidURL(t *testing.T) {
	service := NewHTTPRequestService("http://invalid-domain-that-does-not-exist-12345.com", 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/test", http.MethodGet, nil, "")
	if err == nil {
		t.Fatal("Expected error for invalid domain, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response on error, got %v", resp)
	}
}

func TestMakeRequest_AllHTTPMethods(t *testing.T) {
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
					t.Errorf("Expected method %s, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))
			defer server.Close()

			service := NewHTTPRequestService(server.URL, 30*time.Second)

			resp, err := service.MakeRequest(context.Background(), "/test", method, nil, "")
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", method, err)
			}

			if resp == nil {
				t.Fatalf("Expected response for %s, got nil", method)
			}
		})
	}
}

func TestMakeRequest_JSONResponse(t *testing.T) {
	type Response struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
		Status  bool   `json:"status"`
	}

	expectedResp := Response{
		ID:      42,
		Message: "test message",
		Status:  true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	service := NewHTTPRequestService(server.URL, 30*time.Second)

	resp, err := service.MakeRequest(context.Background(), "/test", http.MethodGet, nil, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var actualResp Response
	err = json.Unmarshal(*resp.ResponsePayload, &actualResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if actualResp.ID != expectedResp.ID {
		t.Errorf("Expected ID %d, got %d", expectedResp.ID, actualResp.ID)
	}

	if actualResp.Message != expectedResp.Message {
		t.Errorf("Expected Message %s, got %s", expectedResp.Message, actualResp.Message)
	}

	if actualResp.Status != expectedResp.Status {
		t.Errorf("Expected Status %v, got %v", expectedResp.Status, actualResp.Status)
	}
}

func TestCheckResponse_Success(t *testing.T) {
	statusCodes := []int{200, 201, 204, 299}

	for _, code := range statusCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: code,
				Request:    &http.Request{},
			}

			err := checkResponse(resp, nil)
			if err != nil {
				t.Errorf("Expected no error for status %d, got %v", code, err)
			}
		})
	}
}

func TestCheckResponse_Errors(t *testing.T) {
	errorCodes := []int{400, 401, 403, 404, 500, 502, 503}

	for _, code := range errorCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: code,
				Status:     http.StatusText(code),
				Request: &http.Request{
					Method: http.MethodGet,
				},
			}

			err := checkResponse(resp, nil)
			if err == nil {
				t.Errorf("Expected error for status %d, got nil", code)
			}
		})
	}
}
