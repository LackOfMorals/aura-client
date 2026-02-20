package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Error implements the error interface
func (e *Error) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
	}

	detail := e.Details[0]
	msg := fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Message, detail.Message)
	if len(e.Details) > 1 {
		msg += fmt.Sprintf(" (and %d more error(s))", len(e.Details)-1)
	}
	return msg
}

// AllErrors returns all error messages as a slice
func (e *Error) AllErrors() []string {
	errors := []string{e.Message}
	for _, detail := range e.Details {
		errors = append(errors, detail.Message)
	}
	return errors
}

// HasMultipleErrors returns true if there are multiple error details
func (e *Error) HasMultipleErrors() bool {
	return len(e.Details) > 1
}

// IsNotFound returns true if the error is a 404
func (e *Error) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is a 401
func (e *Error) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsBadRequest returns true if the error is a 400
func (e *Error) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// RequestService defines the interface for making authenticated API requests.
// This is the middle layer that handles authentication and common API patterns.
type RequestService interface {
	Get(ctx context.Context, endpoint string) (*Response, error)
	Post(ctx context.Context, endpoint string, body string) (*Response, error)
	Put(ctx context.Context, endpoint string, body string) (*Response, error)
	Patch(ctx context.Context, endpoint string, body string) (*Response, error)
	Delete(ctx context.Context, endpoint string) (*Response, error)
}

// Config holds configuration for the API service
type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	APIVersion   string
}

// apiRequestService is the concrete implementation of RequestService
type apiRequestService struct {
	httpClient httpClient.HTTPService
	authMgr    *authManager
	baseURL    string
	apiVersion string
	logger     *slog.Logger
}

// authManager handles token management for the API
type authManager struct {
	clientID     string
	clientSecret string
	tokenType    string
	token        string
	obtainedAt   int64
	expiresAt    int64
	logger       *slog.Logger
	mu           sync.RWMutex
}

// tokenResponse represents the OAuth token response
type tokenResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// NewRequestService creates a new RequestService
func NewRequestService(httpSvc httpClient.HTTPService, cfg Config, logger *slog.Logger) RequestService {
	return &apiRequestService{
		httpClient: httpSvc,
		authMgr: &authManager{
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			logger:       logger,
		},
		baseURL:    cfg.BaseURL,
		apiVersion: cfg.APIVersion,
		logger:     logger,
	}
}

// Get performs an authenticated GET request
func (s *apiRequestService) Get(ctx context.Context, endpoint string) (*Response, error) {
	return s.doAuthenticatedRequest(ctx, http.MethodGet, endpoint, "")
}

// Post performs an authenticated POST request
func (s *apiRequestService) Post(ctx context.Context, endpoint string, body string) (*Response, error) {
	return s.doAuthenticatedRequest(ctx, http.MethodPost, endpoint, body)
}

// Put performs an authenticated PUT request
func (s *apiRequestService) Put(ctx context.Context, endpoint string, body string) (*Response, error) {
	return s.doAuthenticatedRequest(ctx, http.MethodPut, endpoint, body)
}

// Patch performs an authenticated PATCH request
func (s *apiRequestService) Patch(ctx context.Context, endpoint string, body string) (*Response, error) {
	return s.doAuthenticatedRequest(ctx, http.MethodPatch, endpoint, body)
}

// Delete performs an authenticated DELETE request
func (s *apiRequestService) Delete(ctx context.Context, endpoint string) (*Response, error) {
	return s.doAuthenticatedRequest(ctx, http.MethodDelete, endpoint, "")
}

