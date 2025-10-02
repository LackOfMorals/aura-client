package auraAPIClient

import (
	"context"
	"net/http"

	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

const (
	userAgent = "jgHTTPClient"
)

// ListInstances lists the instances in a tenant.
func (a *AuraAPIActionsService) ListInstances(ctx context.Context, token *AuthAPIToken) (*ListInstancesResponse, error) {
	endpoint := a.AuraAPIVersion + "/instances"
	return makeAuthenticatedRequest[ListInstancesResponse](ctx, a, token, endpoint, http.MethodGet, nil)
}

// CreateInstance creates a new instance with the provided configuration.
func (a *AuraAPIActionsService) CreateInstance(ctx context.Context, token *AuthAPIToken, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {
	endpoint := a.AuraAPIVersion + "/instances"

	body, err := utils.Marshall(instanceRequest)
	if err != nil {
		return nil, err
	}

	return makeAuthenticatedRequest[CreateInstanceResponse](ctx, a, token, endpoint, http.MethodPost, body)
}

// DeleteInstance deletes an instance by ID.
func (a *AuraAPIActionsService) DeleteInstance(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := a.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, a, token, endpoint, http.MethodDelete, nil)
}

// GetInstance retrieves an instance by ID.
func (a *AuraAPIActionsService) GetInstance(ctx context.Context, token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {
	endpoint := a.AuraAPIVersion + "/instances/" + instanceID
	return makeAuthenticatedRequest[GetInstanceResponse](ctx, a, token, endpoint, http.MethodGet, nil)
}
