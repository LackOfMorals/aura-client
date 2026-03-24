package aura

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Instances
// instanceService handles instance operations
type instanceService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}

// List returns all instances accessible to the authenticated user
func (i *instanceService) List(ctx context.Context) (*ListInstancesResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

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
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

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

	i.logger.DebugContext(ctx, "instance retrieved successfully", slog.String("instanceID", instanceID), slog.String("name", result.Data.Name), slog.Any("status", result.Data.Status))
	return &result, nil
}

// Create provisions a new database instance
func (i *instanceService) Create(ctx context.Context, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	// Guard against instanceRequest being nil
	if instanceRequest == nil {
		err := errors.New("instanceRequest must not be nil")
		i.logger.ErrorContext(ctx, "instanceRequest must not be nil ", slog.String("error", err.Error()))
		return nil, err
	}

	// Check configuration has min number of parameters set
	err := validateCreateInstanceConfig(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to validate instance configuration", slog.String("error", err.Error()))
		return nil, err
	}

	i.logger.DebugContext(ctx, "creating instance", slog.String("name", instanceRequest.Name), slog.String("tenantID", instanceRequest.TenantID))

	body, err := utils.Marshal(instanceRequest)
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

	i.logger.InfoContext(ctx, "instance created successfully", slog.String("instanceID", result.Data.ID), slog.String("name", result.Data.Name))
	return &result, nil
}

// Delete removes an instance by ID
func (i *instanceService) Delete(ctx context.Context, instanceID string) (*DeleteInstanceResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
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
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "pausing instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, fmt.Sprintf("instances/%s/pause", instanceID), "")
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
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "resuming instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, fmt.Sprintf("instances/%s/resume", instanceID), "")
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
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

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

	body, err := utils.Marshal(instanceRequest)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Patch(ctx, fmt.Sprintf("instances/%s", instanceID), string(body))
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

// Overwrite replaces instance data from another instance
func (i *instanceService) OverwriteFromInstance(ctx context.Context, instanceID string, sourceInstanceID string) (*OverwriteInstanceResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "overwriting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	if sourceInstanceID == "" {
		return nil, fmt.Errorf("must provide sourceInstanceID")
	}

	if sourceInstanceID != "" {
		if err := utils.ValidateInstanceID(sourceInstanceID); err != nil {
			return nil, fmt.Errorf("invalid source instance ID: %w", err)
		}
	}

	requestBody := overwriteInstanceRequest{
		SourceInstanceID: sourceInstanceID,
	}

	body, err := utils.Marshal(requestBody)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, fmt.Sprintf("instances/%s/overwrite", instanceID), string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to overwrite instance from another instance ", slog.String("instanceID", instanceID), slog.String("sourceInstanceID", sourceInstanceID), slog.String("error", err.Error()))
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

// ValidateCreateInstanceConfig performs basic checks that the min number
// configuration options have been supplied when creating an instance
func validateCreateInstanceConfig(instanceConfig *CreateInstanceConfigData) error {

	// Region name cannot be empty
	if instanceConfig.Region == "" {
		return fmt.Errorf("region must not be empty")
	}

	// Memroy cannot be empty
	if instanceConfig.Memory == "" {
		return fmt.Errorf("memory must not be empty")
	}

	// Type cannot be empty
	if instanceConfig.Type == "" {
		return fmt.Errorf("instance type must not be empty")
	}

	// Cloud provider cannot be empty
	if instanceConfig.CloudProvider == "" {
		return fmt.Errorf("cloud provider must not be empty")
	}

	// Instance name cannot be empty or greater than 30 characters
	if instanceConfig.Name == "" {
		return fmt.Errorf("instance name must not be empty")
	}

	if len(instanceConfig.Name) > 30 {
		return fmt.Errorf("instance name must be less than 30 characters long")

	}
	// TenantID cannot be empty
	if instanceConfig.TenantID == "" {
		return fmt.Errorf("tenant ID must not be empty")
	}

	// Check the format of the TenantID
	err := utils.ValidateTenantID(instanceConfig.TenantID)
	if err != nil {
		return fmt.Errorf("tenant ID must be a valid UUID format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)")
	}

	return nil
}

// Overwrite replaces instance data from a snapshot
func (i *instanceService) OverwriteFromSnapshot(ctx context.Context, instanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		i.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.DebugContext(ctx, "overwriting instance", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		i.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	if sourceSnapshotID == "" {
		return nil, fmt.Errorf("must provide sourceSnapshotID")
	}

	requestBody := overwriteInstanceRequest{
		SourceSnapshotID: sourceSnapshotID,
	}

	body, err := utils.Marshal(requestBody)
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to marshal instance request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := i.api.Post(ctx, fmt.Sprintf("instances/%s/overwrite", instanceID), string(body))
	if err != nil {
		i.logger.ErrorContext(ctx, "failed to overwrite instance with a snapshot", slog.String("instanceID", instanceID), slog.String("snapshotID", sourceSnapshotID), slog.String("error", err.Error()))
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
