package auraAPIClient

import (
	"context"
	"net/http"
)

// Retrieves information for a Tenant that includes permitted instance configurations
func (a *AuraAPIActionsService) GetTenant(ctx context.Context, token *AuthAPIToken, TenantID string) (*GetTenantResponse, error) {

	endpoint := a.AuraAPIVersion + "/tenants/" + TenantID

	return makeAuthenticatedRequest[GetTenantResponse](ctx, a, token, endpoint, http.MethodGet, nil)

}

// Lists the tenants in the organisation
func (a *AuraAPIActionsService) ListTenants(ctx context.Context, token *AuthAPIToken) (*ListTenantsResponse, error) {

	endpoint := a.AuraAPIVersion + "/tenants"

	return makeAuthenticatedRequest[ListTenantsResponse](ctx, a, token, endpoint, http.MethodGet, nil)

}
