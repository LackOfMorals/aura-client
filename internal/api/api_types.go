package api

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
)

// Response represents a response from the Aura API
type Response struct {
	StatusCode int
	Body       []byte
}

// Error represents an error response from the Aura API
type Error struct {
	StatusCode int           `json:"status_code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail represents individual error details
type ErrorDetail struct {
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
	Field   string `json:"field,omitempty"`
}

// Config holds configuration for the API service
type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	APIVersion   string
	Timeout      time.Duration // ← moved here from client.go
	MaxRetry     int           // ← moved here from client.go
}

// apiRequestService is the concrete implementation of RequestService
type apiRequestService struct {
	httpClient   httpClient.HTTPService
	authMgr      *authManager
	baseURL      string
	endpointBase string
	logger       *slog.Logger
}

// authManager handles token management for the API
type authManager struct {
	clientID     string
	clientSecret string
	tokenType    string
	token        string

	expiresAt int64
	logger    *slog.Logger
	mu        sync.RWMutex
}

// tokenResponse represents the OAuth token response
type tokenResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
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
