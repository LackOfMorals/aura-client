package aura_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	aura "github.com/LackOfMorals/aura-client"
)

// fakeAuraServer returns an httptest.Server that serves the OAuth token endpoint
// and a JSON error response for everything under /v1/. Tests can override
// onRequest to inspect the headers / body of the API call.
func fakeAuraServer(t *testing.T, status int, body string, onRequest func(r *http.Request)) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	return httptest.NewServer(mux)
}

// =============================================================================
// WithUserAgent
// =============================================================================

func TestWithUserAgent_ReplacesDefault(t *testing.T) {
	var got string
	srv := fakeAuraServer(t, 200, `{"data":[]}`, func(r *http.Request) {
		got = r.Header.Get("User-Agent")
	})
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
		aura.WithUserAgent("my-app/1.0 aura-go-client/test"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got != "my-app/1.0 aura-go-client/test" {
		t.Errorf("User-Agent = %q, want %q", got, "my-app/1.0 aura-go-client/test")
	}
}

func TestWithUserAgent_Empty_Errors(t *testing.T) {
	_, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithUserAgent(""),
	)
	if err == nil {
		t.Fatal("expected error for empty user agent")
	}
}

func TestWithUserAgent_DefaultWhenNotSet(t *testing.T) {
	var got string
	srv := fakeAuraServer(t, 200, `{"data":[]}`, func(r *http.Request) {
		got = r.Header.Get("User-Agent")
	})
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}

	if !strings.HasPrefix(got, "aura-go-client/") {
		t.Errorf("User-Agent = %q, want prefix aura-go-client/", got)
	}
}

// =============================================================================
// WithDefaultHeaders
// =============================================================================

func TestWithDefaultHeaders_AppliedToRequests(t *testing.T) {
	var preview, custom string
	srv := fakeAuraServer(t, 200, `{"data":[]}`, func(r *http.Request) {
		preview = r.Header.Get("Aura-Preview")
		custom = r.Header.Get("X-Trace-Id")
	})
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
		aura.WithDefaultHeaders(map[string]string{
			"Aura-Preview": "true",
			"X-Trace-Id":   "abc123",
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if preview != "true" {
		t.Errorf("Aura-Preview = %q, want true", preview)
	}
	if custom != "abc123" {
		t.Errorf("X-Trace-Id = %q, want abc123", custom)
	}
}

func TestWithDefaultHeaders_CannotOverrideReservedHeaders(t *testing.T) {
	var auth, ua, ct string
	srv := fakeAuraServer(t, 200, `{"data":[]}`, func(r *http.Request) {
		auth = r.Header.Get("Authorization")
		ua = r.Header.Get("User-Agent")
		ct = r.Header.Get("Content-Type")
	})
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
		aura.WithDefaultHeaders(map[string]string{
			"Authorization": "Bearer evil",
			"User-Agent":    "evil/1.0",
			"Content-Type":  "text/plain",
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if auth == "Bearer evil" {
		t.Error("Authorization was overridden by WithDefaultHeaders — must not be possible")
	}
	if ua == "evil/1.0" {
		t.Error("User-Agent was overridden by WithDefaultHeaders — should be ignored, use WithUserAgent instead")
	}
	if ct == "text/plain" {
		t.Error("Content-Type was overridden by WithDefaultHeaders — should be ignored")
	}
}

func TestWithDefaultHeaders_NilOrEmpty_OK(t *testing.T) {
	// Both should be no-ops, not errors.
	if _, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithDefaultHeaders(nil),
	); err != nil {
		t.Errorf("nil headers: %v", err)
	}
	if _, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithDefaultHeaders(map[string]string{}),
	); err != nil {
		t.Errorf("empty headers: %v", err)
	}
}

func TestWithDefaultHeaders_CallerMutationDoesNotAffectClient(t *testing.T) {
	var got string
	srv := fakeAuraServer(t, 200, `{"data":[]}`, func(r *http.Request) {
		got = r.Header.Get("Aura-Preview")
	})
	defer srv.Close()

	headers := map[string]string{"Aura-Preview": "true"}
	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
		aura.WithDefaultHeaders(headers),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Mutate the caller-supplied map after client construction. The client
	// must have copied the map and so should be unaffected.
	headers["Aura-Preview"] = "tampered"

	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got != "true" {
		t.Errorf("Aura-Preview = %q, want true (caller mutation should not affect client)", got)
	}
}

// =============================================================================
// WithHTTPClient
// =============================================================================

// countingTransport wraps another http.RoundTripper and counts every request
// it sees. We use it to prove WithHTTPClient actually swaps the transport.
type countingTransport struct {
	inner http.RoundTripper
	count int64
}

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	atomic.AddInt64(&c.count, 1)
	return c.inner.RoundTrip(r)
}

func TestWithHTTPClient_CustomTransportIsUsed(t *testing.T) {
	srv := fakeAuraServer(t, 200, `{"data":[]}`, nil)
	defer srv.Close()

	transport := &countingTransport{inner: http.DefaultTransport}
	custom := &http.Client{Transport: transport}

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
		aura.WithHTTPClient(custom),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Instances.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}

	// Two requests are expected: one to /oauth/token and one to /v1/instances.
	if got := atomic.LoadInt64(&transport.count); got < 2 {
		t.Errorf("custom transport saw %d requests, want >= 2", got)
	}
}

