package auraAPIClient

import (
	"net/http"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

// Lists the instances in a tenant. Requires a token of type AuthAPIToken with response returned as ListInstancesResponse.
func (a *AuraAPIActionsService) ListInstances(token *AuthAPIToken) (*ListInstancesResponse, error) {

	endpoint := a.AuraAPIVersion + "/instances"

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
	jsonDoc, err := utils.Unmarshal[ListInstancesResponse](*response.ResponsePayload)

	return &jsonDoc, err

}

func (a *AuraAPIActionsService) CreateInstance(token *AuthAPIToken, instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error) {

	endpoint := a.AuraAPIVersion + "/instances"

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := token.Type + " " + token.Token

	header := http.Header{"Content-Type": {"application/json"},
		"User-Agent": {"jgHTTPClient"}, "Authorization": {auth},
	}

	// Marhsall CreateInstanceConfigData into the JSON required as
	//  an arry of bytes for MakeRequest
	body, err := utils.Marshall(instanceRequest)

	if err != nil {
		return nil, err
	}

	response, err := myHTTPClient.MakeRequest(endpoint, http.MethodPost, header, body)

	if err != nil {
		return nil, err
	}

	// Unmarshall payload into JSON
	jsonDoc, err := utils.Unmarshal[CreateInstanceResponse](*response.ResponsePayload)

	return &jsonDoc, err

}

func (a *AuraAPIActionsService) DeleteInstance(token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {

	endpoint := a.AuraAPIVersion + "/instances" + "/" + instanceID

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := token.Type + " " + token.Token

	header := http.Header{"Content-Type": {"application/json"},
		"User-Agent": {"jgHTTPClient"}, "Authorization": {auth},
	}

	response, err := myHTTPClient.MakeRequest(endpoint, http.MethodDelete, header, nil)

	if err != nil {
		return nil, err
	}

	// Unmarshall payload into JSON
	jsonDoc, err := utils.Unmarshal[GetInstanceResponse](*response.ResponsePayload)

	return &jsonDoc, err

}

func (a *AuraAPIActionsService) GetInstance(token *AuthAPIToken, instanceID string) (*GetInstanceResponse, error) {

	endpoint := a.AuraAPIVersion + "/instances" + "/" + instanceID

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
	jsonDoc, err := utils.Unmarshal[GetInstanceResponse](*response.ResponsePayload)

	return &jsonDoc, err

}
