package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// DefaultMaxResponseSize is the maximum size of response body to read (10MB)
	DefaultMaxResponseSize = 10 * 1024 * 1024
)

// HTTPRequestExecutor defines the interface for making HTTP requests.
type HTTPRequestExecutor interface {
	MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (*HTTPResponse, error)
}

// HTTPService defines the interface for HTTP operations.
type HTTPService interface {
	HTTPRequestExecutor
}

// HTTPResponse stores the response from a request, including the payload and original response.
type HTTPResponse struct {
	ResponsePayload *[]byte
	RequestResponse *http.Response
}

// HTTPRequestsService is the concrete implementation of HTTPService.
// It handles HTTP requests with configurable timeouts and connection pooling.
type HTTPRequestsService struct {
	BaseURL string
	Timeout time.Duration
	client  *retryablehttp.Client
	logger  *slog.Logger
}

// NewHTTPRequestService creates a new HTTPService with the specified base URL and timeout.
// The service allows for setting of a custom retry policy
func NewHTTPRequestService(base string, timeout time.Duration, maxRetry int, logger *slog.Logger) HTTPService {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetry
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.HTTPClient.Timeout = timeout

	// Integrate slog logger with retryablehttp
	retryClient.Logger = &slogAdapter{logger: logger}

	return &HTTPRequestsService{
		BaseURL: base,
		Timeout: timeout,
		client:  retryClient,
		logger:  logger,
	}
}

// toHTTPHeader converts a map[string]string to http.Header format (map[string][]string).
func toHTTPHeader(input map[string]string) http.Header {
	h := http.Header{}
	for k, v := range input {
		h[k] = []string{v}
	}
	return h
}

// MakeRequest performs an HTTP request with the specified parameters.
// It validates the response status code and returns an error if the response indicates failure.
// The response body is limited to DefaultMaxResponseSize to prevent memory exhaustion.
// MakeRequest performs an HTTP request with automatic retry logic.
func (c *HTTPRequestsService) MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (*HTTPResponse, error) {
	endpointURL := c.BaseURL + endpoint

	// Create request body reader
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	// Create the retryable HTTP request
	req, err := retryablehttp.NewRequestWithContext(ctx, method, endpointURL, bodyReader)
	if err != nil {
		c.logger.DebugContext(ctx, "failed to create request", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers if provided
	if header != nil {
		req.Header = toHTTPHeader(header)
	}

	// Execute the request (with automatic retries)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.DebugContext(ctx, "request failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Ensure response body is closed
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			c.logger.DebugContext(ctx, "failed to close response body", slog.String("error", cerr.Error()))
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	// Read the response payload with size limit
	payload, err := io.ReadAll(io.LimitReader(resp.Body, int64(DefaultMaxResponseSize)))
	if err != nil {
		c.logger.DebugContext(ctx, "failed to read response body", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Validate HTTP response status code
	if err = checkResponse(resp, payload); err != nil {
		c.logger.DebugContext(ctx, "response status code was not 2XX", slog.String("error", err.Error()))
		return &HTTPResponse{
			ResponsePayload: &payload,
			RequestResponse: resp,
		}, err
	}

	return &HTTPResponse{
		ResponsePayload: &payload,
		RequestResponse: resp,
	}, nil
}

// checkResponse validates the HTTP response status code.
// It returns an error if the status code is outside the 2xx success range.
func checkResponse(resp *http.Response, body []byte) error {

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("HTTP %d %s %s: %s - body: %s",
		resp.StatusCode,
		resp.Request.Method,
		resp.Request.URL,
		resp.Status,
		string(body),
	)
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
