package resources

import (
	"context"

	"github.com/LackOfMorals/aura-client"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// makeAuthenticatedRequest handles the common pattern of making an authenticated API request
// and unmarshalling the response into the desired type
func makeAuthenticatedRequest[T any](
	ctx context.Context,
	a *aura.AuraAPIActionsService,
	token *aura.AuthAPIToken,
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

	userAgent := "aura-go-client"

	header := map[string]string{
		"Content-Type":  contentType,
		"User-Agent":    userAgent,
		"Authorization": auth,
	}

	/*
		header := http.Header{
			"Content-Type":  {contentType},
			"User-Agent":    {userAgent},
			"Authorization": {auth},
		}
	*/

	response, err := a.Http.MakeRequest(ctx, endpoint, method, header, body)
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
