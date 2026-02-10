package aura

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestCmekService creates a cmekService with a mock API service for testing
func createTestCmekService(mock *mockAPIService) *cmekService {
	return &cmekService{
		api:    mock,
		ctx:    context.Background(),
		logger: testLogger(),
	}
}

// TestCmekService_List_Success verifies successful CMEK listing
func TestCmekService_List_Success(t *testing.T) {
	expectedResponse := GetCmeksResponse{
		Data: []GetCmeksData{
			{Id: "cmek-1", Name: "Production Key", TenantId: "tenant-1"},
			{Id: "cmek-2", Name: "Development Key", TenantId: "tenant-1"},
			{Id: "cmek-3", Name: "Testing Key", TenantId: "tenant-2"},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestCmekService(mock)
	result, err := service.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "customer-managed-keys" {
		t.Errorf("expected path 'customer-managed-keys', got '%s'", mock.lastPath)
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 CMEKs, got %d", len(result.Data))
	}
}

// TestCmekService_List_WithTenantFilter verifies tenant ID filtering
func TestCmekService_List_WithTenantFilter(t *testing.T) {
	tenantID := "c1e2c556-a924-5fac-b7f8-bb624ad9761d"
	expectedResponse := GetCmeksResponse{
		Data: []GetCmeksData{
			{Id: "cmek-filtered-1", Name: "Filtered Key 1", TenantId: tenantID},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestCmekService(mock)
	result, err := service.List(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastPath != "customer-managed-keys?tenantid="+tenantID {
		t.Errorf("expected path with tenant filter, got '%s'", mock.lastPath)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 CMEK, got %d", len(result.Data))
	}
}

// TestCmekService_List_InvalidTenantID verifies tenant ID validation
func TestCmekService_List_InvalidTenantID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
	}{
		{"too short", "abc"},
		{"wrong length", "not-valid-uuid"},
	}

	mock := &mockAPIService{}
	service := createTestCmekService(mock)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List(tt.tenantID)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

// TestCmekService_List_EmptyResult verifies empty CMEK list
func TestCmekService_List_EmptyResult(t *testing.T) {
	expectedResponse := GetCmeksResponse{Data: []GetCmeksData{}}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestCmekService(mock)
	result, err := service.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 CMEKs, got %d", len(result.Data))
	}
}

// TestCmekService_List_AuthenticationError verifies auth error handling
func TestCmekService_List_AuthenticationError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.Error{StatusCode: http.StatusUnauthorized, Message: "Invalid credentials"},
	}

	service := createTestCmekService(mock)
	_, err := service.List("")

	if err == nil {
		t.Fatal("expected authentication error")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatal("expected Error type")
	}
	if !apiErr.IsUnauthorized() {
		t.Error("expected IsUnauthorized() to be true")
	}
}
