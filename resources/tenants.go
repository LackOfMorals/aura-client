package resources

import (
	"context"
	"net/http"
)

// TenantService handles tenant operations
type TenantService struct {
	Service *AuraAPIActionsService
}

// Lists all of the tenants
func (t *TenantService) List(ctx context.Context, token *AuthAPIToken) (*ListTenantsResponse, error) {
	endpoint := t.Service.AuraAPIVersion + "/tenants"
	return makeAuthenticatedRequest[ListTenantsResponse](ctx, t.Service, token, endpoint, http.MethodGet, "application/json", "")
}

// Get the details of a tenant
func (t *TenantService) Get(ctx context.Context, token *AuthAPIToken, tenantID string) (*GetTenantResponse, error) {
	endpoint := t.Service.AuraAPIVersion + "/tenants/" + tenantID
	return makeAuthenticatedRequest[GetTenantResponse](ctx, t.Service, token, endpoint, http.MethodGet, "application/json", "")
}
