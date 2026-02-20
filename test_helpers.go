package aura

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// testLogger creates a logger for testing that discards output.
func testLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

// ============================================================================
// mockAPIService — simple mock, does not check context
// ============================================================================

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

// ============================================================================
// mockAPIServiceWithDelay — respects context cancellation, can simulate slow APIs
// ============================================================================

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

// executeWithDelay simulates a slow API call that respects context cancellation.
func (m *mockAPIServiceWithDelay) executeWithDelay(ctx context.Context) (*api.Response, error) {
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

// ============================================================================
// mockAPIServiceWithCallback — supports hooks to inspect context values and
// verify propagation through service layers
// ============================================================================

func (m *mockAPIServiceWithCallback) Get(ctx context.Context, endpoint string) (*api.Response, error) {
	m.lastMethod = "GET"
	m.lastPath = endpoint
	m.callCount++
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
