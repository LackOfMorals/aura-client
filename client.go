// Package aura provides a client for interacting with Neo4j Aura API.
package aura

import (
	"errors"
	"log/slog"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// Core service configuration
type AuraAPIActionsService struct {
	Config    *Config                 // Configuration information
	transport *httpClient.HTTPService // Deals with connectivity over http. Nothing here for users
	authMgr   *authManager            // This will manage auth so it is hidden away from users
	logger    *slog.Logger            // Structured logger for debugging and troubleshooting

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
	logger := slog.Default()

	// Validate required fields
	if cfg.ClientID == "" {
		logger.Error("validation failed", slog.String("reason", "client ID must not be empty"))
		return nil, errors.New("client ID must not be empty")
	}
	if cfg.ClientSecret == "" {
		logger.Error("validation failed", slog.String("reason", "client secret must not be empty"))
		return nil, errors.New("client secret must not be empty")
	}
	if cfg.BaseURL == "" {
		logger.Error("validation failed", slog.String("reason", "base URL must not be empty"))
		return nil, errors.New("base URL must not be empty")
	}
	if cfg.Version == "" {
		logger.Error("validation failed", slog.String("reason", "API version must not be empty"))
		return nil, errors.New("API version must not be empty")
	}
	if cfg.APITimeout <= 0 {
		logger.Error("validation failed", slog.String("reason", "API timeout must be greater than zero"), slog.Duration("timeout", cfg.APITimeout))
		return nil, errors.New("API timeout must be greater than zero")
	}

	logger.Debug("configuration validated",
		slog.String("baseURL", cfg.BaseURL),
		slog.String("version", cfg.Version),
		slog.Duration("apiTimeout", cfg.APITimeout),
	)

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
		logger: logger.With(slog.String("component", "AuraAPIActionsService")),
	}

	// Initialize sub-services with reference to parent and dedicated loggers
	service.Tenants = &TenantService{
		Service: service,
		logger:  service.logger.With(slog.String("service", "TenantService")),
	}
	service.Instances = &InstanceService{
		Service: service,
		logger:  service.logger.With(slog.String("service", "InstanceService")),
	}
	service.Snapshots = &SnapshotService{
		Service: service,
		logger:  service.logger.With(slog.String("service", "SnapshotService")),
	}
	service.Cmek = &CmekService{
		Service: service,
		logger:  service.logger.With(slog.String("service", "CmekService")),
	}
	service.GraphAnalytics = &GDSSessionService{
		Service: service,
		logger:  service.logger.With(slog.String("service", "GDSSessionService")),
	}

	service.logger.Info("Aura API service initialized successfully",
		slog.Int("subServices", 5),
	)

	return service, nil
}
