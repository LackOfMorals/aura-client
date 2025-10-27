package aura

import (
	"context"
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
	Service *AuraAPIActionsService
	logger  *slog.Logger
}

// Lists all of the tenants
func (t *TenantService) List(ctx context.Context) (*ListTenantsResponse, error) {
	t.logger.DebugContext(ctx, "listing tenants")

	// Get or update token if needed
	err := t.Service.authMgr.getToken(ctx, *t.Service.transport)
	if err != nil { // Token process failed
		t.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := t.Service.authMgr.Type + " " + t.Service.authMgr.Token
	endpoint := t.Service.Config.Version + "/tenants"

	t.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[ListTenantsResponse](ctx, *t.Service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		t.logger.ErrorContext(ctx, "failed to list tenants", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "tenants listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}

// Get the details of a tenant
func (t *TenantService) Get(ctx context.Context, tenantID string) (*GetTenantResponse, error) {
	t.logger.DebugContext(ctx, "getting tenant details")

	// Get or update token if needed
	err := t.Service.authMgr.getToken(ctx, *t.Service.transport)
	if err != nil { // Token process failed
		t.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := t.Service.authMgr.Type + " " + t.Service.authMgr.Token
	endpoint := t.Service.Config.Version + "/tenants/" + tenantID

	t.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetTenantResponse](ctx, *t.Service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		t.logger.ErrorContext(ctx, "failed to get tenant details", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "tenant obtained successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}
