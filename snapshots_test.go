package aura

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestSnapshotService creates a snapshotService with a mock API service for testing
func createTestSnapshotService(mock *mockAPIService) *snapshotService {
	return &snapshotService{
		api:     mock,
		timeout: 30 * time.Second,
		logger:  testLogger(),
	}
}

// createTestSnapshotServiceWithTimeout creates a snapshotService with a specific timeout.
// Pass the desired context directly to each method call.
func createTestSnapshotServiceWithTimeout(mock api.RequestService, timeout time.Duration) *snapshotService {
	return &snapshotService{
		api:     mock,
		timeout: timeout,
		logger:  testLogger(),
	}
}

// TestSnapshotService_List_Success verifies successful snapshot listing
func TestSnapshotService_List_Success(t *testing.T) {
	instanceID := "aaaa1234"
	expectedResponse := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{InstanceId: instanceID, SnapshotId: "snapshot-1", Profile: "daily", Status: "completed", Timestamp: "2024-01-01T00:00:00Z"},
			{InstanceId: instanceID, SnapshotId: "snapshot-2", Profile: "hourly", Status: "completed", Timestamp: "2024-01-01T12:00:00Z"},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List(context.Background(), instanceID, "")

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

// TestSnapshotService_Get_Success verifies successful obtaining of snapshot details
func TestSnapshotService_Get_Success(t *testing.T) {
	instanceID := "aaaa1234"
	snapshotID := "snapshot-1"
	expectedResponse := GetSnapshotDataResponse{
		Data: GetSnapshotData{
			InstanceId: instanceID, SnapshotId: snapshotID,
			Profile: "daily", Status: "completed", Timestamp: "2024-01-01T00:00:00Z", Exportable: true,
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Get(context.Background(), instanceID, snapshotID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/snapshots/"+snapshotID {
		t.Errorf("expected path 'instances/%s/snapshots/%s', got '%s'", instanceID, snapshotID, mock.lastPath)
	}
	if result.Data != expectedResponse.Data {
		t.Errorf("result does not match expected response, got '%v'", result)
	}
}

// TestSnapshotService_List_WithDate verifies listing snapshots for specific date
func TestSnapshotService_List_WithDate(t *testing.T) {
	instanceID := "aaaa1234"
	snapshotDate := "2024-01-15"
	responseBody, _ := json.Marshal(GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{InstanceId: instanceID, SnapshotId: "snapshot-date-1", Status: "completed", Timestamp: "2024-01-15T00:00:00Z"},
		},
	})
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List(context.Background(), instanceID, snapshotDate)

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
			_, err := service.List(context.Background(), "aaaa1234", tt.date)
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
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}
	service := createTestSnapshotService(mock)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List(context.Background(), "aaaa1234", tt.date)
			if err != nil {
				t.Errorf("expected no error for valid date '%s', got %v", tt.date, err)
			}
		})
	}
}

// TestSnapshotService_Create_Success verifies snapshot creation
func TestSnapshotService_Create_Success(t *testing.T) {
	instanceID := "aaaa1234"
	responseBody, _ := json.Marshal(CreateSnapshotResponse{
		Data: CreateSnapshotData{SnapshotId: "snapshot-new-456"},
	})
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Create(context.Background(), instanceID)

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
		err: &api.Error{StatusCode: 404, Message: "Instance not found"},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Create(context.Background(), "aaaaaaaa")

	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatal("expected Error type")
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestSnapshotService_Restore_Success verifies snapshot restore
func TestSnapshotService_Restore_Success(t *testing.T) {
	instanceID := "aaaa1234"
	snapshotID := "snapshot-123"
	responseBody, _ := json.Marshal(RestoreSnapshotResponse{
		Data: InstanceData{Id: instanceID, Status: "restoring"},
	})
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.Restore(context.Background(), instanceID, snapshotID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "POST" {
		t.Errorf("expected POST method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/snapshots/"+snapshotID+"/restore" {
		t.Errorf("expected restore path, got '%s'", mock.lastPath)
	}
	if result.Data.Status != "restoring" {
		t.Errorf("expected status 'restoring', got '%s'", result.Data.Status)
	}
}

// TestSnapshotService_List_EmptyResult verifies empty snapshot list
func TestSnapshotService_List_EmptyResult(t *testing.T) {
	responseBody, _ := json.Marshal(GetSnapshotsResponse{Data: []GetSnapshotData{}})
	mock := &mockAPIService{
		response: &api.Response{StatusCode: 200, Body: responseBody},
	}

	service := createTestSnapshotService(mock)
	result, err := service.List(context.Background(), "aaaa1234", "")

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
		err: &api.Error{StatusCode: 401, Message: "Invalid credentials"},
	}

	service := createTestSnapshotService(mock)
	_, err := service.List(context.Background(), "aaaa1234", "")

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

// ============================================================================
// Context-Specific Tests for SnapshotService
// ============================================================================

// TestSnapshotService_Create_ContextTimeout verifies timeout enforcement
func TestSnapshotService_Create_ContextTimeout(t *testing.T) {
	instanceID := "aaaa1234"
	responseBody, _ := json.Marshal(CreateSnapshotResponse{
		Data: CreateSnapshotData{SnapshotId: "snap-123"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{StatusCode: 200, Body: responseBody},
		delay:    2 * time.Second,
	}

	service := createTestSnapshotServiceWithTimeout(mock, 100*time.Millisecond)

	start := time.Now()
	_, err := service.Create(context.Background(), instanceID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

// TestSnapshotService_Get_ContextTimeout verifies Get with timeout
func TestSnapshotService_Get_ContextTimeout(t *testing.T) {
	instanceID := "aaaa1234"
	snapshotID := "snapshot-123"
	responseBody, _ := json.Marshal(GetSnapshotDataResponse{
		Data: GetSnapshotData{InstanceId: instanceID, SnapshotId: snapshotID},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{StatusCode: 200, Body: responseBody},
		delay:    500 * time.Millisecond,
	}

	service := createTestSnapshotServiceWithTimeout(mock, 50*time.Millisecond)

	start := time.Now()
	_, err := service.Get(context.Background(), instanceID, snapshotID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("timeout took too long: %v (expected ~50ms)", elapsed)
	}
}
