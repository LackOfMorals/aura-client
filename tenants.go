package aura

import (
	"log/slog"
	"net/http"
)

// Tenants

// A list of tenants in your organisation, each with summary data
type ListTenantsResponse struct {
	Data []TenantsRepostData `json:"data"`
}

type TenantsRepostData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// Details of a tenant
type GetTenantResponse struct {
	Data TenantRepostData `json:"data"`
}

type TenantRepostData struct {
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

// TenantService handles tenant operations
type TenantService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Lists all of the tenants
func (t *TenantService) List() (*ListTenantsResponse, error) {
	t.logger.DebugContext(t.service.config.ctx, "listing tenants")

	// Get or update token if needed
	err := t.service.authMgr.getToken(t.service.config.ctx, *t.service.transport)
	if err != nil { // Token process failed
		t.logger.ErrorContext(t.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := t.service.authMgr.tokenType + " " + t.service.authMgr.token
	endpoint := t.service.config.version + "/tenants"

	t.logger.DebugContext(t.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[ListTenantsResponse](t.service.config.ctx, *t.service.transport, auth, endpoint, http.MethodGet, content, "", t.logger)
	if err != nil {
		t.logger.ErrorContext(t.service.config.ctx, "failed to list tenants", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.service.config.ctx, "tenants listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}

// Get the details of a tenant
func (t *TenantService) Get(tenantID string) (*GetTenantResponse, error) {
	t.logger.DebugContext(t.service.config.ctx, "getting tenant details")

	// Get or update token if needed
	err := t.service.authMgr.getToken(t.service.config.ctx, *t.service.transport)
	if err != nil { // Token process failed
		t.logger.ErrorContext(t.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := t.service.authMgr.tokenType + " " + t.service.authMgr.token
	endpoint := t.service.config.version + "/tenants/" + tenantID

	t.logger.DebugContext(t.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetTenantResponse](t.service.config.ctx, *t.service.transport, auth, endpoint, http.MethodGet, content, "", t.logger)
	if err != nil {
		t.logger.ErrorContext(t.service.config.ctx, "failed to get tenant details", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.service.config.ctx, "tenant obtained successfully", slog.String("name : ", resp.Data.Name))
	return resp, nil

}
