package aura

import (
	"context"
	"net/http"
)

// GDS Sessions

type GetGDSSessionResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

type GetGDSSessionData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Memory        string `json:"memory"`
	InstanceId    string `json:"instance_id"`
	DatabaseId    string `json:"database_uuid"`
	Status        string `json:"status"`
	Create        string `json:"created_at"`
	Host          string `json:"host"`
	Expiry        string `json:"expiry_date"`
	Ttl           string `json:"ttl"`
	UserId        string `json:"user_id"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
}

// GDSSessionService handles GDS Session operations
type GDSSessionService struct {
	Service *AuraAPIActionsService
}

// GDS Session methods

func (g *GDSSessionService) List(ctx context.Context) (*GetGDSSessionResponse, error) {
	// Get or update token if needed
	err := g.Service.authMgr.getToken(ctx, *g.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := g.Service.authMgr.Type + " " + g.Service.authMgr.Token
	endpoint := g.Service.Config.Version + "/graph-analytics/sessions"

	return makeAuthenticatedRequest[GetGDSSessionResponse](ctx, *g.Service.transport, auth, endpoint, http.MethodGet, content, "")
}
