// Package aura provides a client for interacting with Neo4j Aura API.
package aura

import (
	"errors"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
	"github.com/LackOfMorals/aura-client/resources"
)

// Config holds configuration for the Aura API service.
type Config struct {
	BaseURL      string
	Version      string
	APITimeout   time.Duration
	ClientID     string
	ClientSecret string
}

// DefaultConfig returns a Config with sensible defaults for the Aura API.
func DefaultConfig(clientID, clientSecret string) Config {
	return Config{
		BaseURL:      "https://api.neo4j.io/",
		Version:      "v1",
		APITimeout:   120 * time.Second,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// NewAuraAPIActionsService creates a new Aura API service with grouped sub-services.
// It validates credentials and initializes all sub-services with proper configuration.
// Returns an error if credentials are invalid.
func NewAuraAPIActionsService(clientID, clientSecret string) (*resources.AuraAPIActionsService, error) {
	return NewAuraAPIActionsServiceWithConfig(DefaultConfig(clientID, clientSecret))
}

// NewAuraAPIActionsServiceWithConfig creates a new Aura API service with custom configuration.
// Returns an error if the configuration is invalid.
func NewAuraAPIActionsServiceWithConfig(cfg Config) (*resources.AuraAPIActionsService, error) {
	// Validate required fields
	if cfg.ClientID == "" {
		return nil, errors.New("client ID must not be empty")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("client secret must not be empty")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base URL must not be empty")
	}
	if cfg.Version == "" {
		return nil, errors.New("API version must not be empty")
	}
	if cfg.APITimeout <= 0 {
		return nil, errors.New("API timeout must be greater than zero")
	}

	service := &resources.AuraAPIActionsService{
		BaseURL:      cfg.BaseURL,
		Version:      cfg.Version,
		Timeout:      cfg.APITimeout,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}

	// Initialize HTTP client with configured base URL and timeout
	service.Http = httpClient.NewHTTPRequestService(cfg.BaseURL, cfg.APITimeout)

	// Initialize sub-services with reference to parent
	service.Auth = &resources.AuthService{Service: service}
	service.Tenants = &resources.TenantService{Service: service}
	service.Instances = &resources.InstanceService{Service: service}
	service.Snapshots = &resources.SnapshotService{Service: service}
	service.Cmek = &resources.CmekService{Service: service}

	return service, nil
}
