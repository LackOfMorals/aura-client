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

type OverwriteInstance struct {
	SourceInstanceId string `json:"omitempty source_instance_id"`
	SourceSnapshotId string `json:"omitempty source_snapshot_id"`
}

type OverwriteInstanceResponse struct {
	Data string `json:"data"`
}

// InstanceService handles instance operations
type InstanceService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Instance methods

// List all current instances
func (i *InstanceService) List(ctx context.Context) (*ListInstancesResponse, error) {
	i.logger.DebugContext(ctx, "listing instances")

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[ListInstancesResponse](ctx, *i.service.transport, auth, endpoint, http.MethodGet, content, "", i.logger)
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
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodGet, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name), slog.String("status", resp.Data.Status))
	return resp, nil
}

// Creates an instance
func (i *InstanceService) Create(ctx context.Context, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	i.logger.DebugContext(ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[CreateInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance created successfully", slog.String("instanceID", resp.Data.Id), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Deletes an instance identified by it's ID
func (i *InstanceService) Delete(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "deleting instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodDelete),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodDelete, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Pause an instance identified by it's ID
func (i *InstanceService) Pause(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "pausing instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/pause"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Resumes an instance identified by it's ID
func (i *InstanceService) Resume(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "resuming instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/resume"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Updates an instance identified by it's ID
func (i *InstanceService) Update(ctx context.Context, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	i.logger.DebugContext(ctx, "updating instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPatch),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodPatch, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Overwrites an existing instane identified by it's ID with the contents of another instance using an ondemand snapshot. Alternatively, if the snapshot ID of the other instance is given, that is used instead.
func (i *InstanceService) Overwrite(ctx context.Context, instanceID string, sourceInstanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error) {
	i.logger.DebugContext(ctx, "resuming instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	// create the request body
	// A key will be omitted when empty
	requestBody := OverwriteInstance{
		SourceInstanceId: sourceInstanceID,
		SourceSnapshotId: sourceSnapshotID,
	}

	body, err := utils.Marshall(requestBody)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/overwrite"

	i.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[OverwriteInstanceResponse](ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to overwrite instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "overwriting instance", slog.String("instanceID", instanceID))
	return resp, nil
}
