package aura

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// Tenants

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
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// List returns all tenants accessible to the authenticated user
func (t *tenantService) List() (*ListTenantsResponse, error) {
	t.logger.DebugContext(t.ctx, "listing tenants")

	resp, err := t.api.Get(t.ctx, "tenants")
	if err != nil {
		t.logger.ErrorContext(t.ctx, "failed to list tenants", slog.String("error", err.Error()))
		return nil, err
	}

	var result ListTenantsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(t.ctx, "failed to unmarshal tenants response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.ctx, "tenants listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get retrieves details for a specific tenant by ID
func (t *tenantService) Get(tenantID string) (*GetTenantResponse, error) {
	t.logger.DebugContext(t.ctx, "getting tenant details", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(t.ctx, "tenants/"+tenantID)
	if err != nil {
		t.logger.ErrorContext(t.ctx, "failed to get tenant details", slog.String("tenantID", tenantID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetTenantResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(t.ctx, "failed to unmarshal tenant response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.ctx, "tenant obtained successfully", slog.String("name", result.Data.Name))
	return &result, nil
}

func (t *tenantService) GetMetrics(tenantID string) (*GetTenantMetricsURLResponse, error) {
	t.logger.DebugContext(t.ctx, "getting tenant prometheus metrics url", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(t.ctx, "tenants/"+tenantID+"/metrics-integration")
	if err != nil {
		t.logger.ErrorContext(t.ctx, "failed to get tenant prometheus metrics url", slog.String("tenantID", tenantID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetTenantMetricsURLResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(t.ctx, "failed to unmarshal tenant metrics url response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.ctx, "tenant metrics url obtained successfully", slog.String("name", result.Data.Endpoint))
	return &result, nil

}
