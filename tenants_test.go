package aura

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// createTestTenantService creates a tenantService with a mock API service for testing
func createTestTenantService(mock *mockAPIService) *tenantService {
	return &tenantService{
		api:    mock,
		ctx:    context.Background(),
		logger: testLogger(),
	}
}

// TestTenantService_List_Success verifies successful tenant listing
func TestTenantService_List_Success(t *testing.T) {
	expectedResponse := ListTenantsResponse{
		Data: []TenantsResponseData{
			{
				Id:   "tenant-1",
				Name: "Development Team",
			},
			{
				Id:   "tenant-2",
				Name: "Production Team",
			},
			{
				Id:   "tenant-3",
				Name: "Testing Team",
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

	service := createTestTenantService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastMethod != "GET" {
		t.Errorf("expected GET method, got %s", mock.lastMethod)
	}
	if mock.lastPath != "tenants" {
		t.Errorf("expected path 'tenants', got '%s'", mock.lastPath)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 tenants, got %d", len(result.Data))
	}
	if result.Data[0].Id != "tenant-1" {
		t.Errorf("expected first tenant ID 'tenant-1', got '%s'", result.Data[0].Id)
	}
	if result.Data[0].Name != "Development Team" {
		t.Errorf("expected first tenant name 'Development Team', got '%s'", result.Data[0].Name)
	}
}

// TestTenantService_List_EmptyResult verifies empty tenant list
func TestTenantService_List_EmptyResult(t *testing.T) {
	expectedResponse := ListTenantsResponse{
		Data: []TenantsResponseData{},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestTenantService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 tenants, got %d", len(result.Data))
	}
}

// TestTenantService_Get_Success verifies retrieving a specific tenant
func TestTenantService_Get_Success(t *testing.T) {
	tenantID := "tenant-123"
	expectedResponse := GetTenantResponse{
		Data: TenantResponseData{
			Id:   tenantID,
			Name: "Development Team",
			InstanceConfigurations: []TenantInstanceConfiguration{
				{
					CloudProvider: "gcp",
					Region:        "us-central1",
					RegionName:    "Iowa",
					Type:          "enterprise-db",
					Memory:        "8GB",
					Storage:       "256GB",
					Version:       "5",
				},
				{
					CloudProvider: "aws",
					Region:        "us-east-1",
					RegionName:    "N. Virginia",
					Type:          "enterprise-db",
					Memory:        "16GB",
					Storage:       "512GB",
					Version:       "5",
				},
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

	service := createTestTenantService(mock)
	result, err := service.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastPath != "tenants/"+tenantID {
		t.Errorf("expected path 'tenants/%s', got '%s'", tenantID, mock.lastPath)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Data.Id != tenantID {
		t.Errorf("expected tenant ID '%s', got '%s'", tenantID, result.Data.Id)
	}
	if result.Data.Name != "Development Team" {
		t.Errorf("expected tenant name 'Development Team', got '%s'", result.Data.Name)
	}
	if len(result.Data.InstanceConfigurations) != 2 {
		t.Errorf("expected 2 instance configurations, got %d", len(result.Data.InstanceConfigurations))
	}
}

// TestTenantService_Get_InstanceConfigurations verifies instance configuration details
func TestTenantService_Get_InstanceConfigurations(t *testing.T) {
	tenantID := "tenant-123"
	expectedResponse := GetTenantResponse{
		Data: TenantResponseData{
			Id:   tenantID,
			Name: "Test Tenant",
			InstanceConfigurations: []TenantInstanceConfiguration{
				{
					CloudProvider: "gcp",
					Region:        "europe-west2",
					RegionName:    "London",
					Type:          "enterprise-db",
					Memory:        "32GB",
					Storage:       "1024GB",
					Version:       "5",
				},
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

	service := createTestTenantService(mock)
	result, err := service.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := result.Data.InstanceConfigurations[0]

	if config.CloudProvider != "gcp" {
		t.Errorf("expected cloud provider 'gcp', got '%s'", config.CloudProvider)
	}
	if config.Region != "europe-west2" {
		t.Errorf("expected region 'europe-west2', got '%s'", config.Region)
	}
	if config.RegionName != "London" {
		t.Errorf("expected region name 'London', got '%s'", config.RegionName)
	}
	if config.Memory != "32GB" {
		t.Errorf("expected memory '32GB', got '%s'", config.Memory)
	}
}

// TestTenantService_Get_NotFound verifies 404 handling
func TestTenantService_Get_NotFound(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "Tenant not found",
		},
	}

	service := createTestTenantService(mock)
	result, err := service.Get("nonexistent-tenant")

	if err == nil {
		t.Fatal("expected error for non-existent tenant")
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

// TestTenantService_AuthenticationError verifies auth error handling
func TestTenantService_AuthenticationError(t *testing.T) {
	mock := &mockAPIService{
		err: &api.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid credentials",
		},
	}

	service := createTestTenantService(mock)
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

// TestTenantService_Get_NoInstanceConfigurations verifies tenant without configurations
func TestTenantService_Get_NoInstanceConfigurations(t *testing.T) {
	tenantID := "tenant-empty"
	expectedResponse := GetTenantResponse{
		Data: TenantResponseData{
			Id:                     tenantID,
			Name:                   "Empty Tenant",
			InstanceConfigurations: []TenantInstanceConfiguration{},
		},
	}

	responseBody, _ := json.Marshal(expectedResponse)
	mock := &mockAPIService{
		response: &api.APIResponse{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestTenantService(mock)
	result, err := service.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data.InstanceConfigurations) != 0 {
		t.Errorf("expected 0 instance configurations, got %d", len(result.Data.InstanceConfigurations))
	}
}

// TestTenantService_SingleTenant verifies list with single tenant
func TestTenantService_SingleTenant(t *testing.T) {
	expectedResponse := ListTenantsResponse{
		Data: []TenantsResponseData{
			{
				Id:   "tenant-single",
				Name: "Only Tenant",
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

	service := createTestTenantService(mock)
	result, err := service.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 tenant, got %d", len(result.Data))
	}
	if result.Data[0].Id != "tenant-single" {
		t.Errorf("expected tenant ID 'tenant-single', got '%s'", result.Data[0].Id)
	}
}
