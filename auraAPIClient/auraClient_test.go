package auraAPIClient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuraAPIActionsService(t *testing.T) {
	baseURL := "https://api.neo4j.io"
	version := "/v1"
	timeout := "120"
	clientID := "test-client-id"
	clientSecret := "test-client-secret"

	service := NewAuraAPIActionsService(baseURL, version, timeout, clientID, clientSecret)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	concreteService, ok := service.(*AuraAPIActionsService)
	if !ok {
		t.Fatal("Expected service to be of type *AuraAPIActionsService")
	}

	if concreteService.AuraAPIBaseURL != baseURL {
		t.Errorf("Expected AuraAPIBaseURL to be %s, got %s", baseURL, concreteService.AuraAPIBaseURL)
	}

	if concreteService.AuraAPIVersion != version {
		t.Errorf("Expected AuraAPIVersion to be %s, got %s", version, concreteService.AuraAPIVersion)
	}

	if concreteService.ClientID != clientID {
		t.Errorf("Expected ClientID to be %s, got %s", clientID, concreteService.ClientID)
	}

	if concreteService.ClientSecret != clientSecret {
		t.Errorf("Expected ClientSecret to be %s, got %s", clientSecret, concreteService.ClientSecret)
	}
}

func TestGetAuthToken_Success(t *testing.T) {
	expectedToken := AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-access-token-12345",
		Expiry: 3600,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		if r.URL.Path != "/oauth/token" {
			t.Errorf("Expected path /oauth/token, got %s", r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded")
		}

		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header to be present")
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedToken)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token, err := service.GetAuthToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == nil {
		t.Fatal("Expected token, got nil")
	}

	if token.Type != expectedToken.Type {
		t.Errorf("Expected token type %s, got %s", expectedToken.Type, token.Type)
	}

	if token.Token != expectedToken.Token {
		t.Errorf("Expected token %s, got %s", expectedToken.Token, token.Token)
	}

	if token.Expiry != expectedToken.Expiry {
		t.Errorf("Expected expiry %d, got %d", expectedToken.Expiry, token.Expiry)
	}
}

func TestGetAuthToken_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client"}`))
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "bad-id", "bad-secret")

	token, err := service.GetAuthToken()
	if err == nil {
		t.Fatal("Expected error for unauthorized, got nil")
	}

	if token != nil {
		t.Errorf("Expected nil token on error, got %v", token)
	}
}

