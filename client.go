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
	"maps"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// ============================================================================
// Constants and version
// ============================================================================

// auraAPIVersion is the version of the Aura API this client targets.
// It is intentionally not user-configurable — a new major API version
// will be delivered as a separate module (e.g. aura-api-client/v2).
const auraAPIVersion = "v1"

// AuraAPIClientVersion is the version of this library. At runtime it is read
// from the embedded module metadata via debug.ReadBuildInfo so it always
// matches the version that was actually imported. The fallback literal is used
// only in development builds (go run / go test outside a module).
var AuraAPIClientVersion = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "v1.10.0"
}()

// ============================================================================
// Client types
// ============================================================================

// AuraAPIClient is the main client for interacting with the Neo4j Aura API.
//
//nolint:revive // AuraAPIClient is intentional: the package is named aura and the type name is established in v1.
type AuraAPIClient struct {
	api    api.RequestService // Handles authenticated API requests
	logger *slog.Logger       // Structured logger

	// Grouped services — using interface types for testability.
	Tenants        TenantService
	Instances      InstanceService
	Snapshots      SnapshotService
	CMEK           CMEKService
	GraphAnalytics GDSSessionService
	Prometheus     PrometheusService
}

// config holds internal configuration (unexported).
type config struct {
	baseURL        string            // the base URL of the Aura API
	apiVersion     string            // API version path segment (e.g. "v1"); empty means use auraAPIVersion
	apiTimeout     time.Duration     // how long to wait for a response from an Aura API endpoint
	apiRetryMax    int               // the number of retries to attempt
	clientID       string            // client ID used to obtain an OAuth token
	clientSecret   string            // client secret used to obtain an OAuth token
	userAgent      string            // override for the User-Agent header; empty means use the default
	httpClient     *http.Client      // optional custom HTTP client; nil means construct a default one
	defaultHeaders map[string]string // optional headers added to every API request
	prometheusThresholds *HealthThresholds // optional; nil means use DefaultHealthThresholds()
}

// Option is a functional option for configuring the AuraAPIClient.
type Option func(*options) error

// options holds the configuration that will be applied to the client.
type options struct {
	config config
	logger *slog.Logger
}

// ============================================================================
// Constructor and options
// ============================================================================

// defaultOptions returns options with sensible defaults.
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

// WithHTTPClient supplies a custom *http.Client for the underlying transport.
// The client is wrapped with retry logic, but the caller retains responsibility
// for any timeout, transport, TLS, or proxy configuration on the client itself.
// When this option is set, the per-request timeout configured by WithTimeout
// is still applied via context, but the client's own Timeout field is not
// overridden.
//
// Typical use: install OpenTelemetry or other middleware via a custom Transport.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) error {
		if c == nil {
			return errors.New("http client must not be nil")
		}
		o.config.httpClient = c
		return nil
	}
}

// WithUserAgent overrides the default User-Agent header sent on every request.
// The default is "aura-go-client/<version>". Most callers should append their
// own product token to the default rather than replace it; e.g.
// WithUserAgent("my-app/1.0 aura-go-client/v1.10.0") so server-side logs still
// identify the SDK version.
func WithUserAgent(ua string) Option {
	return func(o *options) error {
		if ua == "" {
			return errors.New("user agent must not be empty")
		}
		o.config.userAgent = ua
		return nil
	}
}

// WithDefaultHeaders supplies additional headers to add to every API request.
// Useful for opt-in feature flags or preview headers exposed by the Aura API.
//
// The Content-Type, User-Agent, and Authorization headers cannot be overridden
// via this option — entries with those keys are silently ignored. Use
// WithUserAgent to customise the User-Agent.
func WithDefaultHeaders(headers map[string]string) Option {
	return func(o *options) error {
		if len(headers) == 0 {
			return nil
		}
		// Copy so a later mutation of the caller's map cannot affect us.
		copyMap := make(map[string]string, len(headers))
		maps.Copy(copyMap, headers)
		o.config.defaultHeaders = copyMap
		return nil
	}
}

// WithAPIVersion overrides the Aura API version path segment. Defaults to "v1".
// Use this only when Aura releases a new major version (e.g. "v2") and you
// need staged access before the SDK ships a new module.
func WithAPIVersion(v string) Option {
	return func(o *options) error {
		if v == "" {
			return errors.New("API version must not be empty")
		}
		o.config.apiVersion = v
		return nil
	}
}

// WithPrometheusThresholds configures the warning/critical boundaries used by
// PrometheusService.GetInstanceHealth. Call DefaultHealthThresholds() to get
// the built-in values and adjust only the fields you need.
func WithPrometheusThresholds(t HealthThresholds) Option {
	return func(o *options) error {
		o.config.prometheusThresholds = &t
		return nil
	}
}

// WithInsecureBaseURL overrides the base URL without enforcing HTTPS.
// This is intended for local development and in-process testing only (e.g. httptest.Server).
// Never use this option against a real Aura environment — OAuth tokens and API
// credentials will be transmitted in cleartext over the network.
func WithInsecureBaseURL(baseURL string) Option {
	return func(o *options) error {
		if baseURL == "" {
			return errors.New("base URL must not be empty")
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

	apiVersion := o.config.apiVersion
	if apiVersion == "" {
		apiVersion = auraAPIVersion
	}

	o.logger.Debug("configuration validated",
		slog.String("baseURL", o.config.baseURL),
		slog.String("apiVersion", apiVersion),
		slog.Duration("apiTimeout", o.config.apiTimeout),
	)

	userAgent := o.config.userAgent
	if userAgent == "" {
		userAgent = "aura-go-client/" + AuraAPIClientVersion
	}

	prometheusThresholds := DefaultHealthThresholds()
	if o.config.prometheusThresholds != nil {
		prometheusThresholds = *o.config.prometheusThresholds
	}

	apiSvc := api.NewRequestService(api.Config{
		ClientID:       o.config.clientID,
		ClientSecret:   o.config.clientSecret,
		BaseURL:        o.config.baseURL,
		APIVersion:     apiVersion,
		Timeout:        o.config.apiTimeout,
		MaxRetry:       o.config.apiRetryMax,
		UserAgent:      userAgent,
		HTTPClient:     o.config.httpClient,
		DefaultHeaders: o.config.defaultHeaders,
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
	service.CMEK = &cmekService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "cmekService")),
	}
	service.GraphAnalytics = &gdsSessionService{
		api:     apiSvc,
		timeout: o.config.apiTimeout,
		logger:  clientLogger.With(slog.String("service", "gdsSessionService")),
	}
	service.Prometheus = &prometheusService{
		api:        apiSvc,
		timeout:    o.config.apiTimeout,
		logger:     clientLogger.With(slog.String("service", "prometheusService")),
		thresholds: prometheusThresholds,
	}

	service.logger.Info("Aura API client initialized successfully",
		slog.Int("services", 6),
		slog.String("version", AuraAPIClientVersion),
		slog.String("apiVersion", apiVersion),
	)

	return service, nil
}

// Close drains idle connections from the underlying HTTP connection pool.
// Call it when the client is no longer needed, typically via defer:
//
//	client, err := aura.NewClient(...)
//	if err != nil { ... }
//	defer client.Close()
//
// It is safe to call from any goroutine and may be called more than once.
func (c *AuraAPIClient) Close() {
	c.api.Close()
}
