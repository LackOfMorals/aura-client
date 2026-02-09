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
	"path/filepath"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
)

// Core service configuration
type AuraAPIClient struct {
	api    api.APIRequestService // Handles authenticated API requests
	ctx    context.Context       // Context for API operations
	logger *slog.Logger          // Structured logger

	// Grouped services - using interface types for testability
	Tenants        TenantService
	Instances      InstanceService
	Snapshots      SnapshotService
	Cmek           CmekService
	GraphAnalytics GDSSessionService
	Prometheus     PrometheusService
	Store          StoreService
}

// config holds internal configuration (unexported)
type config struct {
	baseURL      string          // the base url of the aura api
	version      string          // the version of the aura api to use. Only v1 is supported at this time
	apiTimeout   time.Duration   // How long to wait for a response from an aura api endpoint
	apiRetryMax  int             // The number of retries to attempt
	clientID     string          // client id to obtain a token to use the aura api
	clientSecret string          // client secret to obtain a token to use the aura api
	ctx          context.Context // context for the client
	storePath    string          // path to the SQLite database for configuration storage
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

	// Default store path in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultStorePath := filepath.Join(homeDir, ".aura-client", "store.db")

	return &options{
		config: config{
			baseURL:     "https://api.neo4j.io",
			version:     "v1",
			apiTimeout:  120 * time.Second,
			apiRetryMax: 3,
			ctx:         context.Background(),
			storePath:   defaultStorePath,
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

// WithCredentials sets both client ID and secret
func WithCredentials(clientID, clientSecret string) Option {
	return func(o *options) error {
		o.config.clientID = clientID
		o.config.clientSecret = clientSecret
		return nil
	}
}

// WithTimeout sets a custom API timeout (optional)
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return errors.New("timeout must be greater than zero")
		}
		o.config.apiTimeout = timeout
		return nil
	}
}

// WithMaxRetry sets a custom max number of retries (optional)
func WithMaxRetry(maxRetry int) Option {
	return func(o *options) error {
		if maxRetry <= 0 {
			return errors.New("Max retries must be greater than zero")
		}
		o.config.apiRetryMax = maxRetry
		return nil
	}
}

// WithLogger sets a custom logger (optional)
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		o.logger = logger
		return nil
	}
}

// WithStorePath sets a custom path for the SQLite configuration store (optional)
func WithStorePath(path string) Option {
	return func(o *options) error {
		if path == "" {
			return errors.New("store path cannot be empty")
		}
		o.config.storePath = path
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

	// Create the HTTP service (lowest layer)
	httpSvc := httpClient.NewHTTPService(o.config.baseURL, o.config.version, o.config.apiTimeout, o.config.apiRetryMax, o.logger)

	// Create the API request service (middle layer - handles auth)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     o.config.clientID,
		ClientSecret: o.config.clientSecret,
		BaseURL:      o.config.baseURL,
		APIVersion:   o.config.version,
		Timeout:      o.config.apiTimeout,
	}, o.logger)

	service := &AuraAPIClient{
		api:    apiSvc,
		ctx:    o.config.ctx,
		logger: o.logger.With(slog.String("component", "AuraAPIClient")),
	}

	// Initialize sub-services
	service.Tenants = &tenantService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "tenantService")),
	}
	service.Instances = &instanceService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "instanceService")),
	}
	service.Snapshots = &snapshotService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "snapshotService")),
	}
	service.Cmek = &cmekService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "cmekService")),
	}
	service.GraphAnalytics = &gDSSessionService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "gDSSessionService")),
	}
	service.Prometheus = &prometheusService{
		api:    apiSvc,
		ctx:    service.ctx,
		logger: service.logger.With(slog.String("service", "prometheusService")),
	}

	// Initialize store service
	storeSvc, err := newStoreService(
		o.config.ctx,
		o.config.storePath,
		service.logger.With(slog.String("service", "storeService")),
	)
	if err != nil {
		o.logger.Error("failed to initialize store service", slog.String("error", err.Error()))
		return nil, err
	}
	service.Store = storeSvc

	// Set store reference in instance service for CreateFromStore
	if instSvc, ok := service.Instances.(*instanceService); ok {
		instSvc.store = storeSvc
	}

	service.logger.Info("Aura API service initialized successfully",
		slog.Int("subServices", 7),
	)

	return service, nil
}
