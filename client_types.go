package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// auraAPIVersion is the version of the Aura API this client targets.
// It is intentionally not user-configurable — a new major API version
// will be delivered as a separate module (e.g. aura-api-client/v2).
const auraAPIVersion = "v1"

// AuraAPIClientVersion is the current release version of this library.
// Updated via changie on each release.
const AuraAPIClientVersion = "v1.8.2"

// AuraAPIClient is the main client for interacting with the Neo4j Aura API
type AuraAPIClient struct {
	api    api.RequestService // Handles authenticated API requests
	logger *slog.Logger       // Structured logger

	// Grouped services - using interface types for testability
	Tenants        TenantService
	Instances      InstanceService
	Snapshots      SnapshotService
	Cmek           CmekService
	GraphAnalytics GDSSessionService
	Prometheus     PrometheusService
}

// config holds internal configuration (unexported)
type config struct {
	baseURL      string        // the base URL of the Aura API
	apiTimeout   time.Duration // how long to wait for a response from an Aura API endpoint
	apiRetryMax  int           // the number of retries to attempt
	clientID     string        // client ID used to obtain an OAuth token
	clientSecret string        // client secret used to obtain an OAuth token
}

// Option is a functional option for configuring the AuraAPIClient
type Option func(*options) error

// options holds the configuration that will be applied to the client
type options struct {
	config config
	logger *slog.Logger
}
