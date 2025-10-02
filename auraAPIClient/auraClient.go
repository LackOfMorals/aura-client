// package auraAPIClient provides functionality to use the Neo4j Aura API to provision, managed and then destory Aura instances
package auraAPIClient

import (
	"context"
	"net/http"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

// These are the interfaces that represent the functions available in this package

type GetAuthTokenExecutor interface {
	GetAuthToken() (*AuthAPIToken, error)
}

type ListTenantsExecutor interface {
	ListTenants(context.Context, *AuthAPIToken) (*ListTenantsResponse, error)
}

type GetTenantExecutor interface {
	GetTenant(context.Context, *AuthAPIToken, string) (*GetTenantResponse, error)
}

type ListInstancesExecutor interface {
	ListInstances(context.Context, *AuthAPIToken) (*ListInstancesResponse, error)
}

type CreateInstanceExecutor interface {
	CreateInstance(context.Context, *AuthAPIToken, *CreateInstanceConfigData) (*CreateInstanceResponse, error)
}

type DeleteInstanceExecutor interface {
	DeleteInstance(context.Context, *AuthAPIToken, string) (*GetInstanceResponse, error)
}

type GetInstanceExecutor interface {
	GetInstance(context.Context, *AuthAPIToken, string) (*GetInstanceResponse, error)
}

// Aura API service
type AuraAPIService interface {
	GetAuthTokenExecutor
	ListTenantsExecutor
	GetTenantExecutor
	ListInstancesExecutor
	CreateInstanceExecutor
	DeleteInstanceExecutor
	GetInstanceExecutor
}

// This is the concrete implementation for Aura API Service
type AuraAPIActionsService struct {
	AuraAPIBaseURL string
	AuraAPIVersion string
	AuraAPITimeout string
	ClientID       string
	ClientSecret   string
}

// NewDriver is the entry point to the auraClient driver to create an instance of a Driver. It is the first function to
// be called in order to establish a connection to a neo4j database. It requires the Aura API base URL, the version of the
// Aura API to use, client id, and client secret.
func NewAuraAPIActionsService(baseurl, ver, timeout, id, sec string) AuraAPIService {
	return &AuraAPIActionsService{
		AuraAPIBaseURL: baseurl,
		AuraAPIVersion: ver,
		AuraAPITimeout: timeout,
		ClientID:       id,
		ClientSecret:   sec,
	}
}

// makeAuthenticatedRequest handles the common pattern of making an authenticated API request
// and unmarshalling the response into the desired type
func makeAuthenticatedRequest[T any](
	ctx context.Context,
	a *AuraAPIActionsService,
	token *AuthAPIToken,
	endpoint string,
	method string,
	body []byte,
) (*T, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := token.Type + " " + token.Token

	header := http.Header{
		"Content-Type":  {"application/json"},
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
