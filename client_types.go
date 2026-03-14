package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// Currently set manually to match changie latest
const AuraAPIClientVersion = "v1.6.2"

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
	baseURL      string        // the base url of the aura api
	version      string        // the version of the aura api to use. Only v1 is supported at this time
	apiTimeout   time.Duration // How long to wait for a response from an aura api endpoint
	apiRetryMax  int           // The number of retries to attempt
	clientID     string        // client id to obtain a token to use the aura api
	clientSecret string        // client secret to obtain a token to use the aura api
}

// Option is a functional option for configuring the AuraAPIClient
type Option func(*options) error

// options holds the configuration that will be applied to the client
type options struct {
	config config
	logger *slog.Logger
}