// doAuthenticatedRequest handles the common pattern of making an authenticated API request.
// It trusts the deadline already set on ctx by the calling service layer — no additional
// timeout is applied here. Full URLs (http:// or https://) are used as-is; relative
// paths are resolved by the httpClient layer against the configured base URL.
func (s *apiRequestService) doAuthenticatedRequest(ctx context.Context, method, endpoint, body string) (*Response, error) {
	if err := ctx.Err(); err != nil {
		s.logger.ErrorContext(ctx, "context already cancelled before request", slog.String("error", err.Error()))
		return nil, err
	}

	tokenType, token, err := s.authMgr.ensureValidToken(ctx, s.baseURL, s.httpClient)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"User-Agent":    "aura-go-client",
		"Authorization": tokenType + " " + token,
	}

	s.logger.DebugContext(ctx, "making authenticated API request",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
	)

	var resp *httpClient.HTTPResponse

	switch method {
	case http.MethodGet:
		resp, err = s.httpClient.Get(ctx, endpoint, headers)
	case http.MethodPost:
		resp, err = s.httpClient.Post(ctx, endpoint, headers, body)
	case http.MethodPut:
		resp, err = s.httpClient.Put(ctx, endpoint, headers, body)
	case http.MethodPatch:
		resp, err = s.httpClient.Patch(ctx, endpoint, headers, body)
	case http.MethodDelete:
		resp, err = s.httpClient.Delete(ctx, endpoint, headers)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "HTTP request failed",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := parseError(resp.Body, resp.StatusCode)
		s.logger.DebugContext(ctx, "API returned error",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.Int("statusCode", resp.StatusCode),
			slog.String("message", apiErr.Message),
		)
		return nil, apiErr
	}

	s.logger.DebugContext(ctx, "API request successful",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
		slog.Int("statusCode", resp.StatusCode),
	)

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
	}, nil
}

// ensureValidToken gets or refreshes the authentication token and returns it to the caller.
// Token fields are always read while the mutex is held to prevent data races.
func (am *authManager) ensureValidToken(ctx context.Context, baseURL string, httpSvc httpClient.HTTPService) (tokenType, token string, err error) {
	am.mu.RLock()
	if len(am.token) > 0 && time.Now().Unix() <= am.expiresAt-60 {
		t, tt := am.token, am.tokenType
		am.mu.RUnlock()
		return tt, t, nil
	}
	am.mu.RUnlock()

	am.mu.Lock()
	defer am.mu.Unlock()

	// Double-check after acquiring the write lock — another goroutine may have
	// refreshed the token while we were waiting.
	if len(am.token) > 0 && time.Now().Unix() <= am.expiresAt-60 {
		return am.tokenType, am.token, nil
	}

	am.logger.DebugContext(ctx, "obtaining new authentication token")

	auth := "Basic " + utils.Base64Encode(am.clientID, am.clientSecret)

	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": auth,
	}

	body := url.Values{}
	body.Set("grant_type", "client_credentials")

	resp, err := httpSvc.Post(ctx, baseURL+"/oauth/token", headers, body.Encode())
	if err != nil {
		am.logger.DebugContext(ctx, "failed to obtain token", slog.String("error", err.Error()))
		return "", "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := parseError(resp.Body, resp.StatusCode)
		am.logger.DebugContext(ctx, "token request failed",
			slog.Int("statusCode", resp.StatusCode),
			slog.String("error", apiErr.Message),
		)
		return "", "", apiErr
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(resp.Body, &tokenResp); err != nil {
		am.logger.DebugContext(ctx, "failed to parse token response", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("failed to parse token response: %w", err)
	}

	am.obtainedAt = time.Now().Unix()
	am.token = tokenResp.AccessToken
	am.tokenType = tokenResp.TokenType
	am.expiresAt = time.Now().Unix() + tokenResp.ExpiresIn

	am.logger.DebugContext(ctx, "token obtained successfully",
		slog.Int64("expiresIn", tokenResp.ExpiresIn),
	)

	return am.tokenType, am.token, nil
}

// parseError attempts to parse an error response body from the API
func parseError(responseBody []byte, statusCode int) *Error {
	apiErr := &Error{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
	}

	if len(responseBody) == 0 {
		return apiErr
	}

	var errResponse struct {
		Message string        `json:"message"`
		Errors  []ErrorDetail `json:"errors"`
		Details []ErrorDetail `json:"details"`
	}

	if err := json.Unmarshal(responseBody, &errResponse); err == nil {
		if errResponse.Message != "" {
			apiErr.Message = errResponse.Message
		}
		if len(errResponse.Errors) > 0 {
			apiErr.Details = errResponse.Errors
		} else if len(errResponse.Details) > 0 {
			apiErr.Details = errResponse.Details
		}
	}

	return apiErr
}