func TestListTenants_Success(t *testing.T) {
	expectedResponse := ListTenantsResponse{
		Data: []TenantsRepostData{
			{
				Id:   "tenant-1",
				Name: "Production Tenant",
			},
			{
				Id:   "tenant-2",
				Name: "Development Tenant",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		if r.URL.Path != "/v1/tenants" {
			t.Errorf("Expected path /v1/tenants, got %s", r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Expected Authorization header")
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.ListTenants(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 tenants, got %d", len(response.Data))
	}

	if response.Data[0].Id != "tenant-1" {
		t.Errorf("Expected tenant ID tenant-1, got %s", response.Data[0].Id)
	}

	if response.Data[0].Name != "Production Tenant" {
		t.Errorf("Expected tenant name 'Production Tenant', got %s", response.Data[0].Name)
	}
}

func TestListTenants_EmptyList(t *testing.T) {
	expectedResponse := ListTenantsResponse{
		Data: []TenantsRepostData{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.ListTenants(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected 0 tenants, got %d", len(response.Data))
	}
}

func TestGetTenant_Success(t *testing.T) {
	expectedResponse := GetTenantResponse{
		Data: TenantRepostData{
			Id:   "tenant-123",
			Name: "My Tenant",
			InstanceConfigurations: []TenantInstanceConfiguration{
				{
					CloudProvider: "gcp",
					Region:        "us-central1",
					RegionName:    "Iowa",
					Type:          "professional-ds",
					Memory:        "8GB",
					Storage:       "16GB",
					Version:       "5",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		if r.URL.Path != "/v1/tenants/tenant-123" {
			t.Errorf("Expected path /v1/tenants/tenant-123, got %s", r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.GetTenant(token, "tenant-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Data.Id != "tenant-123" {
		t.Errorf("Expected tenant ID tenant-123, got %s", response.Data.Id)
	}

	if response.Data.Name != "My Tenant" {
		t.Errorf("Expected tenant name 'My Tenant', got %s", response.Data.Name)
	}

	if len(response.Data.InstanceConfigurations) != 1 {
		t.Errorf("Expected 1 instance configuration, got %d", len(response.Data.InstanceConfigurations))
	}

	config := response.Data.InstanceConfigurations[0]
	if config.CloudProvider != "gcp" {
		t.Errorf("Expected cloud provider gcp, got %s", config.CloudProvider)
	}

	if config.Memory != "8GB" {
		t.Errorf("Expected memory 8GB, got %s", config.Memory)
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "tenant not found"}`))
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.GetTenant(token, "nonexistent-tenant")
	if err == nil {
		t.Fatal("Expected error for not found tenant, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response on error, got %v", response)
	}
}

func TestListInstances_Success(t *testing.T) {
	expectedResponse := ListInstancesResponse{
		Data: []ListInstanceData{
			{
				Id:            "instance-1",
				Name:          "prod-db",
				Created:       "2024-01-15T10:30:00Z",
				TenantId:      "tenant-123",
				CloudProvider: "gcp",
			},
			{
				Id:            "instance-2",
				Name:          "dev-db",
				Created:       "2024-02-20T14:45:00Z",
				TenantId:      "tenant-123",
				CloudProvider: "aws",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		if r.URL.Path != "/v1/instances" {
			t.Errorf("Expected path /v1/instances, got %s", r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		// Verify authorization
		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.ListInstances(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(response.Data))
	}

	if response.Data[0].Id != "instance-1" {
		t.Errorf("Expected instance ID instance-1, got %s", response.Data[0].Id)
	}

	if response.Data[0].Name != "prod-db" {
		t.Errorf("Expected instance name 'prod-db', got %s", response.Data[0].Name)
	}

	if response.Data[0].CloudProvider != "gcp" {
		t.Errorf("Expected cloud provider gcp, got %s", response.Data[0].CloudProvider)
	}

	if response.Data[1].Name != "dev-db" {
		t.Errorf("Expected instance name 'dev-db', got %s", response.Data[1].Name)
	}
}

func TestListInstances_EmptyList(t *testing.T) {
	expectedResponse := ListInstancesResponse{
		Data: []ListInstanceData{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "test-token",
		Expiry: 3600,
	}

	response, err := service.ListInstances(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(response.Data))
	}
}

func TestListInstances_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid or expired token"}`))
	}))
	defer server.Close()

	service := NewAuraAPIActionsService(server.URL, "/v1", "120", "client-id", "client-secret")

	token := &AuthAPIToken{
		Type:   "Bearer",
		Token:  "expired-token",
		Expiry: 0,
	}

	response, err := service.ListInstances(token)
	if err == nil {
		t.Fatal("Expected error for unauthorized, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response on error, got %v", response)
	}
}

func TestAuthAPIToken_Structure(t *testing.T) {
	token := AuthAPIToken{
		Type:   "Bearer",
		Token:  "abc123",
		Expiry: 7200,
	}

	if token.Type != "Bearer" {
		t.Errorf("Expected Type 'Bearer', got %s", token.Type)
	}

	if token.Token != "abc123" {
		t.Errorf("Expected Token 'abc123', got %s", token.Token)
	}

	if token.Expiry != 7200 {
		t.Errorf("Expected Expiry 7200, got %d", token.Expiry)
	}
}

func TestTenantInstanceConfiguration_Structure(t *testing.T) {
	config := TenantInstanceConfiguration{
		CloudProvider: "aws",
		Region:        "us-east-1",
		RegionName:    "N. Virginia",
		Type:          "enterprise-ds",
		Memory:        "16GB",
		Storage:       "32GB",
		Version:       "5",
	}

	if config.CloudProvider != "aws" {
		t.Errorf("Expected CloudProvider 'aws', got %s", config.CloudProvider)
	}

	if config.Memory != "16GB" {
		t.Errorf("Expected Memory '16GB', got %s", config.Memory)
	}

	if config.Version != "5" {
		t.Errorf("Expected Version '5', got %s", config.Version)
	}
}

func TestListInstanceData_Structure(t *testing.T) {
	instance := ListInstanceData{
		Id:            "inst-123",
		Name:          "test-instance",
		Created:       "2024-01-01T00:00:00Z",
		TenantId:      "tenant-456",
		CloudProvider: "gcp",
	}

	if instance.Id != "inst-123" {
		t.Errorf("Expected Id 'inst-123', got %s", instance.Id)
	}

	if instance.Name != "test-instance" {
		t.Errorf("Expected Name 'test-instance', got %s", instance.Name)
	}

	if instance.TenantId != "tenant-456" {
		t.Errorf("Expected TenantId 'tenant-456', got %s", instance.TenantId)
	}
}
