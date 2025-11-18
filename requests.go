package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// APIError represents an error response from the Aura API
type APIError struct {
	StatusCode int              `json:"status_code"`
	Message    string           `json:"message"`
	Details    []APIErrorDetail `json:"details,omitempty"`
}

// APIErrorDetail represents individual error details
type APIErrorDetail struct {
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
	}

	// Include first detail in the error message
	detail := e.Details[0]
	msg := fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Message, detail.Message)
	if len(e.Details) > 1 {
		msg += fmt.Sprintf(" (and %d more error(s))", len(e.Details)-1)
	}
	return msg
}

// AllErrors returns all error messages as a slice
func (e *APIError) AllErrors() []string {
	errors := []string{e.Message}
	for _, detail := range e.Details {
		errors = append(errors, detail.Message)
	}
	return errors
}

// HasMultipleErrors returns true if there are multiple error details
func (e *APIError) HasMultipleErrors() bool {
	return len(e.Details) > 1
}

// IsNotFound returns true if the error is a 404
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is a 401
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsBadRequest returns true if the error is a 400
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

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

	// Add timeout for long-running operations of 120 seconds.  If you used 120 only, this would be 120 nanseconds.
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

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
		// Parse the error message we got back
		logger.DebugContext(ctx, "API returned error",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.Int("statusCode", response.RequestResponse.StatusCode),
			slog.String("message", response.RequestResponse.Status),
			slog.String("body", string(*response.ResponsePayload)),
		)

		apiErr := parseAPIError(*response.ResponsePayload, response.RequestResponse.StatusCode)

		return nil, apiErr
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
		logger.DebugContext(ctx, "failed to unmarshal response",
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

// parseAPIError attempts to parse the error response from the API
func parseAPIError(responsePayload []byte, statusCode int) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
	}

	if len(responsePayload) == 0 {
		return apiErr
	}

	// Try to parse error response body - adjust structure based on actual API response
	var errResponse struct {
		Message string           `json:"message"`
		Errors  []APIErrorDetail `json:"errors"`  // common format
		Details []APIErrorDetail `json:"details"` // alternative format
	}

	if err := json.Unmarshal(responsePayload, &errResponse); err == nil {
		if errResponse.Message != "" {
			apiErr.Message = errResponse.Message
		}

		// Use whichever field is populated
		if len(errResponse.Errors) > 0 {
			apiErr.Details = errResponse.Errors
		} else if len(errResponse.Details) > 0 {
			apiErr.Details = errResponse.Details
		}
	}

	return apiErr
}
