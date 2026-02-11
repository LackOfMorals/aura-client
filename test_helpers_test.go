package aura

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// testLogger creates a logger for testing that outputs to stderr
func testLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

// mockAPIService is a mock implementation of api.RequestService for testing
type mockAPIService struct {
	response   *api.Response
	err        error
	lastMethod string
	lastPath   string
	lastBody   string
}

func (m *mockAPIService) Get(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "GET"
	m.lastPath = endpoint
	return m.response, m.err
}

func (m *mockAPIService) Post(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "POST"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Put(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PUT"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Patch(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PATCH"
	m.lastPath = endpoint
	m.lastBody = body
	return m.response, m.err
}

func (m *mockAPIService) Delete(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "DELETE"
	m.lastPath = endpoint
	return m.response, m.err
}

// mockAPIServiceWithDelay is an enhanced mock that can simulate API delays
type mockAPIServiceWithDelay struct {
	response   *api.Response
	err        error
	delay      time.Duration // Simulates slow API response
	lastMethod string
	lastPath   string
	lastBody   string
	callCount  int
}

func (m *mockAPIServiceWithDelay) Get(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "GET"
	m.lastPath = endpoint
	m.callCount++
	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithDelay) Post(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "POST"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++
	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithDelay) Put(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PUT"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++
	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithDelay) Patch(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PATCH"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++
	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithDelay) Delete(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "DELETE"
	m.lastPath = endpoint
	m.callCount++
	return m.executeWithDelay(ctx)
}

// executeWithDelay simulates an API call with optional delay, respecting context cancellation
func (m *mockAPIServiceWithDelay) executeWithDelay(ctx context.Context) (*api.Response, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
			// Delay completed normally
		case <-ctx.Done():
			// Context cancelled during delay
			return nil, ctx.Err()
		}
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return m.response, m.err
}

// mockAPIServiceWithCallback is a mock that supports callback hooks for context verification
// This is useful for tests that need to verify context values or track context propagation

type mockAPIServiceWithCallback struct {
	response   *api.Response
	err        error
	delay      time.Duration
	lastMethod string
	lastPath   string
	lastBody   string
	callCount  int

	// Callback hooks for context verification
	// Return error to short-circuit the call, return nil to continue normally
	OnGet    func(ctx context.Context, endpoint string) error
	OnPost   func(ctx context.Context, endpoint string, body string) error
	OnPut    func(ctx context.Context, endpoint string, body string) error
	OnPatch  func(ctx context.Context, endpoint string, body string) error
	OnDelete func(ctx context.Context, endpoint string) error
}

func (m *mockAPIServiceWithCallback) Get(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "GET"
	m.lastPath = endpoint
	m.callCount++

	// Call hook if provided
	if m.OnGet != nil {
		if err := m.OnGet(ctx, endpoint); err != nil {
			return nil, err
		}
	}

	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithCallback) Post(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "POST"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++

	if m.OnPost != nil {
		if err := m.OnPost(ctx, endpoint, body); err != nil {
			return nil, err
		}
	}

	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithCallback) Put(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PUT"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++

	if m.OnPut != nil {
		if err := m.OnPut(ctx, endpoint, body); err != nil {
			return nil, err
		}
	}

	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithCallback) Patch(ctx context.Context, endpoint string, body string) (*api.Response, error) {
	m.lastMethod = "PATCH"
	m.lastPath = endpoint
	m.lastBody = body
	m.callCount++

	if m.OnPatch != nil {
		if err := m.OnPatch(ctx, endpoint, body); err != nil {
			return nil, err
		}
	}

	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithCallback) Delete(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "DELETE"
	m.lastPath = endpoint
	m.callCount++

	if m.OnDelete != nil {
		if err := m.OnDelete(ctx, endpoint); err != nil {
			return nil, err
		}
	}

	return m.executeWithDelay(ctx)
}

func (m *mockAPIServiceWithCallback) executeWithDelay(ctx context.Context) (*api.Response, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return m.response, m.err
}

// contextTestHelper provides common test utilities for context testing
type contextTestHelper struct {
	defaultTimeout time.Duration
}

func newContextTestHelper() *contextTestHelper {
	return &contextTestHelper{
		defaultTimeout: 30 * time.Second,
	}
}

// createCancelledContext creates a context that's already cancelled
func (h *contextTestHelper) createCancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// createTimeoutContext creates a context with very short timeout
func (h *contextTestHelper) createTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// createNormalContext creates a normal background context for testing
func (h *contextTestHelper) createNormalContext() context.Context {
	return context.Background()
}
