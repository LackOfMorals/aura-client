package aura

import (
	"context"
	"log/slog"
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
	logger  *slog.Logger
}

// Instance methods

// List all current instances
func (i *InstanceService) List(ctx context.Context) (*ListInstancesResponse, error) {
	i.logger.DebugContext(ctx, "listing instances")

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[ListInstancesResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to list instances", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "instances listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil
}

// Get the details of an instance
func (i *InstanceService) Get(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "getting instance details", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name), slog.String("status", resp.Data.Status))
	return resp, nil
}

func (i *InstanceService) Create(ctx context.Context, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	i.logger.DebugContext(ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[CreateInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance created successfully", slog.String("instanceID", resp.Data.Id), slog.String("name", resp.Data.Name))
	return resp, nil
}

func (i *InstanceService) Delete(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "deleting instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodDelete),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodDelete, content, "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

func (i *InstanceService) Pause(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "pausing instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID + "/pause"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

func (i *InstanceService) Resume(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "resuming instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID + "/resume"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPost, content, "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

func (i *InstanceService) Update(ctx context.Context, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "updating instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.Service.authMgr.getToken(ctx, *i.Service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.Service.authMgr.Type + " " + i.Service.authMgr.Token
	endpoint := i.Service.Config.Version + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPatch),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.Service.transport, auth, endpoint, http.MethodPatch, content, string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name))
	return resp, nil
}
