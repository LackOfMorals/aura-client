package resources

import (
	"context"
	"net/http"
	"net/url"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// AuthService handles authentication operations
type AuthService struct {
	Service *AuraAPIActionsService
}

// Obtains a token to use with the Aura API using a Client ID and Client Secret
func (a *AuthService) GetAuthToken(ctx context.Context) (*AuthAPIToken, error) {

	// We'll use this type to store our 'token'
	// So we can make use of makeAuthenticatedRequest

	var authToken AuthAPIToken

	endpoint := "oauth/token"

	authToken.Token = utils.Base64Encode(a.Service.ClientID, a.Service.ClientSecret)
	authToken.Type = "Basic"

	body := url.Values{}

	body.Set("grant_type", "client_credentials")

	return makeAuthenticatedRequest[AuthAPIToken](ctx, a.Service, &authToken, endpoint, http.MethodPost, "application/x-www-form-urlencoded", body.Encode())

}
