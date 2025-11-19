// Package aura provides a Go client library for the Neo4j Aura API.
//
// The client supports all major Aura API operations including instance management,
// snapshots, tenant operations, and customer-managed encryption keys (CMEK).
//
// Example usage:
//
//	client, err := aura.NewClient(
//	    aura.WithCredentials("client-id", "client-secret"),
//	)

//	if err != nil {
//	    log.Fatal(err)
//	}
//
// instances, err := client.Instances.List()
package aura

import (
	"context"
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
	Tenants        *tenantService
	Instances      *instanceService
	Snapshots      *snapshotService
	Cmek           *cmekService
	GraphAnalytics *gDSSessionService
}

// config holds internal configuration (unexported)
type config struct {
	baseURL      string          // the base url of the aura api
	version      string          // the version of the aura api to use. Only v1 is supported at this time
	apiTimeout   time.Duration   // How long to wait for a response from an aura api endpoint
	clientID     string          // client id to obtain a token to use the aura api
	clientSecret string          // client secret to obtain a token to use the aura api
	ctx          context.Context // context for the client
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
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)

	return &options{
		config: config{
			baseURL:    "https://api.neo4j.io/",
			version:    "v1",
			apiTimeout: 120 * time.Second,
			ctx:        context.Background(),
		},
		logger: slog.New(handler),
	}
}

// WithContext sets the context to use
func WithContext(ctx context.Context) Option {
	return func(o *options) error {
		o.config.ctx = ctx
		return nil
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

	trans := httpClient.NewHTTPRequestService(o.config.baseURL, o.config.apiTimeout, o.logger)

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
			logger:     o.logger,
		},
		logger: o.logger.With(slog.String("component", "AuraAPIClient")),
	}

	// Initialize sub-services
	service.Tenants = &tenantService{
		service: service,
		logger:  service.logger.With(slog.String("service", "tenantService")),
	}
	service.Instances = &instanceService{
		service: service,
		logger:  service.logger.With(slog.String("service", "instanceService")),
	}
	service.Snapshots = &snapshotService{
		service: service,
		logger:  service.logger.With(slog.String("service", "snapshotService")),
	}
	service.Cmek = &cmekService{
		service: service,
		logger:  service.logger.With(slog.String("service", "cmekService")),
	}
	service.GraphAnalytics = &gDSSessionService{
		service: service,
		logger:  service.logger.With(slog.String("service", "gDSSessionService")),
	}

	service.logger.Info("Aura API service initialized successfully",
		slog.Int("subServices", 5),
	)

	return service, nil
}
