package aura

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Instances

// List returns all instances accessible to the authenticated user
func (i *instanceService) List(ctx context.Context) (*ListInstancesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "listing instances")

	resp, err := i.api.Get(ctx, "instances")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to list instances", slog.String("error", err.Error()))
		return nil, err
	}

	var result ListInstancesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal instances response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "instances listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get retrieves details for a specific instance by ID
func (i *instanceService) Get(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "getting instance details", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Get(ctx, "instances/"+instanceID)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to get instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", result.Data.Name), slog.String("status", result.Data.Status))
	return &result, nil
}

// Create provisions a new database instance
func (i *instanceService) Create(ctx context.Context, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	// Guard against instanceRequest being nil
	if instanceRequest == nil {
		err := errors.New("instanceRequest must not be nil")
		i.logger.ErrorContext(ctx, "instanceRequest must not be nil ", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantId))

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, "instances", string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to create instance", slog.String("name", instanceRequest.Name), slog.String("error", err.Error()))
		return nil, err
	}

	var result CreateInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal create instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance created successfully", slog.String("instanceID", result.Data.Id), slog.String("name", result.Data.Name))
	return &result, nil
}

// Delete removes an instance by ID
func (i *instanceService) Delete(ctx context.Context, instanceID string) (*DeleteInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "deleting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Delete(ctx, "instances/"+instanceID)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to delete instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result DeleteInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal delete instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance deleted successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Pause suspends an instance by ID
func (i *instanceService) Pause(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "pausing instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, "instances/"+instanceID+"/pause", "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to pause instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal pause instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance paused successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Resume restarts a paused instance by ID
func (i *instanceService) Resume(ctx context.Context, instanceID string) (*GetInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "resuming instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, "instances/"+instanceID+"/resume", "")
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to resume instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal resume instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance resumed successfully", slog.String("instanceID", instanceID))
	return &result, nil
}

// Update modifies an instance's configuration
func (i *instanceService) Update(ctx context.Context, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "updating instance", slog.String("instanceID", instanceID))

	// Guard against instanceRequest being nil
	if instanceRequest == nil {
		err := errors.New("instanceRequest must not be nil")
		i.logger.ErrorContext(ctx, "instanceRequest must not be nil ", slog.String("error", err.Error()))
		return nil, err
	}

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Patch(ctx, "instances/"+instanceID, string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to update instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal update instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance updated successfully", slog.String("instanceID", instanceID), slog.String("name", result.Data.Name))
	return &result, nil
}

// Overwrite replaces instance data from another instance or snapshot
func (i *instanceService) Overwrite(ctx context.Context, instanceID string, sourceInstanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "overwriting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	if sourceInstanceID == "" && sourceSnapshotID == "" {
		return nil, fmt.Errorf("must provide either sourceInstanceID or sourceSnapshotID")
	}

	if sourceInstanceID != "" && sourceSnapshotID != "" {
		return nil, fmt.Errorf("cannot provide both sourceInstanceID and sourceSnapshotID")
	}

	if sourceInstanceID != "" {
		if err := utils.ValidateInstanceID(sourceInstanceID); err != nil {
			return nil, fmt.Errorf("invalid source instance ID: %w", err)
		}
	}

	requestBody := overwriteInstanceRequest{
		SourceInstanceId: sourceInstanceID,
		SourceSnapshotId: sourceSnapshotID,
	}

	body, err := utils.Marshall(requestBody)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, "instances/"+instanceID+"/overwrite", string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to overwrite instance", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return nil, err
	}

	var result OverwriteInstanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		i.logger.ErrorContext(ctx, "failed to unmarshal overwrite instance response", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.InfoContext(ctx, "instance overwrite started", slog.String("instanceID", instanceID))
	return &result, nil
}
