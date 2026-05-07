// Package httpclient provides a low-level HTTP client with configurable retry
// behaviour. It is the transport layer beneath internal/api and has no knowledge
// of Aura-specific concepts such as base URLs, API versions, or authentication.
package httpclient

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// networkAndRateLimitRetryPolicy retries on connection-level errors and HTTP
// 429 Too Many Requests. All other HTTP responses are returned as-is so the
// api layer above can inspect the status code and decide what to do.
func networkAndRateLimitRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Context cancelled/deadline exceeded — do not retry.
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	// Network-level error with no HTTP response — retry.
	if err != nil && resp == nil {
		return true, nil
	}
	// Retry on 429: honour the server's rate-limit signal.
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}
	// Any other HTTP response — do not retry; the api layer decides.
	return false, nil
}

// retryAfterBackoff is a retryablehttp.Backoff that reads the Retry-After
// header on 429 responses and waits exactly that many seconds. For all other
// cases it falls back to the library's default exponential backoff. This
// prevents the SDK from hammering the server faster than it requested.
func retryAfterBackoff(minWait, maxWait time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if s := resp.Header.Get("Retry-After"); s != "" {
			if secs, err := time.ParseDuration(s + "s"); err == nil && secs > 0 {
				if secs > maxWait {
					return maxWait
				}
				return secs
			}
		}
	}
	return retryablehttp.DefaultBackoff(minWait, maxWait, attemptNum, resp)
}

// NewHTTPService creates a new HTTPService backed by a retryable HTTP client.
// Retries are attempted only on network-level errors (no response received);
// HTTP error responses (including 5xx) are always returned to the caller.
// The caller-supplied logger is used for debug output.
//
// If httpClient is non-nil, it is used as-is — the timeout argument is
// ignored and the caller is responsible for any timeout, transport, and
// connection-pool configuration. When httpClient is nil, a default client is
// constructed with production-suitable transport settings.
func NewHTTPService(timeout time.Duration, maxRetry int, logger *slog.Logger, httpClient *http.Client) HTTPService {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetry
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second
	retryClient.Logger = nil // suppress retryablehttp's own logger; we use slog
	retryClient.CheckRetry = networkAndRateLimitRetryPolicy
	retryClient.Backoff = retryAfterBackoff
	// PassthroughErrorHandler returns the final HTTP response instead of
	// discarding it when retries are exhausted. Without this, retryablehttp
	// replaces the last 429 response with a generic "giving up" error and the
	// api layer can never parse the status code into a typed *Error.
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler

	if httpClient != nil {
		retryClient.HTTPClient = httpClient
	} else {
		// Configure an explicit transport with production-suitable connection pool
		// settings. Go's default transport caps MaxIdleConnsPerHost at 2, which
		// causes connection exhaustion under concurrent load since all requests go
		// to the same host. These values are sized for a typical management-plane
		// workload; tune MaxIdleConnsPerHost upward if you issue many parallel calls.
		retryClient.HTTPClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}

	return &httpService{
		timeout: timeout,
		client:  retryClient,
		logger:  logger,
	}
}

// Close drains idle connections from the underlying HTTP connection pool.
// Safe to call from multiple goroutines and may be called more than once.
func (s *httpService) Close() {
	s.client.HTTPClient.CloseIdleConnections()
}

// Get performs an HTTP GET request with the provided headers.
func (s *httpService) Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodGet, url, headers, "")
}

// Post performs an HTTP POST request with the provided headers and body.
func (s *httpService) Post(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPost, url, headers, body)
}

// Put performs an HTTP PUT request with the provided headers and body.
func (s *httpService) Put(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPut, url, headers, body)
}

// Patch performs an HTTP PATCH request with the provided headers and body.
func (s *httpService) Patch(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPatch, url, headers, body)
}

// Delete performs an HTTP DELETE request with the provided headers.
func (s *httpService) Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodDelete, url, headers, "")
}

// doRequest is the shared implementation for all HTTP methods. It builds the
// request, attaches headers and the caller's context, executes it via the
// retryable client, and reads the response body up to DefaultMaxResponseSize.
func (s *httpService) doRequest(ctx context.Context, method, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := retryablehttp.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req = req.WithContext(ctx)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	s.logger.DebugContext(ctx, "executing HTTP request",
		slog.String("method", method),
		slog.String("url", url),
	)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	limitedReader := io.LimitReader(resp.Body, DefaultMaxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	// Detect truncation: if the body was exactly at the limit, check whether
	// there is more data waiting in the underlying reader.
	if int64(len(responseBody)) == DefaultMaxResponseSize {
		var probe [1]byte
		if n, _ := resp.Body.Read(probe[:]); n > 0 {
			return nil, fmt.Errorf("response body exceeds maximum allowed size of %d bytes", DefaultMaxResponseSize)
		}
	}

	s.logger.DebugContext(ctx, "HTTP response received",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", resp.StatusCode),
	)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
		Headers:    resp.Header,
	}, nil
}
