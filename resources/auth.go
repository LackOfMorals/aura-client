package resources

import (
	"context"
	"net/http"
	"net/url"

	aura "github.com/LackOfMorals/aura-client"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// AuthService handles authentication operations
type AuthService struct {
	Service *aura.AuraAPIActionsService
}

// Obtains a token to use with the Aura API using a Client ID and Client Secret
func (a *AuthService) GetAuthToken(ctx context.Context) (*aura.AuthAPIToken, error) {

	// We'll use this type to store our 'token'
	// So we can make use of makeAuthenticatedRequest

	var authToken aura.AuthAPIToken

	endpoint := "oauth/token"

	authToken.Token = utils.Base64Encode(a.service.ClientID, a.service.ClientSecret)
	authToken.Type = "Basic"

	body := url.Values{}

	body.Set("grant_type", "client_credentials")

	return makeAuthenticatedRequest[aura.AuthAPIToken](ctx, a.service, &authToken, endpoint, http.MethodPost, "application/x-www-form-urlencoded", body.Encode())

}
