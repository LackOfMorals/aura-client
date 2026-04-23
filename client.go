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
//	instances, err := client.Instances.List(ctx)
package aura

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// defaultOptions returns options with sensible defaults
func defaultOptions() *options {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)

	return &options{
		config: config{
			baseURL:     "https://api.neo4j.io",
			apiTimeout:  120 * time.Second,
			apiRetryMax: 3,
		},
		logger: slog.New(handler),
	}
}

// WithCredentials sets the client ID and secret used for OAuth authentication.
func WithCredentials(clientID, clientSecret string) Option {
	return func(o *options) error {
		o.config.clientID = clientID
		o.config.clientSecret = clientSecret
		return nil
	}
}

// WithTimeout sets a custom API timeout. Defaults to 120 seconds.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return errors.New("timeout must be greater than zero")
		}
		o.config.apiTimeout = timeout
		return nil
	}
}

// WithMaxRetry sets the maximum number of retries for failed requests. Defaults to 3.
func WithMaxRetry(maxRetry int) Option {
	return func(o *options) error {
		if maxRetry <= 0 {
			return errors.New("max retries must be greater than zero")
		}
		o.config.apiRetryMax = maxRetry
		return nil
	}
}

// WithLogger sets a custom slog.Logger. Defaults to warn-level logging to stderr.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		o.logger = logger
		return nil
	}
}

// WithBaseURL overrides the default API base URL. Useful for staging or sandbox environments.
// The URL must use HTTPS to protect OAuth tokens and API credentials in transit.
func WithBaseURL(baseURL string) Option {
	return func(o *options) error {
		if baseURL == "" {
			return errors.New("base URL must not be empty")
		}
		if !strings.HasPrefix(baseURL, "https://") {
			return errors.New("base URL must use HTTPS to protect credentials in transit")
		}
		o.config.baseURL = baseURL
		return nil
	}
}

// NewClient creates a new Aura API client with functional options.
func NewClient(opts ...Option) (*AuraAPIClient, error) {
	o := defaultOptions()

	for _, opt := range opts {
		if err := opt(o); err != nil {
			o.logger.Error("option application failed", slog.String("error", err.Error()))
			return nil, err
		}
	}

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
	if o.config.apiTimeout <= 0 {
		o.logger.Error("validation failed", slog.String("reason", "API timeout must be greater than zero"), slog.Duration("timeout", o.config.apiTimeout))
		return nil, errors.New("API timeout must be greater than zero")
	}

	o.logger.Debug("configuration validated",
		slog.String("baseURL", o.config.baseURL),
		slog.String("apiVersion", auraAPIVersion),
		slog.Duration("apiTimeout", o.config.apiTimeout),
	)

	apiSvc := api.NewRequestService(api.Config{
		ClientID:     o.config.clientID,
		ClientSecret: o.config.clientSecret,
		BaseURL:      o.config.baseURL,
		APIVersion:   auraAPIVersion,
		Timeout:      o.config.apiTimeout,
		MaxRetry:     o.config.apiRetryMax,
		UserAgent:    "aura-go-client/" + AuraAPIClientVersion,
	}, o.logger)

	clientLogger := o.logger.With(slog.String("component", "AuraAPIClient"))

	service := &AuraAPIClient{
		api:    apiSvc,
		logger: clientLogger,
	}

	service.Tenants = &tenantService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "tenantService")),
	}
	service.Instances = &instanceService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "instanceService")),
	}
	service.Snapshots = &snapshotService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "snapshotService")),
	}
	service.Cmek = &cmekService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "cmekService")),
	}
	service.GraphAnalytics = &gDSSessionService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "gDSSessionService")),
	}
	service.Prometheus = &prometheusService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "prometheusService")),
	}

	service.logger.Info("Aura API client initialized successfully",
		slog.Int("services", 6),
		slog.String("version", AuraAPIClientVersion),
		slog.String("apiVersion", auraAPIVersion),
	)

	return service, nil
}
