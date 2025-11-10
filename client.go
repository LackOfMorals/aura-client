// Package aura provides a client for interacting with Neo4j Aura API.
package aura

import (
	"errors"
	"log/slog"
	"os"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// Core service configuration
type AuraAPIClient struct {
	config    *config                 // Internal configuration (unexported)
	transport *httpClient.HTTPService // Deals with connectivity over http
	authMgr   *authManager            // Manages authentication
	logger    *slog.Logger            // Structured logger

	// Grouped services
	Tenants        *TenantService
	Instances      *InstanceService
	Snapshots      *SnapshotService
	Cmek           *CmekService
	GraphAnalytics *GDSSessionService
}

// config holds internal configuration (unexported)
type config struct {
	baseURL      string
	version      string
	apiTimeout   time.Duration
	clientID     string
	clientSecret string
}

// Option is a functional option for configuring the AuraAPIClient
type Option func(*options) error

// options holds the configuration that will be applied to the client
type options struct {
	config config
	logger *slog.Logger
}

// defaultOptions returns options with sensible defaults
func defaultOptions() *options {
	// Enable debug-level logging to stderr
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewTextHandler(os.Stderr, opts)

	return &options{
		config: config{
			baseURL:    "https://api.neo4j.io/",
			version:    "v1",
			apiTimeout: 120 * time.Second,
		},
		logger: slog.New(handler),
	}
}

// WithClientID sets the client ID (required)
func WithClientID(clientID string) Option {
	return func(o *options) error {
		o.config.clientID = clientID
		return nil
	}
}

// WithClientSecret sets the client secret (required)
func WithClientSecret(clientSecret string) Option {
	return func(o *options) error {
		o.config.clientSecret = clientSecret
		return nil
	}
}

// WithCredentials sets both client ID and secret
func WithCredentials(clientID, clientSecret string) Option {
	return func(o *options) error {
		o.config.clientID = clientID
		o.config.clientSecret = clientSecret
		return nil
	}
}

// WithTimeout sets a custom API timeout
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return errors.New("timeout must be greater than zero")
		}
		o.config.apiTimeout = timeout
		return nil
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		o.logger = logger
		return nil
	}
}

// NewAuraAPIClient creates a new Aura API client with functional options
func NewClient(opts ...Option) (*AuraAPIClient, error) {
	// Start with defaults

	o := defaultOptions()

	// Apply all options
	for _, opt := range opts {
		if err := opt(o); err != nil {
			o.logger.Error("option application failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

	// Validate required fields
	if o.config.clientID == "" {
		o.logger.Error("validation failed", slog.String("reason", "client ID must not be empty"))
		return nil, errors.New("client ID must not be empty")
	}
	if o.config.clientSecret == "" {
		o.logger.Error("validation failed", slog.String("reason", "client secret must not be empty"))
		return nil, errors.New("client secret must not be empty")
	}
	if o.config.baseURL == "" {
		o.logger.Error("validation failed", slog.String("reason", "base URL must not be empty"))
		return nil, errors.New("base URL must not be empty")
	}
	if o.config.version == "" {
		o.logger.Error("validation failed", slog.String("reason", "API version must not be empty"))
		return nil, errors.New("API version must not be empty")
	}
	if o.config.apiTimeout <= 0 {
		o.logger.Error("validation failed", slog.String("reason", "API timeout must be greater than zero"), slog.Duration("timeout", o.config.apiTimeout))
		return nil, errors.New("API timeout must be greater than zero")
	}

	o.logger.Debug("configuration validated",
		slog.String("baseURL", o.config.baseURL),
		slog.String("version", o.config.version),
		slog.Duration("apiTimeout", o.config.apiTimeout),
	)

	trans := httpClient.NewHTTPRequestService(o.config.baseURL, o.config.apiTimeout)

	service := &AuraAPIClient{
		config:    &o.config,
		transport: &trans,
		authMgr: &authManager{
			id:         o.config.clientID,
			secret:     o.config.clientSecret,
			token:      "",
			tokenType:  "",
			expiresAt:  0,
			obtainedAt: 0,
		},
		logger: o.logger.With(slog.String("component", "AuraAPIClient")),
	}

	// Initialize sub-services
	service.Tenants = &TenantService{
		service: service,
		logger:  service.logger.With(slog.String("service", "TenantService")),
	}
	service.Instances = &InstanceService{
		service: service,
		logger:  service.logger.With(slog.String("service", "InstanceService")),
	}
	service.Snapshots = &SnapshotService{
		service: service,
		logger:  service.logger.With(slog.String("service", "SnapshotService")),
	}
	service.Cmek = &CmekService{
		service: service,
		logger:  service.logger.With(slog.String("service", "CmekService")),
	}
	service.GraphAnalytics = &GDSSessionService{
		service: service,
		logger:  service.logger.With(slog.String("service", "GDSSessionService")),
	}

	service.logger.Info("Aura API service initialized successfully",
		slog.Int("subServices", 5),
	)

	return service, nil
}
