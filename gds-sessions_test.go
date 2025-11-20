package aura

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupGDSSessionTestClient creates a test client with a mock server
func setupGDSSessionTestClient(handler http.HandlerFunc) (*AuraAPIClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	
	client, _ := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(10*time.Second),
	)
	
	client.config.baseURL = server.URL + "/"
	
	return client, server
}

// TestGDSSessionService_List_Success verifies successful GDS session listing
func TestGDSSessionService_List_Success(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:            "session-1",
				Name:          "analytics-session-1",
				Memory:        "8GB",
				InstanceId:    "instance-1",
				DatabaseId:    "db-uuid-1",
				Status:        "running",
				Create:        "2024-01-01T00:00:00Z",
				Host:          "session1.gds.neo4j.io",
				Expiry:        "2024-01-02T00:00:00Z",
				Ttl:           "24h",
				UserId:        "user-1",
				TenantId:      "tenant-1",
				CloudProvider: "gcp",
				Region:        "us-central1",
			},
			{
				Id:            "session-2",
				Name:          "analytics-session-2",
				Memory:        "16GB",
				InstanceId:    "instance-2",
				DatabaseId:    "db-uuid-2",
				Status:        "stopped",
				Create:        "2024-01-01T12:00:00Z",
				Host:          "session2.gds.neo4j.io",
				Expiry:        "2024-01-02T12:00:00Z",
				Ttl:           "24h",
				UserId:        "user-2",
				TenantId:      "tenant-1",
				CloudProvider: "aws",
				Region:        "us-east-1",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 GDS sessions, got %d", len(result.Data))
	}
	if result.Data[0].Id != "session-1" {
		t.Errorf("expected first session ID 'session-1', got '%s'", result.Data[0].Id)
	}
	if result.Data[0].Name != "analytics-session-1" {
		t.Errorf("expected first session name 'analytics-session-1', got '%s'", result.Data[0].Name)
	}
	if result.Data[1].Memory != "16GB" {
		t.Errorf("expected second session memory '16GB', got '%s'", result.Data[1].Memory)
	}
}

// TestGDSSessionService_List_EmptyResult verifies empty session list
func TestGDSSessionService_List_EmptyResult(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(getGDSSessionResponse{
				Data: []getGDSSessionData{},
			})
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 GDS sessions, got %d", len(result.Data))
	}
}

// TestGDSSessionService_List_SingleSession verifies listing with single session
func TestGDSSessionService_List_SingleSession(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:            "session-single",
				Name:          "only-session",
				Memory:        "32GB",
				InstanceId:    "instance-1",
				DatabaseId:    "db-uuid-1",
				Status:        "running",
				Create:        "2024-01-01T00:00:00Z",
				Host:          "session.gds.neo4j.io",
				Expiry:        "2024-01-08T00:00:00Z",
				Ttl:           "7d",
				UserId:        "user-1",
				TenantId:      "tenant-1",
				CloudProvider: "gcp",
				Region:        "europe-west2",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 GDS session, got %d", len(result.Data))
	}
	if result.Data[0].Id != "session-single" {
		t.Errorf("expected session ID 'session-single', got '%s'", result.Data[0].Id)
	}
}

// TestGDSSessionService_List_MultipleStatuses verifies sessions with different statuses
func TestGDSSessionService_List_MultipleStatuses(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:         "session-1",
				Name:       "running-session",
				Status:     "running",
				InstanceId: "instance-1",
			},
			{
				Id:         "session-2",
				Name:       "stopped-session",
				Status:     "stopped",
				InstanceId: "instance-2",
			},
			{
				Id:         "session-3",
				Name:       "creating-session",
				Status:     "creating",
				InstanceId: "instance-3",
			},
			{
				Id:         "session-4",
				Name:       "failed-session",
				Status:     "failed",
				InstanceId: "instance-4",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 4 {
		t.Errorf("expected 4 GDS sessions, got %d", len(result.Data))
	}

	// Verify different statuses
	statuses := make(map[string]bool)
	for _, session := range result.Data {
		statuses[session.Status] = true
	}

	expectedStatuses := []string{"running", "stopped", "creating", "failed"}
	for _, status := range expectedStatuses {
		if !statuses[status] {
			t.Errorf("expected to find status '%s' in results", status)
		}
	}
}

// TestGDSSessionService_List_DifferentMemorySizes verifies various memory configurations
func TestGDSSessionService_List_DifferentMemorySizes(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:     "session-1",
				Name:   "small-session",
				Memory: "8GB",
			},
			{
				Id:     "session-2",
				Name:   "medium-session",
				Memory: "16GB",
			},
			{
				Id:     "session-3",
				Name:   "large-session",
				Memory: "32GB",
			},
			{
				Id:     "session-4",
				Name:   "xlarge-session",
				Memory: "64GB",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	memorySizes := []string{"8GB", "16GB", "32GB", "64GB"}
	for i, session := range result.Data {
		if session.Memory != memorySizes[i] {
			t.Errorf("expected memory '%s', got '%s'", memorySizes[i], session.Memory)
		}
	}
}

