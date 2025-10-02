package auraAPIClient

import (
	"context"
	"net/http"
)

// Tenant methods
func (t *TenantService) List(ctx context.Context, token *AuthAPIToken) (*ListTenantsResponse, error) {
	endpoint := t.service.auraAPIVersion + "/tenants"
	return makeAuthenticatedRequest[ListTenantsResponse](ctx, t.service, token, endpoint, http.MethodGet, "application/json", nil)
}

func (t *TenantService) Get(ctx context.Context, token *AuthAPIToken, tenantID string) (*GetTenantResponse, error) {
	endpoint := t.service.auraAPIVersion + "/tenants/" + tenantID
	return makeAuthenticatedRequest[GetTenantResponse](ctx, t.service, token, endpoint, http.MethodGet, "application/json", nil)
}
