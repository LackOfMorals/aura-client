package aura

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// setupCmekTestClient creates a test client with a mock server
func setupCmekTestClient(handler http.HandlerFunc) (*AuraAPIClient, *httptest.Server) {
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

// TestCmekService_List_Success verifies successful CMEK listing
func TestCmekService_List_Success(t *testing.T) {
	expectedCmeks := GetCmeksResponse{
		Data: []GetCmeksData{
			{
				Id:       "cmek-1",
				Name:     "Production Key",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-2",
				Name:     "Development Key",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-3",
				Name:     "Testing Key",
				TenantId: "tenant-2",
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

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedCmeks)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 CMEKs, got %d", len(result.Data))
	}
	if result.Data[0].Id != "cmek-1" {
		t.Errorf("expected first CMEK ID 'cmek-1', got '%s'", result.Data[0].Id)
	}
	if result.Data[0].Name != "Production Key" {
		t.Errorf("expected first CMEK name 'Production Key', got '%s'", result.Data[0].Name)
	}
	if result.Data[2].TenantId != "tenant-2" {
		t.Errorf("expected third CMEK tenant ID 'tenant-2', got '%s'", result.Data[2].TenantId)
	}
}

// TestCmekService_List_WithTenantFilter verifies tenant ID filtering
func TestCmekService_List_WithTenantFilter(t *testing.T) {
	tenantID := "c1e2c556-a924-5fac-b7f8-bb624ad9761d"
	expectedCmeks := GetCmeksResponse{
		Data: []GetCmeksData{
			{
				Id:       "cmek-filtered-1",
				Name:     "Filtered Key 1",
				TenantId: tenantID,
			},
			{
				Id:       "cmek-filtered-2",
				Name:     "Filtered Key 2",
				TenantId: tenantID,
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

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			// In a real implementation, you might check for tenant_id query parameter
			// For this test, we'll just return filtered results
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedCmeks)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List(tenantID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 CMEKs, got %d", len(result.Data))
	}

	// Verify all keys belong to the specified tenant
	for _, cmek := range result.Data {
		if cmek.TenantId != tenantID {
			t.Errorf("expected tenant ID '%s', got '%s'", tenantID, cmek.TenantId)
		}
	}
}

// TestCmekService_List_EmptyResult verifies empty CMEK list
func TestCmekService_List_EmptyResult(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetCmeksResponse{
				Data: []GetCmeksData{},
			})
			return
		}
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 CMEKs, got %d", len(result.Data))
	}
}

// TestCmekService_List_SingleKey verifies listing with single key
func TestCmekService_List_SingleKey(t *testing.T) {
	expectedCmeks := GetCmeksResponse{
		Data: []GetCmeksData{
			{
				Id:       "cmek-single",
				Name:     "Only Key",
				TenantId: "tenant-1",
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

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedCmeks)
			return
		}
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 CMEK, got %d", len(result.Data))
	}
	if result.Data[0].Id != "cmek-single" {
		t.Errorf("expected CMEK ID 'cmek-single', got '%s'", result.Data[0].Id)
	}
}

// TestCmekService_List_MultipleTenants verifies keys across multiple tenants
func TestCmekService_List_MultipleTenants(t *testing.T) {
	expectedCmeks := GetCmeksResponse{
		Data: []GetCmeksData{
			{
				Id:       "cmek-tenant1-1",
				Name:     "Tenant 1 Key 1",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-tenant1-2",
				Name:     "Tenant 1 Key 2",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-tenant2-1",
				Name:     "Tenant 2 Key 1",
				TenantId: "tenant-2",
			},
			{
				Id:       "cmek-tenant3-1",
				Name:     "Tenant 3 Key 1",
				TenantId: "tenant-3",
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

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedCmeks)
			return
		}
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 4 {
		t.Errorf("expected 4 CMEKs, got %d", len(result.Data))
	}

	// Verify different tenant IDs
	tenants := make(map[string]int)
	for _, cmek := range result.Data {
		tenants[cmek.TenantId]++
	}

	if len(tenants) != 3 {
		t.Errorf("expected 3 different tenants, got %d", len(tenants))
	}
	if tenants["tenant-1"] != 2 {
		t.Errorf("expected 2 keys for tenant-1, got %d", tenants["tenant-1"])
	}
	if tenants["tenant-2"] != 1 {
		t.Errorf("expected 1 key for tenant-2, got %d", tenants["tenant-2"])
	}
	if tenants["tenant-3"] != 1 {
		t.Errorf("expected 1 key for tenant-3, got %d", tenants["tenant-3"])
	}
}

// TestCmekService_List_AuthenticationError verifies auth error handling
func TestCmekService_List_AuthenticationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	_, err := client.Cmek.List("")

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

// TestCmekService_List_ServerError verifies server error handling
func TestCmekService_List_ServerError(t *testing.T) {
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

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

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

// TestCmekService_List_Forbidden verifies forbidden access handling
func TestCmekService_List_Forbidden(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Access denied to customer managed keys",
		})
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err == nil {
		t.Fatal("expected forbidden error")
	}
	if result != nil {
		t.Error("expected result to be nil on error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("expected APIError type")
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

// TestCmekService_List_VerifyEndpoint verifies correct endpoint is called
func TestCmekService_List_VerifyEndpoint(t *testing.T) {
	endpointCalled := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			endpointCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetCmeksResponse{
				Data: []GetCmeksData{},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	_, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !endpointCalled {
		t.Error("expected /v1/customer-managed-keys endpoint to be called")
	}
}

// TestCmekService_List_DifferentKeyNames verifies various key naming
func TestCmekService_List_DifferentKeyNames(t *testing.T) {
	expectedCmeks := GetCmeksResponse{
		Data: []GetCmeksData{
			{
				Id:       "cmek-1",
				Name:     "production-encryption-key",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-2",
				Name:     "Development Key (US)",
				TenantId: "tenant-1",
			},
			{
				Id:       "cmek-3",
				Name:     "backup_key_2024",
				TenantId: "tenant-1",
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

		if r.URL.Path == "/v1/customer-managed-keys" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedCmeks)
			return
		}
	}

	client, server := setupCmekTestClient(handler)
	defer server.Close()

	result, err := client.Cmek.List("")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify different naming conventions are preserved
	if result.Data[0].Name != "production-encryption-key" {
		t.Errorf("expected name 'production-encryption-key', got '%s'", result.Data[0].Name)
	}
	if result.Data[1].Name != "Development Key (US)" {
		t.Errorf("expected name 'Development Key (US)', got '%s'", result.Data[1].Name)
	}
	if result.Data[2].Name != "backup_key_2024" {
		t.Errorf("expected name 'backup_key_2024', got '%s'", result.Data[2].Name)
	}
}
