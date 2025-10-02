// package auraAPIClient provides functionality to use the Neo4j Aura API to provision, managed and then destory Aura instances
package auraAPIClient

import (
	"context"
	"net/http"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

// Core service configuration
type AuraAPIActionsService struct {
	auraAPIBaseURL string
	auraAPIVersion string
	auraAPITimeout string
	clientID       string
	clientSecret   string
	timeout        string

	// Grouped services
	Auth      *AuthService
	Tenants   *TenantService
	Instances *InstanceService
}

// AuthService handles authentication operations
type AuthService struct {
	service *AuraAPIActionsService
}

// TenantService handles tenant operations
type TenantService struct {
	service *AuraAPIActionsService
}

// InstanceService handles instance operations
type InstanceService struct {
	service *AuraAPIActionsService
}

// NewAuraAPIActionsService creates a new Aura API service with grouped sub-services
func NewAuraAPIActionsService(baseurl, ver, timeout, id, sec string) *AuraAPIActionsService {
	service := &AuraAPIActionsService{
		auraAPIBaseURL: baseurl,
		auraAPIVersion: ver,
		auraAPITimeout: timeout,
		clientID:       id,
		clientSecret:   sec,
		timeout:        timeout,
	}

	// Initialize sub-services with reference to parent
	service.Auth = &AuthService{service: service}
	service.Tenants = &TenantService{service: service}
	service.Instances = &InstanceService{service: service}

	return service
}

// makeAuthenticatedRequest handles the common pattern of making an authenticated API request
// and unmarshalling the response into the desired type
func makeAuthenticatedRequest[T any](
	ctx context.Context,
	a *AuraAPIActionsService,
	token *AuthAPIToken,
	endpoint string,
	method string,
	contentType string,
	body []byte,
) (*T, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	myHTTPClient := httpClient.NewHTTPRequestService(a.auraAPIBaseURL, a.timeout)

	auth := token.Type + " " + token.Token

	header := http.Header{
		"Content-Type":  {contentType},
		"User-Agent":    {userAgent},
		"Authorization": {auth},
	}

	response, err := myHTTPClient.MakeRequest(endpoint, method, header, body)
	if err != nil {
		return nil, err
	}

	// Unmarshall payload into JSON
	jsonDoc, err := utils.Unmarshal[T](*response.ResponsePayload)
	if err != nil {
		return nil, err
	}

	return &jsonDoc, nil
}
