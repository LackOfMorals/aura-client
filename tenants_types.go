package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// ListTenantsResponse contains a list of tenants in your organisation
type ListTenantsResponse struct {
	Data []TenantsResponseData `json:"data"`
}

type TenantsResponseData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// GetTenantResponse contains details of a tenant
type GetTenantResponse struct {
	Data TenantResponseData `json:"data"`
}

type TenantResponseData struct {
	Id                     string                        `json:"id"`
	Name                   string                        `json:"name"`
	InstanceConfigurations []TenantInstanceConfiguration `json:"instance_configurations"`
}

type TenantInstanceConfiguration struct {
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	RegionName    string `json:"region_name"`
	Type          string `json:"type"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
	Version       string `json:"version"`
}

type GetTenantMetricsURLResponse struct {
	Data GetTenantMetricsURLData `json:"data"`
}

type GetTenantMetricsURLData struct {
	Endpoint string `json:"endpoint"`
}

// tenantService handles tenant operations
type tenantService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