func TestWithHTTPClient_Nil_Errors(t *testing.T) {
	_, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithHTTPClient(nil),
	)
	if err == nil {
		t.Fatal("expected error for nil http client")
	}
}

// =============================================================================
// Sentinel errors
// =============================================================================

func TestSentinelErrors_MatchByStatusCode(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		sentinel error
	}{
		{"400 -> ErrBadRequest", http.StatusBadRequest, aura.ErrBadRequest},
		{"401 -> ErrUnauthorized", http.StatusUnauthorized, aura.ErrUnauthorized},
		{"403 -> ErrForbidden", http.StatusForbidden, aura.ErrForbidden},
		{"404 -> ErrNotFound", http.StatusNotFound, aura.ErrNotFound},
		{"409 -> ErrConflict", http.StatusConflict, aura.ErrConflict},
		{"429 -> ErrTooManyRequests", http.StatusTooManyRequests, aura.ErrTooManyRequests},
		{"500 -> ErrInternalServer", http.StatusInternalServerError, aura.ErrInternalServer},
		{"503 -> ErrServiceUnavail", http.StatusServiceUnavailable, aura.ErrServiceUnavail},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := fakeAuraServer(t, tc.status, `{"errors":[{"message":"boom"}]}`, nil)
			defer srv.Close()

			client, err := aura.NewClient(
				aura.WithCredentials("id", "secret"),
				aura.WithInsecureBaseURL(srv.URL),
			)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			_, err = client.Instances.List(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.sentinel) {
				t.Errorf("errors.Is(err, %v) = false; got err = %v", tc.sentinel, err)
			}
		})
	}
}

func TestSentinelErrors_AsExtractsAuraError(t *testing.T) {
	srv := fakeAuraServer(t, 404, `{"message":"instance not found","errors":[{"message":"no such id","reason":"missing"}]}`, nil)
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Instances.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *aura.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(err, *aura.Error) = false; err = %v", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Message != "instance not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "instance not found")
	}
	if !apiErr.IsNotFound() {
		t.Error("IsNotFound() = false")
	}
}

func TestSentinelErrors_NonMatchingStatus_NotIs(t *testing.T) {
	srv := fakeAuraServer(t, 404, `{}`, nil)
	defer srv.Close()

	client, err := aura.NewClient(
		aura.WithCredentials("id", "secret"),
		aura.WithInsecureBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Instances.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	// 404 must not match unrelated sentinels.
	if errors.Is(err, aura.ErrUnauthorized) {
		t.Error("404 should not match ErrUnauthorized")
	}
	if errors.Is(err, aura.ErrConflict) {
		t.Error("404 should not match ErrConflict")
	}
}
