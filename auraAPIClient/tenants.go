package auraAPIClient

import (
	"net/http"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

// Retrieves information for a Tenant that includes permitted instance configurations
func (a *AuraAPIActionsService) GetTenant(token *AuthAPIToken, TenantID string) (*GetTenantResponse, error) {

	endpoint := a.AuraAPIVersion + "/tenants/" + TenantID

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := token.Type + " " + token.Token

	header := http.Header{"Content-Type": {"application/json"},
		"User-Agent": {"jgHTTPClient"}, "Authorization": {auth},
	}

	response, err := myHTTPClient.MakeRequest(endpoint, http.MethodGet, header, nil)

	if err != nil {
		return nil, err

	}

	// Unmarshall payload into JSON
	jsonDoc, err := utils.Unmarshal[GetTenantResponse](*response.ResponsePayload)

	return &jsonDoc, err

}

// Obtains a token to use with the Aura API using a Client ID and Client Secret
func (a *AuraAPIActionsService) ListTenants(token *AuthAPIToken) (*ListTenantsResponse, error) {

	endpoint := a.AuraAPIVersion + "/tenants"

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := token.Type + " " + token.Token

	header := http.Header{"Content-Type": {"application/json"},
		"User-Agent": {"jgHTTPClient"}, "Authorization": {auth},
	}

	response, err := myHTTPClient.MakeRequest(endpoint, http.MethodGet, header, nil)

	if err != nil {
		return nil, err
	}

	// Unmarshall payload into JSON
	jsonDoc, err := utils.Unmarshal[ListTenantsResponse](*response.ResponsePayload)

	return &jsonDoc, err

}
