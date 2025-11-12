package aura

import (
	"log/slog"
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
	service *AuraAPIClient
	logger  *slog.Logger
}

// GDS Session methods

func (g *GDSSessionService) List() (*GetGDSSessionResponse, error) {

	g.logger.DebugContext(g.service.config.ctx, "Listing GDS Sessions")

	// Get or update token if needed
	err := g.service.authMgr.getToken(g.service.config.ctx, *g.service.transport)
	if err != nil { // Token process failed
		g.logger.ErrorContext(g.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := g.service.authMgr.tokenType + " " + g.service.authMgr.token
	endpoint := g.service.config.version + "/graph-analytics/sessions"

	g.logger.DebugContext(g.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetGDSSessionResponse](g.service.config.ctx, *g.service.transport, auth, endpoint, http.MethodGet, content, "", g.logger)
	if err != nil {
		g.logger.ErrorContext(g.service.config.ctx, "failed to list GDS sessions", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.service.config.ctx, "gds sessions listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}
