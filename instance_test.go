package aura

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// setupInstanceTestClient creates a test client with a mock server
func setupInstanceTestClient(handler http.HandlerFunc) (*AuraAPIClient, *httptest.Server) {
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

// TestInstanceService_List_Success verifies successful instance listing
func TestInstanceService_List_Success(t *testing.T) {
	expectedInstances := ListInstancesResponse{
		Data: []ListInstanceData{
			{
				Id:            "instance-1",
				Name:          "test-instance-1",
				Created:       "2024-01-01T00:00:00Z",
				TenantId:      "tenant-1",
				CloudProvider: "gcp",
			},
			{
				Id:            "instance-2",
				Name:          "test-instance-2",
				Created:       "2024-01-02T00:00:00Z",
				TenantId:      "tenant-1",
				CloudProvider: "aws",
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

		if r.URL.Path == "/v1/instances" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedInstances)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 instances, got %d", len(result.Data))
	}
	if result.Data[0].Id != "instance-1" {
		t.Errorf("expected first instance ID 'instance-1', got '%s'", result.Data[0].Id)
	}
	if result.Data[1].Name != "test-instance-2" {
		t.Errorf("expected second instance name 'test-instance-2', got '%s'", result.Data[1].Name)
	}
}

// TestInstanceService_Get_Success verifies retrieving a specific instance
func TestInstanceService_Get_Success(t *testing.T) {
	instanceID := "aaaa5678"
	expectedInstance := GetInstanceResponse{
		Data: GetInstanceData{
			Id:            instanceID,
			Name:          "my-instance",
			Status:        "running",
			TenantId:      "tenant-1",
			CloudProvider: "gcp",
			ConnectionUrl: "neo4j+s://xxxxx.databases.neo4j.io",
			Region:        "us-east-1",
			Type:          "enterprise-db",
			Memory:        "8GB",
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

		if r.URL.Path == "/v1/instances/"+instanceID && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedInstance)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Get(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Data.Id != instanceID {
		t.Errorf("expected instance ID '%s', got '%s'", instanceID, result.Data.Id)
	}
	if result.Data.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Get_NotFound verifies 404 handling
func TestInstanceService_Get_NotFound(t *testing.T) {
	// Use a valid format instance ID (8 hex chars) that doesn't exist
	nonExistentID := "aaaaaaaa"

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

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Get(nonExistentID)

	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError type, got %T: %v", err, err)
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestInstanceService_Create_Success verifies instance creation
func TestInstanceService_Create_Success(t *testing.T) {
	createRequest := &CreateInstanceConfigData{
		Name:          "new-instance",
		TenantId:      "tenant-1",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	expectedResponse := CreateInstanceResponse{
		Data: CreateInstanceData{
			Id:            "instance-new",
			Name:          "new-instance",
			TenantId:      "tenant-1",
			CloudProvider: "gcp",
			ConnectionUrl: "neo4j+s://xxxxx.databases.neo4j.io",
			Region:        "us-central1",
			Type:          "enterprise-db",
			Username:      "neo4j",
			Password:      "generated-password",
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

		if r.URL.Path == "/v1/instances" && r.Method == http.MethodPost {
			var req CreateInstanceConfigData
			json.NewDecoder(r.Body).Decode(&req)

			if req.Name != createRequest.Name {
				t.Errorf("expected name '%s', got '%s'", createRequest.Name, req.Name)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Create(createRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Data.Name != "new-instance" {
		t.Errorf("expected name 'new-instance', got '%s'", result.Data.Name)
	}
	if result.Data.Password == "" {
		t.Error("expected password to be populated")
	}
}

// TestInstanceService_Delete_Success verifies instance deletion
func TestInstanceService_Delete_Success(t *testing.T) {
	instanceID := "aaaa1234"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetInstanceResponse{
				Data: GetInstanceData{
					Id:     instanceID,
					Status: "destroying",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Delete(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Data.Status != "destroying" {
		t.Errorf("expected status 'destroying', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Pause_Success verifies instance pausing
func TestInstanceService_Pause_Success(t *testing.T) {
	instanceID := "bbbb5678"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/pause" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetInstanceResponse{
				Data: GetInstanceData{
					Id:     instanceID,
					Status: "pausing",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Pause(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data.Status != "pausing" {
		t.Errorf("expected status 'pausing', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Resume_Success verifies instance resuming
func TestInstanceService_Resume_Success(t *testing.T) {
	instanceID := "bbbb1234"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/resume" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetInstanceResponse{
				Data: GetInstanceData{
					Id:     instanceID,
					Status: "resuming",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Resume(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data.Status != "resuming" {
		t.Errorf("expected status 'resuming', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Update_Success verifies instance updates
func TestInstanceService_Update_Success(t *testing.T) {
	instanceID := "f1f1b2b2"
	updateRequest := &UpdateInstanceData{
		Name:   "updated-name",
		Memory: "16GB",
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

		if r.URL.Path == "/v1/instances/"+instanceID && r.Method == http.MethodPatch {
			var req UpdateInstanceData
			json.NewDecoder(r.Body).Decode(&req)

			if req.Name != updateRequest.Name {
				t.Errorf("expected name '%s', got '%s'", updateRequest.Name, req.Name)
			}
			if req.Memory != updateRequest.Memory {
				t.Errorf("expected memory '%s', got '%s'", updateRequest.Memory, req.Memory)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetInstanceResponse{
				Data: GetInstanceData{
					Id:     instanceID,
					Name:   req.Name,
					Memory: req.Memory,
					Status: "updating",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Update(instanceID, updateRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data.Name != "updated-name" {
		t.Errorf("expected name 'updated-name', got '%s'", result.Data.Name)
	}
	if result.Data.Memory != "16GB" {
		t.Errorf("expected memory '16GB', got '%s'", result.Data.Memory)
	}
}

// TestInstanceService_Overwrite_WithSourceInstance verifies overwrite with source instance
func TestInstanceService_Overwrite_WithSourceInstance(t *testing.T) {
	instanceID := "c1c1c2c2"
	sourceInstanceID := "f1f1f2f2"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/overwrite" && r.Method == http.MethodPost {
			var req overwriteInstance
			json.NewDecoder(r.Body).Decode(&req)

			if req.SourceInstanceId != sourceInstanceID {
				t.Errorf("expected source instance '%s', got '%s'", sourceInstanceID, req.SourceInstanceId)
			}
			if req.SourceSnapshotId != "" {
				t.Error("expected source snapshot to be empty")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(OverwriteInstanceResponse{
				Data: "overwrite-job-123",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Overwrite(instanceID, sourceInstanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data == "" {
		t.Error("expected job ID to be populated")
	}
}

// TestInstanceService_Overwrite_WithSnapshot verifies overwrite with snapshot
// Note: The current implementation requires both instanceID and sourceInstanceID to be valid.
// This test verifies overwrite with both source instance and snapshot provided.
func TestInstanceService_Overwrite_WithSnapshot(t *testing.T) {
	instanceID := "aaaa5678"
	sourceInstanceID := "bbbb1234"
	snapshotID := "snapshot-123"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances/"+instanceID+"/overwrite" && r.Method == http.MethodPost {
			var req overwriteInstance
			json.NewDecoder(r.Body).Decode(&req)

			if req.SourceSnapshotId != snapshotID {
				t.Errorf("expected snapshot '%s', got '%s'", snapshotID, req.SourceSnapshotId)
			}
			if req.SourceInstanceId != sourceInstanceID {
				t.Errorf("expected source instance '%s', got '%s'", sourceInstanceID, req.SourceInstanceId)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(OverwriteInstanceResponse{
				Data: "overwrite-job-456",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.Overwrite(instanceID, sourceInstanceID, snapshotID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data == "" {
		t.Error("expected job ID to be populated")
	}
}

// TestInstanceService_AuthenticationError verifies auth error handling
func TestInstanceService_AuthenticationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	_, err := client.Instances.List()

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

// TestInstanceService_List_EmptyResult verifies empty list handling
func TestInstanceService_List_EmptyResult(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/instances" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ListInstancesResponse{
				Data: []ListInstanceData{},
			})
			return
		}
	}

	client, server := setupInstanceTestClient(handler)
	defer server.Close()

	result, err := client.Instances.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 instances, got %d", len(result.Data))
	}
}

// TestInstanceService_ContextCancellation verifies context handling
func TestInstanceService_ContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a proper client with a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	client, _ := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(10*time.Second),
		WithContext(ctx),
	)

	// Update the baseURL to point to our test server
	client.config.baseURL = server.URL + "/"
	if transport, ok := (*client.transport).(*httpClient.HTTPRequestsService); ok {
		transport.BaseURL = server.URL + "/"
	}

	cancel() // Cancel context before request

	_, err := client.Instances.List()

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
