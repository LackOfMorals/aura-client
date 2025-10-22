package resources

import (
	"context"
	"net/http"

	"github.com/LackOfMorals/aura-client"
)

// TenantService handles tenant operations
type TenantService struct {
	Service *aura.AuraAPIActionsService
}

// Lists all of the tenants
func (t *TenantService) List(ctx context.Context, token *aura.AuthAPIToken) (*aura.ListTenantsResponse, error) {
	endpoint := t.service.AuraAPIVersion + "/tenants"
	return makeAuthenticatedRequest[aura.ListTenantsResponse](ctx, t.service, token, endpoint, http.MethodGet, "application/json", "")
}

// Get the details of a tenant
func (t *TenantService) Get(ctx context.Context, token *aura.AuthAPIToken, tenantID string) (*aura.GetTenantResponse, error) {
	endpoint := t.service.AuraAPIVersion + "/tenants/" + tenantID
	return makeAuthenticatedRequest[aura.GetTenantResponse](ctx, t.service, token, endpoint, http.MethodGet, "application/json", "")
}
