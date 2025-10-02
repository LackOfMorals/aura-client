package auraAPIClient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockServer creates a test HTTP server that returns predefined responses
func mockServer(t *testing.T, expectedMethod string, expectedPath string, expectedAuth string, responseCode int, responseBody interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != expectedMethod {
			t.Errorf("Expected method %s, got %s", expectedMethod, r.Method)
		}

		// Verify path
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify authorization header
		if expectedAuth != "" {
			auth := r.Header.Get("Authorization")
			if auth != expectedAuth {
				t.Errorf("Expected auth %s, got %s", expectedAuth, auth)
			}
		}

		// Verify User-Agent
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "jgHTTPClient" {
			t.Errorf("Expected User-Agent 'jgHTTPClient', got %s", userAgent)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		if responseBody != nil {
			json.NewEncoder(w).Encode(responseBody)
		}
	}))
}

func TestListInstances_Success(t *testing.T) {
	// Mock response
	mockResponse := ListInstancesResponse{
		Data: []ListInstanceData{
			{
				Id:   "instance-1",
				Name: "Test Instance 1",
			},
			{
				Id:   "instance-2",
				Name: "Test Instance 2",
			},
		},
	}

	server := mockServer(t, http.MethodGet, "/v1/instances", "Bearer test-token", http.StatusOK, mockResponse)
	defer server.Close()

	// Create service
	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	// Call function
	ctx := context.Background()
	result, err := service.ListInstances(ctx, token)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Data) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(result.Data))
	}

	if result.Data[0].Id != "instance-1" {
		t.Errorf("Expected instance Id 'instance-1', got %s", result.Data[0].Id)
	}
}

func TestListInstances_EmptyList(t *testing.T) {
	mockResponse := ListInstancesResponse{
		Data: []ListInstanceData{},
	}

	server := mockServer(t, http.MethodGet, "/v1/instances", "Bearer test-token", http.StatusOK, mockResponse)
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	ctx := context.Background()
	result, err := service.ListInstances(ctx, token)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("Expected empty list, got %d items", len(result.Data))
	}
}

func TestListInstances_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: "http://localhost:9999",
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	result, err := service.ListInstances(ctx, token)

	if err == nil {
		t.Fatal("Expected error due to cancelled context, got nil")
	}

	if result != nil {
		t.Error("Expected nil result with cancelled context")
	}
}

func TestListInstances_ContextTimeout(t *testing.T) {
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep to ensure timeout
	time.Sleep(5 * time.Millisecond)

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: "http://localhost:9999",
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	result, err := service.ListInstances(ctx, token)

	if err == nil {
		t.Fatal("Expected error due to timeout, got nil")
	}

	if result != nil {
		t.Error("Expected nil result with timeout")
	}
}

func TestGetInstance_Success(t *testing.T) {
	mockResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     "instance-123",
			Name:   "My Instance",
			Status: "running",
		},
	}

	server := mockServer(t, http.MethodGet, "/v1/instances/instance-123", "Bearer test-token", http.StatusOK, mockResponse)
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	ctx := context.Background()
	result, err := service.GetInstance(ctx, token, "instance-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Data.Id != "instance-123" {
		t.Errorf("Expected instance Id 'instance-123', got %s", result.Data.Id)
	}

	if result.Data.Status != "running" {
		t.Errorf("Expected status 'running', got %s", result.Data.Status)
	}
}

func TestCreateInstance_Success(t *testing.T) {
	mockResponse := CreateInstanceResponse{
		Data: CreateInstanceData{
			Id:       "new-instance-456",
			Name:     "New Instance",
			Username: "neo4j",
		},
	}

	server := mockServer(t, http.MethodPost, "/v1/instances", "Bearer test-token", http.StatusCreated, mockResponse)
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	instanceConfig := &CreateInstanceConfigData{
		Name:          "New Instance",
		Memory:        "8GB",
		Region:        "us-east-1",
		Version:       "5",
		TenantId:      "tenant-123",
		CloudProvider: "gcp",
	}

	ctx := context.Background()
	result, err := service.CreateInstance(ctx, token, instanceConfig)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Data.Id != "new-instance-456" {
		t.Errorf("Expected instance Id 'new-instance-456', got %s", result.Data.Id)
	}

	if result.Data.Name != "New Instance" {
		t.Errorf("Expected name 'New Instance', got %s", result.Data.Name)
	}
}

func TestDeleteInstance_Success(t *testing.T) {
	mockResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     "instance-789",
			Status: "deleting",
		},
	}

	server := mockServer(t, http.MethodDelete, "/v1/instances/instance-789", "Bearer test-token", http.StatusAccepted, mockResponse)
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	ctx := context.Background()
	result, err := service.DeleteInstance(ctx, token, "instance-789")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Data.Id != "instance-789" {
		t.Errorf("Expected instance Id 'instance-789', got %s", result.Data.Id)
	}

	if result.Data.Status != "deleting" {
		t.Errorf("Expected status 'deleting', got %s", result.Data.Status)
	}
}

func TestMakeAuthenticatedRequest_UserAgentConstant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			t.Errorf("Expected User-Agent constant to be used")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	ctx := context.Background()
	_, _ = service.ListInstances(ctx, token)
}

func TestMakeAuthenticatedRequest_AuthorizationHeader(t *testing.T) {
	expectedAuth := "Bearer my-secret-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != expectedAuth {
			t.Errorf("Expected Authorization header '%s', got '%s'", expectedAuth, auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "my-secret-token",
	}

	ctx := context.Background()
	_, _ = service.ListInstances(ctx, token)
}

func TestMakeAuthenticatedRequest_ContentTypeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer server.Close()

	service := &AuraAPIActionsService{
		AuraAPIBaseURL: server.URL,
		AuraAPIVersion: "/v1",
	}

	token := &AuthAPIToken{
		Type:  "Bearer",
		Token: "test-token",
	}

	ctx := context.Background()
	_, _ = service.ListInstances(ctx, token)
}
