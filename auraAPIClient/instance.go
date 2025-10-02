package auraAPIClient

import (
	"context"
	"net/http"

	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

const (
	userAgent = "jgHTTPClient"
)

// Instance methods
func (i *InstanceService) List(ctx context.Context, token *AuthAPIToken) (*ListInstancesResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances"
	return makeAuthenticatedRequest[ListInstancesResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", nil)
}

func (i *InstanceService) Get(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", nil)
}

func (i *InstanceService) Create(ctx context.Context, token *AuthAPIToken, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[CreateInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", body)
}

func (i *InstanceService) Delete(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodDelete, "application/json", nil)
}

func (i *InstanceService) Pause(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID + "/pause"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", nil)
}

func (i *InstanceService) Resume(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID + "/resume"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", nil)
}
