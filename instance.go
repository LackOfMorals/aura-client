package aura

import (
	"context"
	"net/http"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Instances
// List of instances in a tenant
type ListInstancesResponse struct {
	Data []ListInstanceData `json:"data"`
}

type ListInstanceData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Created       string `json:"created_at"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
}

type CreateInstanceConfigData struct {
	Name          string `json:"name"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Version       string `json:"version"`
	Memory        string `json:"memory"`
}

type CreateInstanceResponse struct {
	Data CreateInstanceData `json:"data"`
}

type CreateInstanceData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	ConnectionUrl string `json:"connection_url"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type UpdateInstanceData struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

type GetInstanceResponse struct {
	Data GetInstanceData `json:"data"`
}

type GetInstanceData struct {
	Id              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	TenantId        string  `json:"tenant_id"`
	CloudProvider   string  `json:"cloud_provider"`
	ConnectionUrl   string  `json:"connection_url"`
	Region          string  `json:"region"`
	Type            string  `json:"type"`
	Memory          string  `json:"memory"`
	Storage         *string `json:"storage"`
	CDCEnrichment   string  `json:"cdc_enrichment_mode"`
	GDSPlugin       bool    `json:"graph_analytics_plugin"`
	MetricsURL      string  `json:"metrics_integration_url"`
	Secondaries     int     `json:"secondaries_count"`
	VectorOptimized bool    `json:"vector_optimized"`
}

// InstanceService handles instance operations
type InstanceService struct {
	Service *AuraAPIActionsService
}

// Instance methods

// List all current instances
func (i *InstanceService) List(ctx context.Context) (*ListInstancesResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances"

	return makeAuthenticatedRequest[ListInstancesResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodGet, content, "")
}

// Get the details of an instance
func (i *InstanceService) Get(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodGet, content, "")
}

func (i *InstanceService) Create(ctx context.Context, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[CreateInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, string(body))
}

func (i *InstanceService) Delete(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodDelete, content, "")
}

func (i *InstanceService) Pause(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID + "/pause"

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, "")
}

func (i *InstanceService) Resume(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID + "/resume"

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, "")
}

func (i *InstanceService) Update(ctx context.Context, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPatch, content, string(body))
}
