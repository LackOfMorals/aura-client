package resources

import (
	"context"
	"net/http"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// InstanceService handles instance operations
type InstanceService struct {
	Service *AuraAPIActionsService
}

// Instance methods
func (i *InstanceService) List(ctx context.Context, token *AuthAPIToken) (*ListInstancesResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances"
	return makeAuthenticatedRequest[ListInstancesResponse](ctx, i.Service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Get(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.Service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Create(ctx context.Context, token *AuthAPIToken, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[CreateInstanceResponse](ctx, i.Service, token, endpoint, http.MethodPost, "application/json", string(body))
}

func (i *InstanceService) Delete(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.Service, token, endpoint, http.MethodDelete, "application/json", "")
}

func (i *InstanceService) Pause(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances/" + instanceID + "/pause"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.Service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Resume(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances/" + instanceID + "/resume"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.Service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Update(ctx context.Context, token *AuthAPIToken, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	endpoint := i.Service.AuraAPIVersion + "/instances/" + instanceID

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.Service, token, endpoint, http.MethodPatch, "application/json", string(body))
}
