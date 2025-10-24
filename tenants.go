package aura

import (
	"context"
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
}

// Lists all of the tenants
func (t *TenantService) List(ctx context.Context) (*ListTenantsResponse, error) {
	// Get or update token if needed
	err := t.Service.authMgr.getToken(ctx, *t.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := t.Service.authMgr.Type + " " + t.Service.authMgr.Token
	endpoint := t.Service.Config.Version + "/tenants"

	return makeAuthenticatedRequest[ListTenantsResponse](ctx, *t.Service.transport, auth, endpoint, http.MethodGet, content, "")
}

// Get the details of a tenant
func (t *TenantService) Get(ctx context.Context, tenantID string) (*GetTenantResponse, error) {
	// Get or update token if needed
	err := t.Service.authMgr.getToken(ctx, *t.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := t.Service.authMgr.Type + " " + t.Service.authMgr.Token
	endpoint := t.Service.Config.Version + "/tenants/" + tenantID

	return makeAuthenticatedRequest[GetTenantResponse](ctx, *t.Service.transport, auth, endpoint, http.MethodGet, content, "")
}
