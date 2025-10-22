package resources

import (
	"context"
	"net/http"
)

// GDSSessionService handles GDS Session operations
type GDSSessionService struct {
	Service *AuraAPIActionsService
}

// GDS Session methods

func (g *GDSSessionService) List(ctx context.Context, token *AuthAPIToken) (*GetGDSSessionResponse, error) {

	endpoint := g.Service.Version + "/graph-analytics/sessions"

	return makeAuthenticatedRequest[GetGDSSessionResponse](ctx, g.Service, token, endpoint, http.MethodGet, "application/json", "")
}
