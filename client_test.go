package aura

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestNewClient_Success verifies successful client creation with credentials
func TestNewClient_Success(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.config == nil {
		t.Error("expected config to be initialized")
	}
	if client.transport == nil {
		t.Error("expected transport to be initialized")
	}
	if client.authMgr == nil {
		t.Error("expected authMgr to be initialized")
	}
	if client.logger == nil {
		t.Error("expected logger to be initialized")
	}
}

// TestNewClient_SubServicesInitialized verifies all sub-services are created
func TestNewClient_SubServicesInitialized(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.Tenants == nil {
		t.Error("expected Tenants service to be initialized")
	}
	if client.Instances == nil {
		t.Error("expected Instances service to be initialized")
	}
	if client.Snapshots == nil {
		t.Error("expected Snapshots service to be initialized")
	}
	if client.Cmek == nil {
		t.Error("expected Cmek service to be initialized")
	}
	if client.GraphAnalytics == nil {
		t.Error("expected GraphAnalytics service to be initialized")
	}
}

// TestNewClient_AuthManagerInitialized verifies auth manager setup
func TestNewClient_AuthManagerInitialized(t *testing.T) {
	clientID := "my-client-id"
	clientSecret := "my-client-secret"

	client, err := NewClient(
		WithCredentials(clientID, clientSecret),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.authMgr.id != clientID {
		t.Errorf("expected authMgr.id to be '%s', got '%s'", clientID, client.authMgr.id)
	}
	if client.authMgr.secret != clientSecret {
		t.Errorf("expected authMgr.secret to be '%s', got '%s'", clientSecret, client.authMgr.secret)
	}
	if client.authMgr.token != "" {
		t.Error("expected authMgr.token to be empty initially")
	}
	if client.authMgr.expiresAt != 0 {
		t.Error("expected authMgr.expiresAt to be 0 initially")
	}
}

// TestNewClient_EmptyClientID validates error for missing client ID
func TestNewClient_EmptyClientID(t *testing.T) {
	client, err := NewClient(
		WithClientID(""),
		WithClientSecret("test-secret"),
	)

	if err == nil {
		t.Error("expected error for empty client ID, got nil")
	}
	if err.Error() != "client ID must not be empty" {
		t.Errorf("expected error message 'client ID must not be empty', got '%s'", err.Error())
	}
	if client != nil {
		t.Error("expected client to be nil when validation fails")
	}
}

// TestNewClient_EmptyClientSecret validates error for missing client secret
func TestNewClient_EmptyClientSecret(t *testing.T) {
	client, err := NewClient(
		WithClientID("test-id"),
		WithClientSecret(""),
	)

	if err == nil {
		t.Error("expected error for empty client secret, got nil")
	}
	if err.Error() != "client secret must not be empty" {
		t.Errorf("expected error message 'client secret must not be empty', got '%s'", err.Error())
	}
	if client != nil {
		t.Error("expected client to be nil when validation fails")
	}
}

// TestNewClient_EmptyCredentials validates both credentials must be provided
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
			client, err := NewClient(
				WithClientID(tt.clientID),
				WithClientSecret(tt.clientSecret),
			)

			if err == nil {
				t.Error("expected error, got nil")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
			if client != nil {
				t.Error("expected client to be nil")
			}
		})
	}
}

// TestWithTimeout_Valid verifies custom timeout configuration
func TestWithTimeout_Valid(t *testing.T) {
	customTimeout := 60 * time.Second

	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(customTimeout),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.config.apiTimeout != customTimeout {
		t.Errorf("expected timeout %v, got %v", customTimeout, client.config.apiTimeout)
	}
}

// TestWithTimeout_Zero validates error for zero timeout
func TestWithTimeout_Zero(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(0),
	)

	if err == nil {
		t.Error("expected error for zero timeout, got nil")
	}
	if err.Error() != "timeout must be greater than zero" {
		t.Errorf("expected timeout error, got '%s'", err.Error())
	}
	if client != nil {
		t.Error("expected client to be nil")
	}
}

