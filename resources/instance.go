package resources

import (
	"context"
	"net/http"

	"github.com/LackOfMorals/aura-client"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// InstanceService handles instance operations
type InstanceService struct {
	Service *aura.AuraAPIActionsService
}

// Instance methods
func (i *InstanceService) List(ctx context.Context, token *aura.AuthAPIToken) (*aura.ListInstancesResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances"
	return makeAuthenticatedRequest[aura.ListInstancesResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Get(ctx context.Context, token *aura.AuthAPIToken, instanceID string) (*aura.GetInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[aura.GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Create(ctx context.Context, token *aura.AuthAPIToken, instanceRequest *aura.CreateInstanceConfigData) (*aura.CreateInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[aura.CreateInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", string(body))
}

func (i *InstanceService) Delete(ctx context.Context, token *aura.AuthAPIToken, instanceID string) (*aura.GetInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[aura.GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodDelete, "application/json", "")
}

func (i *InstanceService) Pause(ctx context.Context, token *aura.AuthAPIToken, instanceID string) (*aura.GetInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances/" + instanceID + "/pause"
	return makeAuthenticatedRequest[aura.GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Resume(ctx context.Context, token *aura.AuthAPIToken, instanceID string) (*aura.GetInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances/" + instanceID + "/resume"
	return makeAuthenticatedRequest[aura.GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Update(ctx context.Context, token *aura.AuthAPIToken, instanceID string, instanceRequest *aura.UpdateInstanceData) (*aura.GetInstanceResponse, error) {
	endpoint := i.service.AuraAPIVersion + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[aura.GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPatch, "application/json", string(body))
}
