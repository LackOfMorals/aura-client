package aura

import (
	"context"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// makeAuthenticatedRequest handles the common pattern of making an authenticated API request
// and unmarshalling the response into the desired type
func makeAuthenticatedRequest[T any](
	ctx context.Context,
	h httpClient.HTTPService,
	auth string,
	endpoint string,
	method string,
	contentType string,
	body string,
	logger *slog.Logger,
) (*T, error) {

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		logger.ErrorContext(ctx, "context already cancelled before request", slog.String("error", err.Error()))
		return nil, err
	}

	userAgent := "aura-go-client"

	header := map[string]string{
		"Content-Type":  contentType,
		"User-Agent":    userAgent,
		"Authorization": auth,
	}

	logger.DebugContext(ctx, "making HTTP request",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
		slog.String("contentType", contentType),
	)

	response, err := h.MakeRequest(ctx, endpoint, method, header, body)
	if err != nil {
		logger.ErrorContext(ctx, "HTTP request failed",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.DebugContext(ctx, "HTTP request successful",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
		slog.Int("statusCode", response.RequestResponse.StatusCode),
	)

	// Unmarshall JSON payload into the receiving struct type
	// that will be returned
	returnedStruct, err := utils.Unmarshal[T](*response.ResponsePayload)
	if err != nil {
		logger.ErrorContext(ctx, "failed to unmarshal response",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.DebugContext(ctx, "response unmarshalled successfully",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
	)

	return &returnedStruct, nil
}