// TestGDSSessionService_List_MultipleCloudProviders verifies sessions across cloud providers
func TestGDSSessionService_List_MultipleCloudProviders(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:            "session-1",
				Name:          "gcp-session",
				CloudProvider: "gcp",
				Region:        "us-central1",
			},
			{
				Id:            "session-2",
				Name:          "aws-session",
				CloudProvider: "aws",
				Region:        "us-east-1",
			},
			{
				Id:            "session-3",
				Name:          "azure-session",
				CloudProvider: "azure",
				Region:        "eastus",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 GDS sessions, got %d", len(result.Data))
	}

	// Verify different cloud providers
	providers := make(map[string]bool)
	for _, session := range result.Data {
		providers[session.CloudProvider] = true
	}

	expectedProviders := []string{"gcp", "aws", "azure"}
	for _, provider := range expectedProviders {
		if !providers[provider] {
			t.Errorf("expected to find cloud provider '%s' in results", provider)
		}
	}
}

// TestGDSSessionService_List_FullSessionDetails verifies all session fields
func TestGDSSessionService_List_FullSessionDetails(t *testing.T) {
	expectedSession := getGDSSessionData{
		Id:            "session-full",
		Name:          "complete-session",
		Memory:        "16GB",
		InstanceId:    "instance-abc123",
		DatabaseId:    "db-uuid-xyz789",
		Status:        "running",
		Create:        "2024-01-15T10:30:00Z",
		Host:          "session-full.gds.neo4j.io",
		Expiry:        "2024-01-22T10:30:00Z",
		Ttl:           "7d",
		UserId:        "user-abc",
		TenantId:      "tenant-xyz",
		CloudProvider: "gcp",
		Region:        "europe-west2",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(getGDSSessionResponse{
				Data: []getGDSSessionData{expectedSession},
			})
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result.Data))
	}

	session := result.Data[0]

	if session.Id != expectedSession.Id {
		t.Errorf("expected ID '%s', got '%s'", expectedSession.Id, session.Id)
	}
	if session.Name != expectedSession.Name {
		t.Errorf("expected name '%s', got '%s'", expectedSession.Name, session.Name)
	}
	if session.Memory != expectedSession.Memory {
		t.Errorf("expected memory '%s', got '%s'", expectedSession.Memory, session.Memory)
	}
	if session.InstanceId != expectedSession.InstanceId {
		t.Errorf("expected instance ID '%s', got '%s'", expectedSession.InstanceId, session.InstanceId)
	}
	if session.DatabaseId != expectedSession.DatabaseId {
		t.Errorf("expected database ID '%s', got '%s'", expectedSession.DatabaseId, session.DatabaseId)
	}
	if session.Status != expectedSession.Status {
		t.Errorf("expected status '%s', got '%s'", expectedSession.Status, session.Status)
	}
	if session.Create != expectedSession.Create {
		t.Errorf("expected create time '%s', got '%s'", expectedSession.Create, session.Create)
	}
	if session.Host != expectedSession.Host {
		t.Errorf("expected host '%s', got '%s'", expectedSession.Host, session.Host)
	}
	if session.Expiry != expectedSession.Expiry {
		t.Errorf("expected expiry '%s', got '%s'", expectedSession.Expiry, session.Expiry)
	}
	if session.Ttl != expectedSession.Ttl {
		t.Errorf("expected TTL '%s', got '%s'", expectedSession.Ttl, session.Ttl)
	}
	if session.UserId != expectedSession.UserId {
		t.Errorf("expected user ID '%s', got '%s'", expectedSession.UserId, session.UserId)
	}
	if session.TenantId != expectedSession.TenantId {
		t.Errorf("expected tenant ID '%s', got '%s'", expectedSession.TenantId, session.TenantId)
	}
	if session.CloudProvider != expectedSession.CloudProvider {
		t.Errorf("expected cloud provider '%s', got '%s'", expectedSession.CloudProvider, session.CloudProvider)
	}
	if session.Region != expectedSession.Region {
		t.Errorf("expected region '%s', got '%s'", expectedSession.Region, session.Region)
	}
}

// TestGDSSessionService_List_AuthenticationError verifies auth error handling
func TestGDSSessionService_List_AuthenticationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	_, err := client.GraphAnalytics.List()

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

// TestGDSSessionService_List_ServerError verifies server error handling
func TestGDSSessionService_List_ServerError(t *testing.T) {
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

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

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

// TestGDSSessionService_List_VerifyEndpoint verifies correct endpoint is called
func TestGDSSessionService_List_VerifyEndpoint(t *testing.T) {
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			endpointCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(getGDSSessionResponse{
				Data: []getGDSSessionData{},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	_, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !endpointCalled {
		t.Error("expected /v1/graph-analytics/sessions endpoint to be called")
	}
}

// TestGDSSessionService_List_DifferentTTLs verifies various TTL values
func TestGDSSessionService_List_DifferentTTLs(t *testing.T) {
	expectedSessions := getGDSSessionResponse{
		Data: []getGDSSessionData{
			{
				Id:   "session-1",
				Name: "short-lived",
				Ttl:  "1h",
			},
			{
				Id:   "session-2",
				Name: "daily",
				Ttl:  "24h",
			},
			{
				Id:   "session-3",
				Name: "weekly",
				Ttl:  "7d",
			},
			{
				Id:   "session-4",
				Name: "monthly",
				Ttl:  "30d",
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

		if r.URL.Path == "/v1/graph-analytics/sessions" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedSessions)
			return
		}
	}

	client, server := setupGDSSessionTestClient(handler)
	defer server.Close()

	result, err := client.GraphAnalytics.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ttls := []string{"1h", "24h", "7d", "30d"}
	for i, session := range result.Data {
		if session.Ttl != ttls[i] {
			t.Errorf("expected TTL '%s', got '%s'", ttls[i], session.Ttl)
		}
	}
}
