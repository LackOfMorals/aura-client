package aura

import (
	"log/slog"
	"net/http"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Instances
// List of instances in a tenant
type listInstancesResponse struct {
	Data []listInstanceData `json:"data"`
}

type listInstanceData struct {
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

type createInstanceResponse struct {
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

type getInstanceResponse struct {
	Data getInstanceData `json:"data"`
}

type getInstanceData struct {
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

type overwriteInstance struct {
	SourceInstanceId string `json:"source_instance_id,omitempty"`
	SourceSnapshotId string `json:"source_snapshot_id,omitempty"`
}

type overwriteInstanceResponse struct {
	Data string `json:"data"`
}

// InstanceService handles instance operations
type instanceService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Instance methods

// List all current instances
func (i *instanceService) List() (*listInstancesResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "listing instances")

	endpoint := i.service.config.version + "/instances"

	resp, err := makeServiceRequest[listInstancesResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodGet, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to list instances", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "instances listed successfully", slog.Int("count", len(resp.Data)))
	return resp, nil
}

// Get the details of an instance
func (i *instanceService) Get(instanceID string) (*getInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "getting instance details", slog.String("instanceID", instanceID))

	endpoint := i.service.config.version + "/instances/" + instanceID

	resp, err := makeServiceRequest[getInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodGet, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.service.config.ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name), slog.String("status", resp.Data.Status))
	return resp, nil
}

// Creates an instance
func (i *instanceService) Create(instanceRequest *CreateInstanceConfigData) (*createInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	endpoint := i.service.config.version + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := makeServiceRequest[createInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodPost, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance created successfully", slog.String("instanceID", resp.Data.Id), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Deletes an instance identified by it's ID
func (i *instanceService) Delete(instanceID string) (*getInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "deleting instance", slog.String("instanceID", instanceID))

	endpoint := i.service.config.version + "/instances/" + instanceID

	resp, err := makeServiceRequest[getInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodDelete, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Pause an instance identified by it's ID
func (i *instanceService) Pause(instanceID string) (*getInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "pausing instance", slog.String("instanceID", instanceID))

	endpoint := i.service.config.version + "/instances/" + instanceID + "/pause"

	resp, err := makeServiceRequest[getInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodPost, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Resumes an instance identified by it's ID
func (i *instanceService) Resume(instanceID string) (*getInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "resuming instance", slog.String("instanceID", instanceID))

	endpoint := i.service.config.version + "/instances/" + instanceID + "/resume"

	resp, err := makeServiceRequest[getInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodPost, "", i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return resp, nil
}

// Updates an instance identified by it's ID
func (i *instanceService) Update(instanceID string, instanceRequest *UpdateInstanceData) (*getInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "updating instance", slog.String("instanceID", instanceID))

	endpoint := i.service.config.version + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := makeServiceRequest[getInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodPatch, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", resp.Data.Name))
	return resp, nil
}

// Overwrites an existing instane identified by it's ID with the contents of another instance using an ondemand snapshot. Alternatively, if the snapshot ID of the other instance is given, that is used instead.
func (i *instanceService) Overwrite(instanceID string, sourceInstanceID string, sourceSnapshotID string) (*overwriteInstanceResponse, error) {
	i.logger.DebugContext(i.service.config.ctx, "overwriting instance", slog.String("instanceID", instanceID))

	// create the request body
	// A key will be omitted when empty
	requestBody := overwriteInstance{
		SourceInstanceId: sourceInstanceID,
		SourceSnapshotId: sourceSnapshotID,
	}

	body, err := utils.Marshall(requestBody)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	endpoint := i.service.config.version + "/instances/" + instanceID + "/overwrite"

	resp, err := makeServiceRequest[overwriteInstanceResponse](i.service.config.ctx, *i.service.transport, i.service.authMgr, endpoint, http.MethodPost, string(body), i.logger)
	if err != nil {
		i.logger.ErrorContext(i.service.config.ctx, "failed to overwrite instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.service.config.ctx, "overwriting instance", slog.String("instanceID", instanceID))
	return resp, nil
}
