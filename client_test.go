package aura

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-secret"

	cfg := DefaultConfig(clientID, clientSecret)

	if cfg.ClientID != clientID {
		t.Errorf("expected ClientID %s, got %s", clientID, cfg.ClientID)
	}
	if cfg.ClientSecret != clientSecret {
		t.Errorf("expected ClientSecret %s, got %s", clientSecret, cfg.ClientSecret)
	}
	if cfg.BaseURL != "https://api.neo4j.io/" {
		t.Errorf("expected BaseURL https://api.neo4j.io/, got %s", cfg.BaseURL)
	}
	if cfg.Version != "v1" {
		t.Errorf("expected Version v1, got %s", cfg.Version)
	}
	if cfg.APITimeout != 120*time.Second {
		t.Errorf("expected APITimeout 120s, got %v", cfg.APITimeout)
	}
}

func TestNewAuraAPIActionsService(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid credentials",
			clientID:     "valid-id",
			clientSecret: "valid-secret",
			expectError:  false,
		},
		{
			name:         "empty client ID",
			clientID:     "",
			clientSecret: "valid-secret",
			expectError:  true,
			errorMsg:     "client ID must not be empty",
		},
		{
			name:         "empty client secret",
			clientID:     "valid-id",
			clientSecret: "",
			expectError:  true,
			errorMsg:     "client secret must not be empty",
		},
		{
			name:         "both credentials empty",
			clientID:     "",
			clientSecret: "",
			expectError:  true,
			errorMsg:     "client ID must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewAuraAPIActionsService(tt.clientID, tt.clientSecret)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if err != nil && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				if service != nil {
					t.Error("expected service to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if service == nil {
					t.Error("expected service to be non-nil")
				}
				if service.ClientID != tt.clientID {
					t.Errorf("expected ClientID %s, got %s", tt.clientID, service.ClientID)
				}
			}
		})
	}
}

func TestNewAuraAPIActionsServiceWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "v1",
				APITimeout:   120 * time.Second,
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
			expectError: false,
		},
		{
			name: "empty client ID",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "v1",
				APITimeout:   120 * time.Second,
				ClientID:     "",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "client ID must not be empty",
		},
		{
			name: "empty client secret",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "v1",
				APITimeout:   120 * time.Second,
				ClientID:     "test-id",
				ClientSecret: "",
			},
			expectError: true,
			errorMsg:    "client secret must not be empty",
		},
		{
			name: "empty base URL",
			config: Config{
				BaseURL:      "",
				Version:      "v1",
				APITimeout:   120 * time.Second,
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "base URL must not be empty",
		},
		{
			name: "empty API version",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "",
				APITimeout:   120 * time.Second,
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "API version must not be empty",
		},
		{
			name: "zero timeout",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "v1",
				APITimeout:   0,
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "API timeout must be greater than zero",
		},
		{
			name: "negative timeout",
			config: Config{
				BaseURL:      "https://api.neo4j.io/",
				Version:      "v1",
				APITimeout:   -1 * time.Second,
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "API timeout must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewAuraAPIActionsServiceWithConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if err != nil && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				if service != nil {
					t.Error("expected service to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if service == nil {
					t.Error("expected service to be non-nil")
				}
				// Verify service was initialized with correct values
				if service.ClientID != tt.config.ClientID {
					t.Errorf("expected ClientID %s, got %s", tt.config.ClientID, service.ClientID)
				}
				if service.Timeout != tt.config.APITimeout {
					t.Errorf("expected Timeout %v, got %v", tt.config.APITimeout, service.Timeout)
				}
			}
		})
	}
}

func TestServiceInitialization(t *testing.T) {
	service, err := NewAuraAPIActionsService("test-id", "test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Verify all sub-services are initialized
	if service.Auth == nil {
		t.Error("Auth service not initialized")
	}
	if service.Tenants == nil {
		t.Error("Tenants service not initialized")
	}
	if service.Instances == nil {
		t.Error("Instances service not initialized")
	}
	if service.Snapshots == nil {
		t.Error("Snapshots service not initialized")
	}
	if service.Cmek == nil {
		t.Error("Cmek service not initialized")
	}
	if service.Http == nil {
		t.Error("Http client not initialized")
	}
}
