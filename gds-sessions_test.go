package aura

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestGDSSessionService creates a gDSSessionService with a mock API service for testing
func createTestGDSSessionService(mock *mockAPIService) *gDSSessionService {
	return &gDSSessionService{
		api:    mock,
		ctx:    context.Background(),
		logger: testLogger(),
	}
}

// TestGDSSessionService_List_Success verifies successful GDS session listing
func TestGDSSessionService_List_Success(t *testing.T) {
	expectedResponse := GetGDSSessionResponse{
		Data: []GetGDSSessionData{
			{
				Id:            "session-1",
				Name:          "analytics-session-1",
				Memory:        "8GB",
				InstanceId:    "instance-1",
				DatabaseId:    "db-uuid-1",
				Status:        "running",
				Create:        "2024-01-01T00:00:00Z",
				Host:          "session1.gds.neo4j.io",
				Expiry:        "2024-01-02T00:00:00Z",
				Ttl:           "24h",
				UserId:        "user-1",
				TenantId:      "tenant-1",
				CloudProvider: "gcp",
				Region:        "us-central1",
			},
			{
				Id:            "session-2",
				Name:          "analytics-session-2",
				Memory:        "16GB",
				InstanceId:    "instance-2",
				Status:        "stopped",
				CloudProvider: "aws",
				Region:        "us-east-1",
			},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "graph-analytics/sessions" {
		t.Errorf("expected path 'graph-analytics/sessions', got '%s'", mock.lastPath)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 GDS sessions, got %d", len(result.Data))
	}
	if result.Data[0].Id != "session-1" {
		t.Errorf("expected first session ID 'session-1', got '%s'", result.Data[0].Id)
	}
	if result.Data[0].Name != "analytics-session-1" {
		t.Errorf("expected first session name 'analytics-session-1', got '%s'", result.Data[0].Name)
	}
}

// TestGDSSessionService_List_EmptyResult verifies empty session list
func TestGDSSessionService_List_EmptyResult(t *testing.T) {
	expectedResponse := GetGDSSessionResponse{Data: []GetGDSSessionData{}}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 GDS sessions, got %d", len(result.Data))
	}
}

// TestGDSSessionService_List_SingleSession verifies listing with single session
func TestGDSSessionService_List_SingleSession(t *testing.T) {
	expectedResponse := GetGDSSessionResponse{
		Data: []GetGDSSessionData{
			{
				Id:            "session-single",
				Name:          "only-session",
				Memory:        "32GB",
				Status:        "running",
				CloudProvider: "gcp",
				Region:        "europe-west2",
			},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 GDS session, got %d", len(result.Data))
	}
	if result.Data[0].Id != "session-single" {
		t.Errorf("expected session ID 'session-single', got '%s'", result.Data[0].Id)
	}
}

// TestGDSSessionService_List_MultipleStatuses verifies sessions with different statuses
func TestGDSSessionService_List_MultipleStatuses(t *testing.T) {
	expectedResponse := GetGDSSessionResponse{
		Data: []GetGDSSessionData{
			{Id: "session-1", Status: "running"},
			{Id: "session-2", Status: "stopped"},
			{Id: "session-3", Status: "creating"},
			{Id: "session-4", Status: "failed"},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 4 {
		t.Errorf("expected 4 GDS sessions, got %d", len(result.Data))
	}

	// Verify different statuses
	statuses := make(map[string]bool)
	for _, session := range result.Data {
		statuses[session.Status] = true
	}

	expectedStatuses := []string{"running", "stopped", "creating", "failed"}
	for _, status := range expectedStatuses {
		if !statuses[status] {
			t.Errorf("expected to find status '%s' in results", status)
		}
	}
}

// TestGDSSessionService_List_FullSessionDetails verifies all session fields
func TestGDSSessionService_List_FullSessionDetails(t *testing.T) {
	expectedSession := GetGDSSessionData{
		Id:            "session-full",
		Name:          "complete-session",
		Memory:        "16GB",
		InstanceId:    "instance-abc123",
		DatabaseId:    "db-uuid-xyz789",
		Status:        "running",
		Create:        "2024-01-15T10:30:00Z",
		Host:          "session-full.gds.neo4j.io",
		Expiry:        "2024-01-22T10:30:00Z",
		Ttl:           "7d",
		UserId:        "user-abc",
		TenantId:      "tenant-xyz",
		CloudProvider: "gcp",
		Region:        "europe-west2",
	}

	responseBody, _ := json.Marshal(GetGDSSessionResponse{Data: []GetGDSSessionData{expectedSession}})
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result.Data))
	}

	session := result.Data[0]

	if session.Id != expectedSession.Id {
		t.Errorf("expected ID '%s', got '%s'", expectedSession.Id, session.Id)
	}
	if session.Name != expectedSession.Name {
		t.Errorf("expected name '%s', got '%s'", expectedSession.Name, session.Name)
	}
	if session.Memory != expectedSession.Memory {
		t.Errorf("expected memory '%s', got '%s'", expectedSession.Memory, session.Memory)
	}
	if session.Status != expectedSession.Status {
		t.Errorf("expected status '%s', got '%s'", expectedSession.Status, session.Status)
	}
	if session.CloudProvider != expectedSession.CloudProvider {
		t.Errorf("expected cloud provider '%s', got '%s'", expectedSession.CloudProvider, session.CloudProvider)
	}
	if session.Region != expectedSession.Region {
		t.Errorf("expected region '%s', got '%s'", expectedSession.Region, session.Region)
	}
}

// TestGDSSessionService_List_AuthenticationError verifies auth error handling
func TestGDSSessionService_List_AuthenticationError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid credentials"},
	}

	service := createTestGDSSessionService(mock)
	_, err := service.List()

	if err == nil {
		t.Fatal("expected authentication error")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if !apiErr.IsUnauthorized() {
		t.Error("expected IsUnauthorized() to be true")
	}
}

// TestGDSSessionService_List_ServerError verifies server error handling
func TestGDSSessionService_List_ServerError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{StatusCode: http.StatusBadRequest, Message: "Bad request error"},
	}

	service := createTestGDSSessionService(mock)
	result, err := service.List()

	if err == nil {
		t.Fatal("expected server error")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
}
