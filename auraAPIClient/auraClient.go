// package auraAPIClient provides functionality to use the Neo4j Aura API to provision, managed and then destory Aura instances
package auraAPIClient

import (
	"context"
	"fmt"
	"time"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
	utils "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/utils"
)

const (
	BaseURL    = "https://api.neo4j.io/"
	ApiVersion = "v1"
	ApiTimeout = 120 * time.Second
)

// Core service configuration
type AuraAPIActionsService struct {
	auraAPIBaseURL string
	auraAPIVersion string
	auraAPITimeout time.Duration
	clientID       string
	clientSecret   string
	timeout        time.Duration

	http httpClient.HTTPService

	// Grouped services
	Auth      *AuthService
	Tenants   *TenantService
	Instances *InstanceService
	Snapshots *SnapshotService
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

// SnapshotService handles snapshot operations
type SnapshotService struct {
	service *AuraAPIActionsService
}

// NewAuraAPIActionsService creates a new Aura API service with grouped sub-services
func NewAuraAPIActionsService(id, sec string) *AuraAPIActionsService {

	service := &AuraAPIActionsService{
		auraAPIBaseURL: BaseURL,
		auraAPIVersion: ApiVersion,
		auraAPITimeout: ApiTimeout,
		clientID:       id,
		clientSecret:   sec,
		timeout:        ApiTimeout,
	}

	// Reuse a single HTTP client/service instance with configured base URL and timeout
	service.http = httpClient.NewHTTPRequestService(service.auraAPIBaseURL, service.timeout)

	// Initialize sub-services with reference to parent
	service.Auth = &AuthService{service: service}
	service.Tenants = &TenantService{service: service}
	service.Instances = &InstanceService{service: service}
	service.Snapshots = &SnapshotService{service: service}

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
	body string,
) (*T, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	auth := token.Type + " " + token.Token

	var header map[string]string

	// Initializing the Map
	header = make(map[string]string)

	header["Content-Type"] = contentType
	header["User-Agent"] = userAgent
	header["Authorization"] = auth

	/*
		header := http.Header{
			"Content-Type":  {contentType},
			"User-Agent":    {userAgent},
			"Authorization": {auth},
		}
	*/

	response, err := a.http.MakeRequest(ctx, endpoint, method, header, body)
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

func checkDate(t string) error {

	_, err := time.Parse(time.DateOnly, t)
	if err != nil {
		return fmt.Errorf("Date must in the format of YYYY-MM-DD")
	}

	return nil

}
