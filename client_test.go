package aura

import (
	"testing"
	"time"
)

// TestDefaultConfig verifies that DefaultConfig returns expected values
func TestDefaultConfig(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"

	cfg := DefaultConfig(clientID, clientSecret)

	if cfg.BaseURL != "https://api.neo4j.io/" {
		t.Errorf("expected BaseURL to be 'https://api.neo4j.io/', got '%s'", cfg.BaseURL)
	}
	if cfg.version != "v1" {
		t.Errorf("expected Version to be 'v1', got '%s'", cfg.version)
	}
	if cfg.APITimeout != 120*time.Second {
		t.Errorf("expected APITimeout to be 120s, got %v", cfg.APITimeout)
	}
	if cfg.ClientID != clientID {
		t.Errorf("expected ClientID to be '%s', got '%s'", clientID, cfg.ClientID)
	}
	if cfg.ClientSecret != clientSecret {
		t.Errorf("expected ClientSecret to be '%s', got '%s'", clientSecret, cfg.ClientSecret)
	}
}

// TestNewClient_Success verifies successful service creation
func TestNewClient_Success(t *testing.T) {
	service, err := NewClient("test-id", "test-secret")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if service == nil {
		t.Fatal("expected service to be non-nil")
	}
	if service.config == nil {
		t.Error("expected Config to be initialized")
	}
	if service.transport == nil {
		t.Error("expected transport to be initialized")
	}
	if service.authMgr == nil {
		t.Error("expected authMgr to be initialized")
	}
	if service.logger == nil {
		t.Error("expected logger to be initialized")
	}
}

// TestNewClient_SubServicesInitialized verifies all sub-services are created
func TestNewClient_SubServicesInitialized(t *testing.T) {
	service, err := NewClient("test-id", "test-secret")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service.Tenants == nil {
		t.Error("expected Tenants service to be initialized")
	}
	if service.Instances == nil {
		t.Error("expected Instances service to be initialized")
	}
	if service.Snapshots == nil {
		t.Error("expected Snapshots service to be initialized")
	}
	if service.Cmek == nil {
		t.Error("expected Cmek service to be initialized")
	}
	if service.GraphAnalytics == nil {
		t.Error("expected GraphAnalytics service to be initialized")
	}
}

// TestNewClient_AuthManagerInitialized verifies auth manager setup
func TestNewClient_AuthManagerInitialized(t *testing.T) {
	clientID := "my-client-id"
	clientSecret := "my-client-secret"

	service, err := NewClient(clientID, clientSecret)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service.authMgr.Id != clientID {
		t.Errorf("expected authMgr.Id to be '%s', got '%s'", clientID, service.authMgr.Id)
	}
	if service.authMgr.Secret != clientSecret {
		t.Errorf("expected authMgr.Secret to be '%s', got '%s'", clientSecret, service.authMgr.Secret)
	}
	if service.authMgr.Token != "" {
		t.Error("expected authMgr.Token to be empty initially")
	}
	if service.authMgr.ExpiresAt != 0 {
		t.Error("expected authMgr.ExpiresAt to be 0 initially")
	}
}

// TestNewClientWithConfig_EmptyClientID validates error for missing client ID
func TestNewClientWithConfig_EmptyClientID(t *testing.T) {
	cfg := Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "v1",
		APITimeout:   120 * time.Second,
		ClientID:     "", // Empty
		ClientSecret: "test-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for empty client ID, got nil")
	}
	if err.Error() != "client ID must not be empty" {
		t.Errorf("expected error message 'client ID must not be empty', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_EmptyClientSecret validates error for missing client secret
func TestNewClientWithConfig_EmptyClientSecret(t *testing.T) {
	cfg := Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "v1",
		APITimeout:   120 * time.Second,
		ClientID:     "test-id",
		ClientSecret: "", // Empty
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for empty client secret, got nil")
	}
	if err.Error() != "client secret must not be empty" {
		t.Errorf("expected error message 'client secret must not be empty', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_EmptyBaseURL validates error for missing base URL
func TestNewClientWithConfig_EmptyBaseURL(t *testing.T) {
	cfg := Config{
		BaseURL:      "", // Empty
		Version:      "v1",
		APITimeout:   120 * time.Second,
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for empty base URL, got nil")
	}
	if err.Error() != "base URL must not be empty" {
		t.Errorf("expected error message 'base URL must not be empty', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_EmptyVersion validates error for missing version
func TestNewClientWithConfig_EmptyVersion(t *testing.T) {
	cfg := Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "", // Empty
		APITimeout:   120 * time.Second,
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for empty version, got nil")
	}
	if err.Error() != "API version must not be empty" {
		t.Errorf("expected error message 'API version must not be empty', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_ZeroTimeout validates error for zero timeout
func TestNewClientWithConfig_ZeroTimeout(t *testing.T) {
	cfg := Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "v1",
		APITimeout:   0, // Zero
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for zero timeout, got nil")
	}
	if err.Error() != "API timeout must be greater than zero" {
		t.Errorf("expected error message 'API timeout must be greater than zero', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_NegativeTimeout validates error for negative timeout
func TestNewClientWithConfig_NegativeTimeout(t *testing.T) {
	cfg := Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "v1",
		APITimeout:   -10 * time.Second, // Negative
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err == nil {
		t.Error("expected error for negative timeout, got nil")
	}
	if err.Error() != "API timeout must be greater than zero" {
		t.Errorf("expected error message 'API timeout must be greater than zero', got '%s'", err.Error())
	}
	if service != nil {
		t.Error("expected service to be nil when validation fails")
	}
}

// TestNewClientWithConfig_CustomConfig verifies custom config values are used
func TestNewClientWithConfig_CustomConfig(t *testing.T) {
	customTimeout := 60 * time.Second
	cfg := Config{
		BaseURL:      "https://custom.neo4j.io/",
		Version:      "v2",
		APITimeout:   customTimeout,
		ClientID:     "custom-id",
		ClientSecret: "custom-secret",
	}

	service, err := NewClientWithConfig(cfg)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if service.config.BaseURL != cfg.BaseURL {
		t.Errorf("expected BaseURL '%s', got '%s'", cfg.BaseURL, service.config.BaseURL)
	}
	if service.config.version != cfg.version {
		t.Errorf("expected Version '%s', got '%s'", cfg.version, service.config.version)
	}
	if service.config.APITimeout != cfg.APITimeout {
		t.Errorf("expected APITimeout %v, got %v", cfg.APITimeout, service.config.APITimeout)
	}
}

// TestNewClient_EmptyCredentials validates both constructors reject empty credentials
func TestNewClient_EmptyCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		expectedErr  string
	}{
		{
			name:         "both empty",
			clientID:     "",
			clientSecret: "",
			expectedErr:  "client ID must not be empty",
		},
		{
			name:         "empty ID only",
			clientID:     "",
			clientSecret: "secret",
			expectedErr:  "client ID must not be empty",
		},
		{
			name:         "empty secret only",
			clientID:     "id",
			clientSecret: "",
			expectedErr:  "client secret must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewClient(tt.clientID, tt.clientSecret)

			if err == nil {
				t.Error("expected error, got nil")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
			if service != nil {
				t.Error("expected service to be nil")
			}
		})
	}
}
