package aura

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/api"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Instances

// ListInstancesResponse contains a list of instances in a tenant
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

type overwriteInstanceRequest struct {
	SourceInstanceId string `json:"source_instance_id,omitempty"`
	SourceSnapshotId string `json:"source_snapshot_id,omitempty"`
}

type OverwriteInstanceResponse struct {
	Data string `json:"data"`
}

// instanceService handles instance operations
type instanceService struct {
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// List returns all instances accessible to the authenticated user
func (i *instanceService) List() (*ListInstancesResponse, error) {
	i.logger.DebugContext(i.ctx, "listing instances")

	resp, err := i.api.Get(i.ctx, "instances")
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to list instances", slog.String("error", err.Error()))
		return nil, err
	}

	var result ListInstancesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal instances response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.ctx, "instances listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get retrieves details for a specific instance by ID
func (i *instanceService) Get(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "getting instance details", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	resp, err := i.api.Get(i.ctx, "instances/"+instanceID)
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(i.ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", result.Data.Name), slog.String("status", result.Data.Status))
	return &result, nil
}

// Create provisions a new database instance
func (i *instanceService) Create(instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(i.ctx, "instances", string(body))
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	var result CreateInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal create instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "instance created successfully", slog.String("instanceID", result.Data.Id), slog.String("name", result.Data.Name))
	return &result, nil
}

// Delete removes an instance by ID
func (i *instanceService) Delete(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "deleting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	resp, err := i.api.Delete(i.ctx, "instances/"+instanceID)
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal delete instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Pause suspends an instance by ID
func (i *instanceService) Pause(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "pausing instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	resp, err := i.api.Post(i.ctx, "instances/"+instanceID+"/pause", "")
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal pause instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Resume restarts a paused instance by ID
func (i *instanceService) Resume(instanceID string) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "resuming instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	resp, err := i.api.Post(i.ctx, "instances/"+instanceID+"/resume", "")
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal resume instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Update modifies an instance's configuration
func (i *instanceService) Update(instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "updating instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Patch(i.ctx, "instances/"+instanceID, string(body))
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal update instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", result.Data.Name))
	return &result, nil
}

// Overwrite replaces instance data from another instance or snapshot
func (i *instanceService) Overwrite(instanceID string, sourceInstanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error) {
	i.logger.DebugContext(i.ctx, "overwriting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	if err := utils.ValidateInstanceID(sourceInstanceID); err != nil {
		return nil, err
	}

	requestBody := overwriteInstanceRequest{
		SourceInstanceId: sourceInstanceID,
		SourceSnapshotId: sourceSnapshotID,
	}

	body, err := utils.Marshall(requestBody)
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(i.ctx, "instances/"+instanceID+"/overwrite", string(body))
	if err != nil {
		i.logger.ErrorContext(i.ctx, "failed to overwrite instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result OverwriteInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(i.ctx, "failed to unmarshal overwrite instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(i.ctx, "overwriting instance", slog.String("instanceID", instanceID))
	return &result, nil
}
