package aura

import (
	"context"
	"log/slog"
	http "net/http"
	"net/url"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Used for authentication with endpoints
// Stores the auth token for use with the Aura API
type APIAuth struct {
	Type   string `json:"token_type"`
	Token  string `json:"access_token"`
	Expiry int64  `json:"expires_in"`
}

// If needed, updates AuthManager token to make a request to the aura api otherwise it does nothing as the current token is still valid
func (am *authManager) getToken(ctx context.Context, httpClt httpClient.HTTPService) error {
	logger := slog.Default()
	var err error

	// See if we have a token.  If this was the first time this function was called, token will be empty.
	if len(am.Token) > 0 {
		// We do have a token, is it still valid?
		logger.DebugContext(ctx, "already have a token", slog.String("debug", ""))
		if time.Now().Unix() <= am.ExpiresAt-60 {
			// We are not within 60 seconds of expiring .  Our token is still valid
			logger.DebugContext(ctx, "token is still valid", slog.String("debug", ""))
			return nil
		}
	}

	// To get a token, we use Basic Auth for the Aura API token endpoint
	auth := "Basic" + " " + utils.Base64Encode(am.Id, am.Secret)

	endpoint := "oauth/token"

	body := url.Values{}

	body.Set("grant_type", "client_credentials")

	newAuraAPIToken, err := makeAuthenticatedRequest[APIAuth](ctx, httpClt, auth, endpoint, http.MethodPost, "application/x-www-form-urlencoded", body.Encode())
	if err != nil {
		// Didn't get a token
		logger.ErrorContext(ctx, "unable to obtain an auth token", slog.String("error", err.Error()))
		return err
	}

	// Update the token details
	am.ObtainedAt = time.Now().Unix()
	am.Token = newAuraAPIToken.Token
	am.Type = newAuraAPIToken.Type
	am.ExpiresAt = time.Now().Unix() + newAuraAPIToken.Expiry

	return nil

}
