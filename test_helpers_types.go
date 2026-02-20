package aura

import (
	"context"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// mockAPIService is a basic mock of api.RequestService.
// It records the last call details but does not inspect the context.
type mockAPIService struct {
	response   *api.Response
	err        error
	lastMethod string
	lastPath   string
	lastBody   string
}

// mockAPIServiceWithDelay is a mock that can simulate slow responses and respects
// context cancellation / deadlines.
type mockAPIServiceWithDelay struct {
	response   *api.Response
	err        error
	delay      time.Duration
	lastMethod string
	lastPath   string
	lastBody   string
	callCount  int
}

// mockAPIServiceWithCallback is a mock that accepts optional callback hooks so tests
// can inspect the context or other call parameters at the point the API is invoked.
type mockAPIServiceWithCallback struct {
	response   *api.Response
	err        error
	delay      time.Duration
	lastMethod string
	lastPath   string
	lastBody   string
	callCount  int

	OnGet    func(ctx context.Context, endpoint string) error
	OnPost   func(ctx context.Context, endpoint string, body string) error
	OnPut    func(ctx context.Context, endpoint string, body string) error
	OnPatch  func(ctx context.Context, endpoint string, body string) error
	OnDelete func(ctx context.Context, endpoint string) error
}
