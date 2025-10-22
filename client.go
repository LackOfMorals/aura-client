package aura

import (
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
	"github.com/LackOfMorals/aura-client/resources"
)

// Test

type TestCreateInstanceConfigData struct {
	Name          string `json:"name"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Version       string `json:"version"`
	Memory        string `json:"memory"`
}

const (
	BaseURL    = "https://api.neo4j.io/"
	ApiVersion = "v1"
	ApiTimeout = 120 * time.Second
)

// NewAuraAPIActionsService creates a new Aura API service with grouped sub-services
func NewAuraAPIActionsService(id, sec string) *resources.AuraAPIActionsService {

	service := &resources.AuraAPIActionsService{
		AuraAPIBaseURL: BaseURL,
		AuraAPIVersion: ApiVersion,
		AuraAPITimeout: ApiTimeout,
		ClientID:       id,
		ClientSecret:   sec,
		Timeout:        ApiTimeout,
	}

	// Reuse a single HTTP client/service instance with configured base URL and timeout
	service.Http = httpClient.NewHTTPRequestService(service.AuraAPIBaseURL, service.Timeout)

	// Initialize sub-services with reference to parent
	service.Auth = &resources.AuthService{Service: service}
	service.Tenants = &resources.TenantService{Service: service}
	service.Instances = &resources.InstanceService{Service: service}
	service.Snapshots = &resources.SnapshotService{Service: service}
	service.Cmek = &resources.CmekService{Service: service}

	return service
}
