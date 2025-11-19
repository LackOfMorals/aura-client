package aura

import (
	"log/slog"
	"net/http"
)

// Tenants

// A list of tenants in your organisation, each with summary data
type listTenantsResponse struct {
	Data []tenantsReponseData `json:"data"`
}

type tenantsReponseData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// Details of a tenant
type getTenantResponse struct {
	Data tenantReponseData `json:"data"`
}

type tenantReponseData struct {
	Id                     string                        `json:"id"`
	Name                   string                        `json:"name"`
	InstanceConfigurations []tenantInstanceConfiguration `json:"instance_configurations"`
}

type tenantInstanceConfiguration struct {
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	RegionName    string `json:"region_name"`
	Type          string `json:"type"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
	Version       string `json:"version"`
}

// TenantService handles tenant operations
type tenantService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Lists all of the tenants
func (t *tenantService) List() (*listTenantsResponse, error) {
	t.logger.DebugContext(t.service.config.ctx, "listing tenants")

	endpoint := t.service.config.version + "/tenants"

	t.logger.DebugContext(t.service.config.ctx, "making service request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeServiceRequest[listTenantsResponse](t.service.config.ctx, *t.service.transport, t.service.authMgr, endpoint, http.MethodGet, "", t.logger)
	if err != nil {
		t.logger.ErrorContext(t.service.config.ctx, "failed to list tenants", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.service.config.ctx, "tenants listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}

// Get the details of a tenant
func (t *tenantService) Get(tenantID string) (*getTenantResponse, error) {
	t.logger.DebugContext(t.service.config.ctx, "getting tenant details")

	endpoint := t.service.config.version + "/tenants/" + tenantID

	t.logger.DebugContext(t.service.config.ctx, "making service request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeServiceRequest[getTenantResponse](t.service.config.ctx, *t.service.transport, t.service.authMgr, endpoint, http.MethodGet, "", t.logger)
	if err != nil {
		t.logger.ErrorContext(t.service.config.ctx, "failed to get tenant details", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(t.service.config.ctx, "tenant obtained successfully", slog.String("name : ", resp.Data.Name))
	return resp, nil

}
