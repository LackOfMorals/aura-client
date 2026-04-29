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

// NewHTTPService creates a new HTTPService backed by a retryable HTTP client.
// The client enforces the given timeout per request and retries up to maxRetry
// times on transient failures. The caller-supplied logger is used for debug
// output; pass a warn-level logger to suppress verbose output in production.
func NewHTTPService(timeout time.Duration, maxRetry int, logger *slog.Logger) HTTPService {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetry
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second
	retryClient.Logger = nil // suppress retryablehttp's own logger; we use slog

	retryClient.HTTPClient = &http.Client{
		Timeout: timeout,
	}

	return &httpService{
		timeout: timeout,
		client:  retryClient,
		logger:  logger,
	}
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
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, DefaultMaxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
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
