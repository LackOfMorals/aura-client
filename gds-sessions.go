package aura

import (
	"log/slog"
	"net/http"
)

// GDS Sessions

type getGDSSessionResponse struct {
	Data []getGDSSessionData `json:"data"`
}

type getGDSSessionData struct {
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
type gDSSessionService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// GDS Session methods

func (g *gDSSessionService) List() (*getGDSSessionResponse, error) {

	g.logger.DebugContext(g.service.config.ctx, "Listing GDS Sessions")

	endpoint := g.service.config.version + "/graph-analytics/sessions"

	resp, err := makeServiceRequest[getGDSSessionResponse](g.service.config.ctx, *g.service.transport, g.service.authMgr, endpoint, http.MethodGet, "", g.logger)
	if err != nil {
		g.logger.ErrorContext(g.service.config.ctx, "failed to list GDS sessions", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.service.config.ctx, "gds sessions listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}
