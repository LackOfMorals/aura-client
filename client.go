// Package aura provides a client for interacting with Neo4j Aura API.
package aura

import (
	"errors"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// Core service configuration
type AuraAPIActionsService struct {
	Config    *Config                 // Configuration information
	transport *httpClient.HTTPService // Deals with connectivity over http. Nothing here for users
	authMgr   *authManager            // This will manage auth so it is hidden away from users

	// Grouped services
	Tenants        *TenantService
	Instances      *InstanceService
	Snapshots      *SnapshotService
	Cmek           *CmekService
	GraphAnalytics *GDSSessionService
}

// Token management
type authManager struct {
	Id         string `json:"omitempty"`    // the client id
	Secret     string `json:"omitempty"`    // the client secret
	Type       string `json:"token_type"`   // e.g Bearer
	Token      string `json:"access_token"` // the token from aura api auth endpoint
	ObtainedAt int64  `json:"omitempty"`    // The time when the token was obtained in number of seconds since midnight Jan 1st 1970
	ExpiresAt  int64  `json:"expires_in"`   // token duration in seconds
}

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
func NewAuraAPIActionsService(clientID, clientSecret string) (*AuraAPIActionsService, error) {
	return NewAuraAPIActionsServiceWithConfig(DefaultConfig(clientID, clientSecret))
}

// NewAuraAPIActionsServiceWithConfig creates a new Aura API service with custom configuration.
// Returns an error if the configuration is invalid.
func NewAuraAPIActionsServiceWithConfig(cfg Config) (*AuraAPIActionsService, error) {
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

	trans := httpClient.NewHTTPRequestService(cfg.BaseURL, cfg.APITimeout)

	service := &AuraAPIActionsService{
		Config:    &cfg,
		transport: &trans,
		authMgr: &authManager{
			Id:         cfg.ClientID,
			Secret:     cfg.ClientSecret,
			Token:      "",
			Type:       "",
			ExpiresAt:  0,
			ObtainedAt: 0,
		},
	}

	// Initialize sub-services with reference to parent
	service.Tenants = &TenantService{Service: service}
	service.Instances = &InstanceService{Service: service}
	service.Snapshots = &SnapshotService{Service: service}
	service.Cmek = &CmekService{Service: service}
	service.GraphAnalytics = &GDSSessionService{Service: service}

	return service, nil
}
