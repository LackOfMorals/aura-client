package aura

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestSnapshotService creates a snapshotService with a mock API service for testing
func createTestSnapshotService(mock *mockAPIService) *snapshotService {
	return &snapshotService{
		api:    mock,
		ctx:    context.Background(),
		logger: testLogger(),
	}
}

// TestSnapshotService_List_Success verifies successful snapshot listing
func TestSnapshotService_List_Success(t *testing.T) {
	instanceID := "instance-123"
	expectedResponse := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{InstanceId: instanceID, SnapshotId: "snapshot-1", Profile: "daily", Status: "completed", Timestamp: "2024-01-01T00:00:00Z"},
			{InstanceId: instanceID, SnapshotId: "snapshot-2", Profile: "hourly", Status: "completed", Timestamp: "2024-01-01T12:00:00Z"},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List(instanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/snapshots" {
		t.Errorf("expected path 'instances/%s/snapshots', got '%s'", instanceID, mock.lastPath)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(result.Data))
	}
}

// TestSnapshotService_List_WithDate verifies listing snapshots for specific date
func TestSnapshotService_List_WithDate(t *testing.T) {
	instanceID := "instance-123"
	snapshotDate := "2024-01-15"
	expectedResponse := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{InstanceId: instanceID, SnapshotId: "snapshot-date-1", Status: "completed", Timestamp: "2024-01-15T00:00:00Z"},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List(instanceID, snapshotDate)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastPath != "instances/"+instanceID+"/snapshots?date="+snapshotDate {
		t.Errorf("expected path with date, got '%s'", mock.lastPath)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(result.Data))
	}
}

// TestSnapshotService_List_InvalidDateFormat verifies date format validation
func TestSnapshotService_List_InvalidDateFormat(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{"wrong separator", "2024/01/15"},
		{"too short", "2024-01"},
		{"too long", "2024-01-15-extra"},
		{"random string", "not-a-date"},
	}

	mock := &mockAPIService{}
	service := createTestSnapshotService(mock)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List("instance-123", tt.date)
			if err == nil {
				t.Errorf("expected error for invalid date format '%s'", tt.date)
			}
		})
	}
}

// TestSnapshotService_List_ValidDateFormats verifies valid date formats
func TestSnapshotService_List_ValidDateFormats(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{"valid date", "2024-01-15"},
		{"leap year", "2024-02-29"},
		{"end of year", "2024-12-31"},
		{"start of year", "2024-01-01"},
	}

	responseBody, _ := json.Marshal(GetSnapshotsResponse{Data: []GetSnapshotData{}})
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}
	service := createTestSnapshotService(mock)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List("instance-123", tt.date)
			if err != nil {
				t.Errorf("expected no error for valid date '%s', got %v", tt.date, err)
			}
		})
	}
}

// TestSnapshotService_Create_Success verifies snapshot creation
func TestSnapshotService_Create_Success(t *testing.T) {
	instanceID := "instance-123"
	expectedResponse := CreateSnapshotResponse{
		Data: CreateSnapshotData{SnapshotId: "snapshot-new-456"},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Create(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "POST" {
		t.Errorf("expected POST method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/snapshots" {
		t.Errorf("expected path 'instances/%s/snapshots', got '%s'", instanceID, mock.lastPath)
	}
	if result.Data.SnapshotId != "snapshot-new-456" {
		t.Errorf("expected snapshot ID 'snapshot-new-456', got '%s'", result.Data.SnapshotId)
	}
}

// TestSnapshotService_Create_InstanceNotFound verifies error when instance doesn't exist
func TestSnapshotService_Create_InstanceNotFound(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{StatusCode: http.StatusNotFound, Message: "Instance not found"},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Create("nonexistent-instance")

	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestSnapshotService_List_EmptyResult verifies empty snapshot list
func TestSnapshotService_List_EmptyResult(t *testing.T) {
	expectedResponse := GetSnapshotsResponse{Data: []GetSnapshotData{}}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List("instance-123", "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(result.Data))
	}
}

// TestSnapshotService_AuthenticationError verifies auth error handling
func TestSnapshotService_AuthenticationError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid credentials"},
	}

	service := createTestSnapshotService(mock)
	_, err := service.List("instance-123", "")

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
