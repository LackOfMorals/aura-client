package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// DefaultMaxResponseSize is the maximum size of response body to read (10MB)
	DefaultMaxResponseSize = 10 * 1024 * 1024
)

// HTTPResponse stores the response from a request, including the payload and original response.
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// HTTPService defines the interface for HTTP operations.
// This is the low-level HTTP layer that handles raw HTTP requests.
type HTTPService interface {
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error)
	Put(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error)
	Patch(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error)
	Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
}

// httpService is the concrete implementation of HTTPService.
// It handles HTTP requests with configurable timeouts, retries, and connection pooling.
type httpService struct {
	baseURL    string
	apiVersion string
	timeout    time.Duration
	client     *retryablehttp.Client
	logger     *slog.Logger
}

// NewHTTPService creates a new HTTPService with the specified configuration.
func NewHTTPService(baseURL string, apiVersion string, timeout time.Duration, maxRetry int, logger *slog.Logger) HTTPService {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetry
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.HTTPClient.Timeout = timeout
	retryClient.Logger = &slogAdapter{logger: logger}

	return &httpService{
		baseURL:    baseURL,
		apiVersion: apiVersion,
		timeout:    timeout,
		client:     retryClient,
		logger:     logger,
	}
}

// Get performs an HTTP GET request.
func (s *httpService) Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodGet, url, headers, "")
}

// Post performs an HTTP POST request.
func (s *httpService) Post(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPost, url, headers, body)
}

// Put performs an HTTP PUT request.
func (s *httpService) Put(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPut, url, headers, body)
}

// Patch performs an HTTP PATCH request.
func (s *httpService) Patch(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodPatch, url, headers, body)
}

// Delete performs an HTTP DELETE request.
func (s *httpService) Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	return s.doRequest(ctx, http.MethodDelete, url, headers, "")
}

// doRequest performs the actual HTTP request with the specified parameters.
// It automatically detects whether the endpoint is a full URL (starting with http:// or https://)
// or a relative path. Full URLs are used as-is, while relative paths are appended to the base URL.
func (s *httpService) doRequest(ctx context.Context, method, endpoint string, headers map[string]string, body string) (*HTTPResponse, error) {
	// Determine if endpoint is already a full URL
	var fullURL string
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// Endpoint is already a full URL, use it as-is
		fullURL = endpoint
	} else {
		// Endpoint is relative, prepend base URL
		fullURL = s.baseURL + "/" + s.apiVersion + "/" + endpoint
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		s.logger.DebugContext(ctx, "failed to create request",
			slog.String("method", method),
			slog.String("url", fullURL),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	s.logger.DebugContext(ctx, "executing HTTP request",
		slog.String("method", method),
		slog.String("url", fullURL),
	)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.DebugContext(ctx, "request failed",
			slog.String("method", method),
			slog.String("url", fullURL),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body with size limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, int64(DefaultMaxResponseSize)))
	if err != nil {
		s.logger.DebugContext(ctx, "failed to read response body",
			slog.String("method", method),
			slog.String("url", fullURL),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	s.logger.DebugContext(ctx, "HTTP request completed",
		slog.String("method", method),
		slog.String("url", fullURL),
		slog.Int("statusCode", resp.StatusCode),
		slog.String("body", string(respBody)),
	)

	return httpResp, nil
}

// slogAdapter adapts slog.Logger to retryablehttp.LeveledLogger interface
type slogAdapter struct {
	logger *slog.Logger
}

func (s *slogAdapter) Error(msg string, keysAndValues ...any) {
	s.logger.Error(msg, keysAndValues...)
}

func (s *slogAdapter) Info(msg string, keysAndValues ...any) {
	s.logger.Info(msg, keysAndValues...)
}

func (s *slogAdapter) Debug(msg string, keysAndValues ...any) {
	s.logger.Debug(msg, keysAndValues...)
}

func (s *slogAdapter) Warn(msg string, keysAndValues ...any) {
	s.logger.Warn(msg, keysAndValues...)
}
