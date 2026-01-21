package aura

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupTenantTestClient creates a test client with a mock server
func setupTenantTestClient(handler http.HandlerFunc) (*AuraAPIClient, *httptest.Server) {
	server := httptest.NewServer(handler)

	client, _ := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(10*time.Second),
	)

	client.config.baseURL = server.URL + "/"

	return client, server
}

// TestTenantService_List_Success verifies successful tenant listing
func TestTenantService_List_Success(t *testing.T) {
	expectedTenants := listTenantsResponse{
		Data: []tenantsResponseData{
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

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/tenants" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenants)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
	if result.Data[2].Name != "Testing Team" {
		t.Errorf("expected third tenant name 'Testing Team', got '%s'", result.Data[2].Name)
	}
}

// TestTenantService_List_EmptyResult verifies empty tenant list
func TestTenantService_List_EmptyResult(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/tenants" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(listTenantsResponse{
				Data: []tenantsResponseData{},
			})
			return
		}
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.List()

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
	expectedTenant := getTenantResponse{
		Data: tenantReponseData{
			Id:   tenantID,
			Name: "Development Team",
			InstanceConfigurations: []tenantInstanceConfiguration{
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

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/tenants/"+tenantID && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenant)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
	expectedTenant := getTenantResponse{
		Data: tenantReponseData{
			Id:   tenantID,
			Name: "Test Tenant",
			InstanceConfigurations: []tenantInstanceConfiguration{
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

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/tenants/"+tenantID && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenant)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.Get(tenantID)

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
	if config.Type != "enterprise-db" {
		t.Errorf("expected type 'enterprise-db', got '%s'", config.Type)
	}
	if config.Memory != "32GB" {
		t.Errorf("expected memory '32GB', got '%s'", config.Memory)
	}
	if config.Storage != "1024GB" {
		t.Errorf("expected storage '1024GB', got '%s'", config.Storage)
	}
	if config.Version != "5" {
		t.Errorf("expected version '5', got '%s'", config.Version)
	}
}

// TestTenantService_Get_NotFound verifies 404 handling
func TestTenantService_Get_NotFound(t *testing.T) {
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
			"message": "Tenant not found",
		})
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.Get("nonexistent-tenant")

	if err == nil {
		t.Fatal("expected error for non-existent tenant")
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

// TestTenantService_Get_NoInstanceConfigurations verifies tenant without configurations
func TestTenantService_Get_NoInstanceConfigurations(t *testing.T) {
	tenantID := "tenant-empty"
	expectedTenant := getTenantResponse{
		Data: tenantReponseData{
			Id:                     tenantID,
			Name:                   "Empty Tenant",
			InstanceConfigurations: []tenantInstanceConfiguration{},
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

		if r.URL.Path == "/v1/tenants/"+tenantID && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenant)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data.InstanceConfigurations) != 0 {
		t.Errorf("expected 0 instance configurations, got %d", len(result.Data.InstanceConfigurations))
	}
}

// TestTenantService_Get_MultipleCloudProviders verifies tenant with multiple cloud providers
func TestTenantService_Get_MultipleCloudProviders(t *testing.T) {
	tenantID := "tenant-multi-cloud"
	expectedTenant := getTenantResponse{
		Data: tenantReponseData{
			Id:   tenantID,
			Name: "Multi-Cloud Tenant",
			InstanceConfigurations: []tenantInstanceConfiguration{
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
					Memory:        "8GB",
					Storage:       "256GB",
					Version:       "5",
				},
				{
					CloudProvider: "azure",
					Region:        "eastus",
					RegionName:    "East US",
					Type:          "enterprise-db",
					Memory:        "8GB",
					Storage:       "256GB",
					Version:       "5",
				},
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

		if r.URL.Path == "/v1/tenants/"+tenantID && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenant)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.Get(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data.InstanceConfigurations) != 3 {
		t.Errorf("expected 3 instance configurations, got %d", len(result.Data.InstanceConfigurations))
	}

	// Verify different cloud providers
	providers := make(map[string]bool)
	for _, config := range result.Data.InstanceConfigurations {
		providers[config.CloudProvider] = true
	}

	expectedProviders := []string{"gcp", "aws", "azure"}
	for _, provider := range expectedProviders {
		if !providers[provider] {
			t.Errorf("expected to find cloud provider '%s' in results", provider)
		}
	}
}

// TestTenantService_AuthenticationError verifies auth error handling
func TestTenantService_AuthenticationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	_, err := client.Tenants.List()

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

// TestTenantService_List_ServerError verifies server error handling
func TestTenantService_List_ServerError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Internal server error",
		})
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.List()

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
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

// TestTenantService_SingleTenant verifies list with single tenant
func TestTenantService_SingleTenant(t *testing.T) {
	expectedTenants := listTenantsResponse{
		Data: []tenantsResponseData{
			{
				Id:   "tenant-single",
				Name: "Only Tenant",
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

		if r.URL.Path == "/v1/tenants" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTenants)
			return
		}
	}

	client, server := setupTenantTestClient(handler)
	defer server.Close()

	result, err := client.Tenants.List()

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
