package aura

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestInstanceService creates an instanceService with a mock API service for testing
func createTestInstanceService(mock *mockAPIService) *instanceService {
	return &instanceService{
		api:     mock,
		ctx:     context.Background(),
		timeout: 30 * time.Second,
		logger:  testLogger(),
	}
}

// createTestInstanceServiceWithContext creates an instanceService with custom context
func createTestInstanceServiceWithContext(mock api.RequestService, ctx context.Context, timeout time.Duration) *instanceService {
	return &instanceService{
		api:     mock,
		ctx:     ctx,
		timeout: timeout,
		logger:  testLogger(),
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
		response: &api.Response{
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
		response: &api.Response{
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
		err: &api.Error{
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

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("expected api.Error type, got %T: %v", err, err)
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
		response: &api.Response{
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
		response: &api.Response{
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
		response: &api.Response{
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
		response: &api.Response{
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
		response: &api.Response{
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
		response: &api.Response{
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
	sourceInstanceID := ""
	snapshotID := "snapshot-123"

	expectedResponse := OverwriteInstanceResponse{
		Data: "overwrite-job-456",
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{
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

// TestInstanceService_Overwrite_Validation verifies overwrite validation
func TestInstanceService_Overwrite_Validation(t *testing.T) {
	tests := []struct {
		name             string
		instanceID       string
		sourceInstanceID string
		sourceSnapshotID string
		expectError      bool
		errorContains    string
	}{
		{
			name:             "both sources empty",
			instanceID:       "aaaa1234",
			sourceInstanceID: "",
			sourceSnapshotID: "",
			expectError:      true,
			errorContains:    "must provide either",
		},
		{
			name:             "both sources provided",
			instanceID:       "aaaa1234",
			sourceInstanceID: "bbbb5678",
			sourceSnapshotID: "snapshot-123",
			expectError:      true,
			errorContains:    "cannot provide both",
		},
		{
			name:             "only source instance",
			instanceID:       "aaaa1234",
			sourceInstanceID: "bbbb5678",
			sourceSnapshotID: "",
			expectError:      false,
		},
		{
			name:             "only source snapshot",
			instanceID:       "aaaa1234",
			sourceInstanceID: "",
			sourceSnapshotID: "snapshot-123",
			expectError:      false,
		},
		{
			name:             "invalid source instance ID",
			instanceID:       "aaaa1234",
			sourceInstanceID: "invalid",
			sourceSnapshotID: "",
			expectError:      true,
			errorContains:    "invalid source instance ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseBody, _ := json.Marshal(OverwriteInstanceResponse{Data: "job-123"})
			mock := &mockAPIService{
				response: &api.Response{StatusCode: 200, Body: responseBody},
			}
			service := createTestInstanceService(mock)

			_, err := service.Overwrite(tt.instanceID, tt.sourceInstanceID, tt.sourceSnapshotID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestInstanceService_List_EmptyResult verifies empty list handling
func TestInstanceService_List_EmptyResult(t *testing.T) {
	expectedResponse := ListInstancesResponse{
		Data: []ListInstanceData{},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.Response{
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
		err: &api.Error{
			StatusCode: 401,
			Message:    "Invalid credentials",
		},
	}

	service := createTestInstanceService(mock)
	_, err := service.List()

	if err == nil {
		t.Fatal("expected authentication error")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatal("expected api.Error type")
	}
	if !apiErr.IsUnauthorized() {
		t.Error("expected IsUnauthorized() to be true")
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

// TestInstanceService_List_ContextCancelled verifies immediate cancellation handling
func TestInstanceService_List_ContextCancelled(t *testing.T) {
	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Mock that would succeed if called
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 0,
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context cancelled error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}

	// Should fail very quickly (< 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("cancellation took too long: %v (expected < 100ms)", elapsed)
	}
}

// TestInstanceService_Get_ContextTimeout verifies timeout enforcement
func TestInstanceService_Get_ContextTimeout(t *testing.T) {
	instanceID := "aaaa5678"

	// Mock with 2 second delay
	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: instanceID, Name: "test"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second, // Slow API
	}

	// Service with very short timeout
	service := createTestInstanceServiceWithContext(
		mock,
		context.Background(),
		100*time.Millisecond, // Shorter than delay
	)

	start := time.Now()
	_, err := service.Get(instanceID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout around 100ms, not wait full 2s
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout took too long: %v (expected ~100ms)", elapsed)
	}
}

// TestInstanceService_Create_ParentContextCancellation verifies parent cancellation propagates
func TestInstanceService_Create_ParentContextCancellation(t *testing.T) {
	createRequest := &CreateInstanceConfigData{
		Name:          "test-instance",
		TenantId:      "tenant-1",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	// Parent context that we'll cancel mid-operation
	parentCtx, parentCancel := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(CreateInstanceResponse{
		Data: CreateInstanceData{Id: "new-id", Name: "test-instance"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 1 * time.Second, // Simulates slow create
	}

	service := createTestInstanceServiceWithContext(mock, parentCtx, 30*time.Second)

	// Cancel parent context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		parentCancel()
	}()

	start := time.Now()
	_, err := service.Create(createRequest)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}

	// Should stop around 100ms (when cancelled), not wait full 1s
	if elapsed > 500*time.Millisecond {
		t.Errorf("cancellation took too long: %v (expected ~100ms)", elapsed)
	}

	// Verify mock was called (operation started)
	if mock.callCount == 0 {
		t.Error("expected API to be called")
	}
}

// TestInstanceService_Update_TimeoutRespected verifies per-operation timeout
func TestInstanceService_Update_TimeoutRespected(t *testing.T) {
	instanceID := "aaaa1234"
	updateRequest := &UpdateInstanceData{Name: "new-name", Memory: "16GB"}

	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: instanceID, Name: "new-name"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 500 * time.Millisecond,
	}

	// Parent context has longer timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	// Service timeout is shorter
	service := createTestInstanceServiceWithContext(mock, parentCtx, 200*time.Millisecond)

	start := time.Now()
	_, err := service.Update(instanceID, updateRequest)
	elapsed := time.Since(start)

	// Should timeout at service timeout (200ms), not parent timeout (10s)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout around 200ms
	if elapsed > 400*time.Millisecond {
		t.Errorf("timeout took too long: %v (expected ~200ms)", elapsed)
	}
}

// TestInstanceService_Delete_ConcurrentCancellation tests multiple operations with cancellation
func TestInstanceService_Delete_ConcurrentCancellation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: "test-id", Status: "destroying"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second,
	}

	service := createTestInstanceServiceWithContext(mock, parentCtx, 30*time.Second)

	// Start multiple operations
	done := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			instanceID := "aaaa123" + string(rune('0'+id))
			_, err := service.Delete(instanceID)
			done <- err
		}(i)
	}

	// Cancel after 100ms
	time.Sleep(100 * time.Millisecond)
	parentCancel()

	// All operations should fail with cancellation
	for i := 0; i < 3; i++ {
		err := <-done
		if err == nil {
			t.Error("expected cancellation error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("operation %d: expected context.Canceled, got: %v", i, err)
		}
	}
}

// TestInstanceService_Pause_ContextNotLeaked verifies defer cancel() prevents leaks
func TestInstanceService_Pause_ContextNotLeaked(t *testing.T) {
	instanceID := "bbbb5678"

	// We'll verify this by ensuring the test completes quickly
	// If contexts leaked, they'd accumulate and slow down
	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: instanceID, Status: "pausing"},
	})
	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)

	// Run operation many times
	for i := 0; i < 100; i++ {
		_, err := service.Pause(instanceID)
		if err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}
	}

	// If we got here without memory issues, contexts are being cleaned up
	// This is a basic sanity check - real leak detection would need runtime profiling
}

// TestInstanceService_Resume_QuickCancellation verifies cancellation before API call
func TestInstanceService_Resume_QuickCancellation(t *testing.T) {
	// Context with immediate deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)

	instanceID := "bbbb1234"
	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: instanceID, Status: "resuming"},
	})
	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)

	_, err := service.Resume(instanceID)

	if err == nil {
		t.Fatal("expected deadline exceeded error")
	}

	// Should be deadline exceeded (timeout happened before operation)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context error, got: %v", err)
	}
}

// TestInstanceService_List_ContextPropagation verifies context values propagate
func TestInstanceService_List_ContextPropagation(t *testing.T) {
	// Create context with value
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request-id"), "test-123")

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})

	// Custom mock that checks context
	contextChecked := false
	mock := &mockAPIServiceContextCheck{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		onGet: func(receivedCtx context.Context) {
			// Verify context value is present
			if val := receivedCtx.Value(contextKey("request-id")); val == "test-123" {
				contextChecked = true
			}
		},
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)
	_, err := service.List()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contextChecked {
		t.Error("context value was not propagated to API call")
	}
}

// TestInstanceService_Create_MultipleTimeouts verifies timeout hierarchy
func TestInstanceService_Create_MultipleTimeouts(t *testing.T) {
	createRequest := &CreateInstanceConfigData{
		Name:          "test-instance",
		TenantId:      "tenant-1",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	responseBody, _ := json.Marshal(CreateInstanceResponse{
		Data: CreateInstanceData{Id: "new-id", Name: "test"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 500 * time.Millisecond,
	}

	// Parent context: 5 seconds
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer parentCancel()

	// Service timeout: 100ms (shorter)
	service := createTestInstanceServiceWithContext(mock, parentCtx, 100*time.Millisecond)

	start := time.Now()
	_, err := service.Create(createRequest)
	elapsed := time.Since(start)

	// Should timeout at service timeout (100ms), not parent (5s)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout around 100ms
	if elapsed > 300*time.Millisecond {
		t.Errorf("timeout took too long: %v (expected ~100ms, not 5s)", elapsed)
	}
}

// TestInstanceService_Overwrite_CancellationDuringValidation verifies early cancellation
func TestInstanceService_Overwrite_CancellationDuringValidation(t *testing.T) {
	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	instanceID := "aaaa1234"
	sourceInstanceID := "bbbb5678"

	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       []byte(`{"data":"job-123"}`),
		},
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)

	_, err := service.Overwrite(instanceID, sourceInstanceID, "")

	// Should fail with context error
	if err == nil {
		t.Fatal("expected context error")
	}

	// API should not be called if context already cancelled
	if mock.lastMethod != "" {
		t.Error("API should not be called when context already cancelled")
	}
}

// ============================================================================
// Additional Test Mocks for Context Testing
// ============================================================================

// mockAPIServiceContextCheck is a mock that can verify context propagation
type mockAPIServiceContextCheck struct {
	response *api.Response
	err      error
	onGet    func(context.Context)
}

func (m *mockAPIServiceContextCheck) Get(ctx context.Context, endpoint string) (*api.Response, error) {
	if m.onGet != nil {
		m.onGet(ctx)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.response, m.err
}

func (m *mockAPIServiceContextCheck) Post(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.response, m.err
}

func (m *mockAPIServiceContextCheck) Put(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.response, m.err
}

func (m *mockAPIServiceContextCheck) Patch(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.response, m.err
}

func (m *mockAPIServiceContextCheck) Delete(ctx context.Context, endpoint string) (*api.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.response, m.err
}
