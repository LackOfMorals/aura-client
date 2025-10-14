package auraAPIClient

import (
	"context"
	"fmt"
	"net/http"
)

const (
	userAgent = "jgHTTPClient"
)

// Instance methods
func (i *InstanceService) List(ctx context.Context, token *AuthAPIToken) (*ListInstancesResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances"
	return makeAuthenticatedRequest[ListInstancesResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Get(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodGet, "application/json", "")
}

func (i *InstanceService) Create(ctx context.Context, token *AuthAPIToken, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances"

	body := fmt.Sprintf("%s", instanceRequest)

	/*
		body, err := utils.Marshal(instanceRequest)
		if err != nil {
			return nil, err
		}
	*/

	return makeAuthenticatedRequest[CreateInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", body)
}

func (i *InstanceService) Delete(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodDelete, "application/json", "")
}

func (i *InstanceService) Pause(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID + "/pause"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Resume(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID + "/resume"
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPost, "application/json", "")
}

func (i *InstanceService) Update(ctx context.Context, token *AuthAPIToken, instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error) {
	endpoint := i.service.auraAPIVersion + "/instances/" + instanceID

	body := fmt.Sprintf("%s", instanceRequest)

	/*
		body, err := utils.Marshal(instanceRequest)
		if err != nil {
			return nil, err
		}
	*/
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, i.service, token, endpoint, http.MethodPatch, "application/json", body)
}
