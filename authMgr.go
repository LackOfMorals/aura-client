package aura

import (
	"context"
	"log/slog"
	http "net/http"
	"net/url"
	"sync"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// Token management
type authManager struct {
	id         string       // the client id to obtain a token to use with the aura api
	secret     string       // the client secret to obtain a token to use with the aura api
	tokenType  string       // the type of token from the aura api auth endpoint
	token      string       // the token from aura api auth endpoint
	obtainedAt int64        // The time when the token was obtained in number of seconds since midnight Jan 1st 1970
	expiresAt  int64        // token duration in seconds
	logger     *slog.Logger // the logger...
	mu         sync.RWMutex // mutex to ensure thread safety in the unlikely event of concurrent access.

}

// Used for authentication with endpoints
// Stores the auth token back from the auth endpoint
type apiAuth struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
	Expiry    int64  `json:"expires_in"`
}

// If needed, updates AuthManager token to make a request to the aura api otherwise it does nothing as the current token is still valid
func (am *authManager) getToken(ctx context.Context, httpClt httpClient.HTTPService) error {
	var err error

	am.mu.RLock()
	// See if we have a token.  If this was the first time this function was called, token will be empty.
	if len(am.token) > 0 && time.Now().Unix() <= am.expiresAt-60 {
		am.logger.DebugContext(ctx, "token is still valid", slog.String("debug", ""))
		am.mu.RUnlock()
		return nil
	}
	am.mu.RUnlock()

	am.mu.Lock()
	defer am.mu.Unlock()

	// Double-check after acquiring write lock
	if len(am.token) > 0 && time.Now().Unix() <= am.expiresAt-60 {
		return nil
	}

	//  Aura Auth endpoint requires Basic Auth
	auth := "Basic" + " " + utils.Base64Encode(am.id, am.secret)

	endpoint := "oauth/token"

	body := url.Values{}

	body.Set("grant_type", "client_credentials")

	newToken, err := makeAuthenticatedRequest[apiAuth](ctx, httpClt, auth, endpoint, http.MethodPost, "application/x-www-form-urlencoded", body.Encode(), am.logger)
	if err != nil {
		// Didn't get a token
		am.logger.DebugContext(ctx, "unable to obtain an auth token", slog.String("error", err.Error()))
		return err
	}

	// Update the token details
	am.obtainedAt = time.Now().Unix()
	am.token = newToken.Token
	am.tokenType = newToken.TokenType
	am.expiresAt = time.Now().Unix() + newToken.Expiry

	return nil

}
