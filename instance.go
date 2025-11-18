package aura

import (
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
	SourceInstanceId string `json:"source_instance_id,omitempty"`
	SourceSnapshotId string `json:"source_snapshot_id,omitempty"`
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
func (i *InstanceService) List() (*ListInstancesResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "listing instances")

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances"

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[ListInstancesResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodGet, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to list instances", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "instances listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil
}

// Get the details of an instance
func (i *InstanceService) Get(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "getting instance details", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodGet, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name), slog.String("status", resp.Data.Status))
	return resp, nil
}

// Creates an instance
func (i *InstanceService) Create(instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[CreateInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance created successfully", slog.String("instanceID", resp.Data.Id), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Deletes an instance identified by it's ID
func (i *InstanceService) Delete(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "deleting instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodDelete),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodDelete, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Pause an instance identified by it's ID
func (i *InstanceService) Pause(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "pausing instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/pause"

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Resumes an instance identified by it's ID
func (i *InstanceService) Resume(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "resuming instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/resume"

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Updates an instance identified by it's ID
func (i *InstanceService) Update(instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "updating instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodPatch),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodPatch, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Overwrites an existing instane identified by it's ID with the contents of another instance using an ondemand snapshot. Alternatively, if the snapshot ID of the other instance is given, that is used instead.
func (i *InstanceService) Overwrite(instanceID string, sourceInstanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "overwriting instance", slog.String("instanceID", instanceID))

	// Get or update token if needed
	err := i.service.authMgr.getToken(i.service.config.ctx, *i.service.transport)
	if err != nil { // Token process failed
		i.logger.ErrorContext(i.service.config.ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
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
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := i.service.authMgr.tokenType + " " + i.service.authMgr.token
	endpoint := i.service.config.version + "/instances/" + instanceID + "/overwrite"

	i.logger.DebugContext(i.service.config.ctx, "making authenticated request",
		slog.String("method", http.MethodPost),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[OverwriteInstanceResponse](i.service.config.ctx, *i.service.transport, auth, endpoint, http.MethodPost, content, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to overwrite instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "overwriting instance", slog.String("instanceID", instanceID))
	return resp, nil
}
