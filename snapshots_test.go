package aura

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// setupSnapshotTestClient creates a test client with a mock server
func setupSnapshotTestClient(handler http.HandlerFunc) (*AuraAPIClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	
	client, _ := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(10*time.Second),
	)
	
	// Update both the config baseURL and the transport's BaseURL
	client.config.baseURL = server.URL + "/"
	if transport, ok := (*client.transport).(*httpClient.HTTPRequestsService); ok {
		transport.BaseURL = server.URL + "/"
	}
	
	return client, server
}

// TestSnapshotService_List_Success verifies successful snapshot listing
func TestSnapshotService_List_Success(t *testing.T) {
	instanceID := "instance-123"
	expectedSnapshots := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-1",
				Profile:    "daily",
				Status:     "completed",
				Timestamp:  "2024-01-01T00:00:00Z",
			},
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-2",
				Profile:    "hourly",
				Status:     "completed",
				Timestamp:  "2024-01-01T12:00:00Z",
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSnapshots)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.List(instanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(result.Data))
	}
	if result.Data[0].SnapshotId != "snapshot-1" {
		t.Errorf("expected first snapshot ID 'snapshot-1', got '%s'", result.Data[0].SnapshotId)
	}
	if result.Data[1].Profile != "hourly" {
		t.Errorf("expected second snapshot profile 'hourly', got '%s'", result.Data[1].Profile)
	}
}

// TestSnapshotService_List_WithDate verifies listing snapshots for specific date
func TestSnapshotService_List_WithDate(t *testing.T) {
	instanceID := "instance-123"
	snapshotDate := "2024-01-15"
	expectedSnapshots := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-date-1",
				Profile:    "daily",
				Status:     "completed",
				Timestamp:  "2024-01-15T00:00:00Z",
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodGet {
			// Verify date parameter
			queryDate := r.URL.Query().Get("date")
			if queryDate != snapshotDate {
				t.Errorf("expected date parameter '%s', got '%s'", snapshotDate, queryDate)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSnapshots)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.List(instanceID, snapshotDate)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(result.Data))
	}
	if result.Data[0].Timestamp != "2024-01-15T00:00:00Z" {
		t.Errorf("expected timestamp '2024-01-15T00:00:00Z', got '%s'", result.Data[0].Timestamp)
	}
}

// TestSnapshotService_List_InvalidDateFormat verifies date format validation
func TestSnapshotService_List_InvalidDateFormat(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{
			name: "wrong separator",
			date: "2024/01/15",
		},
		{
			name: "missing leading zero",
			date: "2024-1-15",
		},
		{
			name: "invalid date",
			date: "2024-13-45",
		},
		{
			name: "too short",
			date: "2024-01",
		},
		{
			name: "too long",
			date: "2024-01-15-extra",
		},
		{
			name: "random string",
			date: "not-a-date",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Snapshots.List("instance-123", tt.date)

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
		{
			name: "valid date",
			date: "2024-01-15",
		},
		{
			name: "leap year",
			date: "2024-02-29",
		},
		{
			name: "end of year",
			date: "2024-12-31",
		},
		{
			name: "start of year",
			date: "2024-01-01",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GetSnapshotsResponse{
			Data: []GetSnapshotData{},
		})
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Snapshots.List("instance-123", tt.date)

			if err != nil {
				t.Errorf("expected no error for valid date '%s', got %v", tt.date, err)
			}
		})
	}
}

// TestSnapshotService_List_EmptyDate verifies listing without date parameter
func TestSnapshotService_List_EmptyDate(t *testing.T) {
	instanceID := "instance-123"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodGet {
			// Verify no date parameter
			if r.URL.Query().Get("date") != "" {
				t.Error("expected no date parameter in query")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetSnapshotsResponse{
				Data: []GetSnapshotData{},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	_, err := client.Snapshots.List(instanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestSnapshotService_Create_Success verifies snapshot creation
func TestSnapshotService_Create_Success(t *testing.T) {
	instanceID := "instance-123"
	expectedResponse := CreateSnapshotResponse{
		Data: CreateSnapshotData{
			SnapshotId: "snapshot-new-456",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.Create(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Data.SnapshotId != "snapshot-new-456" {
		t.Errorf("expected snapshot ID 'snapshot-new-456', got '%s'", result.Data.SnapshotId)
	}
}

// TestSnapshotService_Create_InstanceNotFound verifies error when instance doesn't exist
func TestSnapshotService_Create_InstanceNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Instance not found",
		})
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.Create("nonexistent-instance")

	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestSnapshotService_List_EmptyResult verifies empty snapshot list
func TestSnapshotService_List_EmptyResult(t *testing.T) {
	instanceID := "instance-123"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetSnapshotsResponse{
				Data: []GetSnapshotData{},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.List(instanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(result.Data))
	}
}

// TestSnapshotService_List_MultipleStatuses verifies snapshots with different statuses
func TestSnapshotService_List_MultipleStatuses(t *testing.T) {
	instanceID := "instance-123"
	expectedSnapshots := GetSnapshotsResponse{
		Data: []GetSnapshotData{
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-1",
				Status:     "completed",
			},
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-2",
				Status:     "in_progress",
			},
			{
				InstanceId: instanceID,
				SnapshotId: "snapshot-3",
				Status:     "failed",
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/snapshots" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSnapshots)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.List(instanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(result.Data))
	}

	// Verify different statuses
	statuses := make(map[string]bool)
	for _, snapshot := range result.Data {
		statuses[snapshot.Status] = true
	}

	expectedStatuses := []string{"completed", "in_progress", "failed"}
	for _, status := range expectedStatuses {
		if !statuses[status] {
			t.Errorf("expected to find status '%s' in results", status)
		}
	}
}

// TestSnapshotService_AuthenticationError verifies auth error handling
func TestSnapshotService_AuthenticationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	_, err := client.Snapshots.List("instance-123", "")

	if err == nil {
		t.Fatal("expected authentication error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if !apiErr.IsUnauthorized() {
		t.Error("expected IsUnauthorized() to be true")
	}
}

// TestSnapshotService_Create_ServerError verifies server error handling
func TestSnapshotService_Create_ServerError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		// Use 400 instead of 500 to avoid retry behavior in retryablehttp
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Bad request error",
		})
	}

	client, server := setupSnapshotTestClient(handler)
	defer server.Close()

	result, err := client.Snapshots.Create("instance-123")

	if err == nil {
		t.Fatal("expected server error")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
}
