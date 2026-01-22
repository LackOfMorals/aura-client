package aura

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// mockAPIService is a mock implementation of api.APIRequestService for testing
type mockAPIService struct {
	response   *api.APIResponse
	err        error
	lastMethod string
	lastPath   string
	lastBody   string
}

func (m *mockAPIService) Get(ctx context.Context, endpoint string) (*api.APIResponse, error) {
	m.lastMethod = "GET"
	m.lastPath = endpoint
	return m.response, m.err
}

func (m *mockAPIService) Post(ctx context.Context, endpoint string, body string) (*api.APIResponse, error) {
	m.lastMethod = "POST"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Put(ctx context.Context, endpoint string, body string) (*api.APIResponse, error) {
	m.lastMethod = "PUT"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Patch(ctx context.Context, endpoint string, body string) (*api.APIResponse, error) {
	m.lastMethod = "PATCH"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Delete(ctx context.Context, endpoint string) (*api.APIResponse, error) {
	m.lastMethod = "DELETE"
	m.lastPath = endpoint
	return m.response, m.err
}

// createTestInstanceService creates an instanceService with a mock API service for testing
func createTestInstanceService(mock *mockAPIService) *instanceService {
	return &instanceService{
		api:    mock,
		ctx:    context.Background(),
		logger: testLogger(),
	}
}

// TestInstanceService_List_Success verifies successful instance listing
func TestInstanceService_List_Success(t *testing.T) {
	expectedResponse := ListInstancesResponse{
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

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances" {
		t.Errorf("expected path 'instances', got '%s'", mock.lastPath)
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
}

// TestInstanceService_Get_Success verifies retrieving a specific instance
func TestInstanceService_Get_Success(t *testing.T) {
	instanceID := "aaaa5678"
	expectedResponse := GetInstanceResponse{
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

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Get(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastPath != "instances/"+instanceID {
		t.Errorf("expected path 'instances/%s', got '%s'", instanceID, mock.lastPath)
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

// TestInstanceService_Get_InvalidID verifies validation of instance ID
func TestInstanceService_Get_InvalidID(t *testing.T) {
	tests := []struct {
		name       string
		instanceID string
	}{
		{"empty", ""},
		{"too short", "abc"},
		{"invalid chars", "!@#$%^&*"},
	}

	mock := &mockAPIService{}
	service := createTestInstanceService(mock)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Get(tt.instanceID)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

// TestInstanceService_Get_NotFound verifies 404 handling
func TestInstanceService_Get_NotFound(t *testing.T) {
	instanceID := "aaaaaaaa"
	mock := &mockAPIService{
		err: &api.APIError{
			StatusCode: 404,
			Message:    "Instance not found",
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Get(instanceID)

	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*api.APIError)
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

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Create(createRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "POST" {
		t.Errorf("expected POST method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances" {
		t.Errorf("expected path 'instances', got '%s'", mock.lastPath)
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

	// Verify request body
	var sentRequest CreateInstanceConfigData
	json.Unmarshal([]byte(mock.lastBody), &sentRequest)
	if sentRequest.Name != createRequest.Name {
		t.Errorf("expected sent name '%s', got '%s'", createRequest.Name, sentRequest.Name)
	}
}

// TestInstanceService_Delete_Success verifies instance deletion
func TestInstanceService_Delete_Success(t *testing.T) {
	instanceID := "aaaa1234"

	expectedResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     instanceID,
			Status: "destroying",
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Delete(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "DELETE" {
		t.Errorf("expected DELETE method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID {
		t.Errorf("expected path 'instances/%s', got '%s'", instanceID, mock.lastPath)
	}
	if result.Data.Status != "destroying" {
		t.Errorf("expected status 'destroying', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Pause_Success verifies instance pausing
func TestInstanceService_Pause_Success(t *testing.T) {
	instanceID := "bbbb5678"

	expectedResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     instanceID,
			Status: "pausing",
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Pause(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "POST" {
		t.Errorf("expected POST method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/pause" {
		t.Errorf("expected path 'instances/%s/pause', got '%s'", instanceID, mock.lastPath)
	}
	if result.Data.Status != "pausing" {
		t.Errorf("expected status 'pausing', got '%s'", result.Data.Status)
	}
}

// TestInstanceService_Resume_Success verifies instance resuming
func TestInstanceService_Resume_Success(t *testing.T) {
	instanceID := "bbbb1234"

	expectedResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     instanceID,
			Status: "resuming",
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Resume(instanceID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastPath != "instances/"+instanceID+"/resume" {
		t.Errorf("expected path 'instances/%s/resume', got '%s'", instanceID, mock.lastPath)
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

	expectedResponse := GetInstanceResponse{
		Data: GetInstanceData{
			Id:     instanceID,
			Name:   "updated-name",
			Memory: "16GB",
			Status: "updating",
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Update(instanceID, updateRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "PATCH" {
		t.Errorf("expected PATCH method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID {
		t.Errorf("expected path 'instances/%s', got '%s'", instanceID, mock.lastPath)
	}
	if result.Data.Name != "updated-name" {
		t.Errorf("expected name 'updated-name', got '%s'", result.Data.Name)
	}
	if result.Data.Memory != "16GB" {
		t.Errorf("expected memory '16GB', got '%s'", result.Data.Memory)
	}
}

// TestInstanceService_Overwrite_Success verifies overwrite with source instance
func TestInstanceService_Overwrite_Success(t *testing.T) {
	instanceID := "c1c1c2c2"
	sourceInstanceID := "f1f1f2f2"

	expectedResponse := OverwriteInstanceResponse{
		Data: "overwrite-job-123",
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Overwrite(instanceID, sourceInstanceID, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "POST" {
		t.Errorf("expected POST method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "instances/"+instanceID+"/overwrite" {
		t.Errorf("expected path 'instances/%s/overwrite', got '%s'", instanceID, mock.lastPath)
	}
	if result.Data == "" {
		t.Error("expected job ID to be populated")
	}

	// Verify request body contains source instance
	var sentRequest overwriteInstanceRequest
	json.Unmarshal([]byte(mock.lastBody), &sentRequest)
	if sentRequest.SourceInstanceId != sourceInstanceID {
		t.Errorf("expected source instance '%s', got '%s'", sourceInstanceID, sentRequest.SourceInstanceId)
	}
}

// TestInstanceService_Overwrite_WithSnapshot verifies overwrite with snapshot
func TestInstanceService_Overwrite_WithSnapshot(t *testing.T) {
	instanceID := "aaaa5678"
	sourceInstanceID := "bbbb1234"
	snapshotID := "snapshot-123"

	expectedResponse := OverwriteInstanceResponse{
		Data: "overwrite-job-456",
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.Overwrite(instanceID, sourceInstanceID, snapshotID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Data == "" {
		t.Error("expected job ID to be populated")
	}

	// Verify request body contains snapshot
	var sentRequest overwriteInstanceRequest
	json.Unmarshal([]byte(mock.lastBody), &sentRequest)
	if sentRequest.SourceSnapshotId != snapshotID {
		t.Errorf("expected snapshot '%s', got '%s'", snapshotID, sentRequest.SourceSnapshotId)
	}
}

// TestInstanceService_List_EmptyResult verifies empty list handling
func TestInstanceService_List_EmptyResult(t *testing.T) {
	expectedResponse := ListInstancesResponse{
		Data: []ListInstanceData{},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 instances, got %d", len(result.Data))
	}
}

// TestInstanceService_AuthenticationError verifies auth error handling
func TestInstanceService_AuthenticationError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{
			StatusCode: 401,
			Message:    "Invalid credentials",
		},
	}

	service := createTestInstanceService(mock)
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
