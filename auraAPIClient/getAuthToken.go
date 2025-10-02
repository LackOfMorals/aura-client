package auraAPIClient

import (
	"encoding/json"
	"net/http"
	"net/url"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

// Obtains a token to use with the Aura API using a Client ID and Client Secret
func (a *AuraAPIActionsService) GetAuthToken() (*AuthAPIToken, error) {

	endpoint := "oauth/token"

	myHTTPClient := httpClient.NewHTTPRequestService(a.AuraAPIBaseURL, "120")

	auth := "Basic " + utils.Base64Encode(a.ClientID, a.ClientSecret)

	header := http.Header{"Content-Type": {"application/x-www-form-urlencoded"},
		"User-Agent": {"jgHTTPClient"}, "Authorization": {auth},
	}

	body := url.Values{}

	body.Set("grant_type", "client_credentials")

	response, err := myHTTPClient.MakeRequest(endpoint, http.MethodPost, header, []byte(body.Encode()))

	if err != nil {
		return nil, err
	}

	var authToken AuthAPIToken

	// Unmarshall response into a JSON payload
	err = json.Unmarshal(*response.ResponsePayload, &authToken)
	if err != nil {
		return nil, err
	}

	return &authToken, err

}
