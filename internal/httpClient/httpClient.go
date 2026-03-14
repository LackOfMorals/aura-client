package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/LackOfMorals/aura-client/internal/utils"
	"github.com/hashicorp/go-retryablehttp"
)

// NewHTTPService creates a new HTTPService with the specified configuration.
func NewHTTPService(baseURL string, timeout time.Duration, maxRetry int, logger *slog.Logger) HTTPService {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetry
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.HTTPClient.Timeout = timeout
	retryClient.Logger = &slogAdapter{logger: logger}

	return &httpService{
		baseURL: baseURL,
		timeout: timeout,
		client:  retryClient,
		logger:  logger,
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
// endpoint is the complete path to the endpoint that will be called
func (s *httpService) doRequest(ctx context.Context, method, endpoint string, headers map[string]string, body string) (*HTTPResponse, error) {

	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		s.logger.DebugContext(ctx, "failed to create request",
			slog.String("method", method),
			slog.String("url", endpoint),
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
		slog.String("url", endpoint),
	)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.DebugContext(ctx, "request failed",
			slog.String("method", method),
			slog.String("url", endpoint),
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
			slog.String("url", endpoint),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	// Log response from request.  body is limited to 200 bytes to avoid flooding the log.
	s.logger.DebugContext(ctx, "HTTP request completed",
		slog.String("method", method),
		slog.String("url", endpoint),
		slog.Int("statusCode", resp.StatusCode),
		slog.Int("bodySize", len(respBody)),
		slog.String("bodyPreview", utils.TruncateString(string(respBody), 200)),
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