// TestWithTimeout_Negative validates error for negative timeout
func TestWithTimeout_Negative(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithTimeout(-10*time.Second),
	)

	if err == nil {
		t.Error("expected error for negative timeout, got nil")
	}
	if client != nil {
		t.Error("expected client to be nil")
	}
}

// TestWithContext verifies custom context configuration
func TestWithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithContext(ctx),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify context is stored
	if client.config.ctx != ctx {
		t.Error("expected custom context to be stored")
	}
}

// TestWithLogger_Valid verifies custom logger configuration
func TestWithLogger_Valid(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	customLogger := slog.New(handler)

	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithLogger(customLogger),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.logger == nil {
		t.Error("expected logger to be set")
	}
}

// TestWithLogger_Nil validates error for nil logger
func TestWithLogger_Nil(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
		WithLogger(nil),
	)

	if err == nil {
		t.Error("expected error for nil logger, got nil")
	}
	if err.Error() != "logger cannot be nil" {
		t.Errorf("expected logger error, got '%s'", err.Error())
	}
	if client != nil {
		t.Error("expected client to be nil")
	}
}

// TestDefaultOptions verifies default configuration values
func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if opts.config.baseURL != "https://api.neo4j.io/" {
		t.Errorf("expected default baseURL 'https://api.neo4j.io/', got '%s'", opts.config.baseURL)
	}
	if opts.config.version != "v1" {
		t.Errorf("expected default version 'v1', got '%s'", opts.config.version)
	}
	if opts.config.apiTimeout != 120*time.Second {
		t.Errorf("expected default timeout 120s, got %v", opts.config.apiTimeout)
	}
	if opts.logger == nil {
		t.Error("expected default logger to be initialized")
	}
	if opts.config.ctx == nil {
		t.Error("expected default context to be initialized")
	}
}

// TestNewClient_MultipleOptions verifies combining multiple options
func TestNewClient_MultipleOptions(t *testing.T) {
	customTimeout := 90 * time.Second
	ctx := context.Background()

	client, err := NewClient(
		WithClientID("test-id"),
		WithClientSecret("test-secret"),
		WithTimeout(customTimeout),
		WithContext(ctx),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.config.clientID != "test-id" {
		t.Errorf("expected clientID 'test-id', got '%s'", client.config.clientID)
	}
	if client.config.clientSecret != "test-secret" {
		t.Errorf("expected clientSecret 'test-secret', got '%s'", client.config.clientSecret)
	}
	if client.config.apiTimeout != customTimeout {
		t.Errorf("expected timeout %v, got %v", customTimeout, client.config.apiTimeout)
	}
	if client.config.ctx != ctx {
		t.Error("expected custom context")
	}
}

// TestNewClient_DefaultValues verifies defaults when options not provided
func TestNewClient_DefaultValues(t *testing.T) {
	client, err := NewClient(
		WithCredentials("test-id", "test-secret"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check defaults
	if client.config.baseURL != "https://api.neo4j.io/" {
		t.Errorf("expected default baseURL, got '%s'", client.config.baseURL)
	}
	if client.config.version != "v1" {
		t.Errorf("expected default version 'v1', got '%s'", client.config.version)
	}
	if client.config.apiTimeout != 120*time.Second {
		t.Errorf("expected default timeout 120s, got %v", client.config.apiTimeout)
	}
}

// TestWithCredentials verifies the convenience method
func TestWithCredentials(t *testing.T) {
	clientID := "test-id"
	clientSecret := "test-secret"

	client, err := NewClient(
		WithCredentials(clientID, clientSecret),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.config.clientID != clientID {
		t.Errorf("expected clientID '%s', got '%s'", clientID, client.config.clientID)
	}
	if client.config.clientSecret != clientSecret {
		t.Errorf("expected clientSecret '%s', got '%s'", clientSecret, client.config.clientSecret)
	}
}
