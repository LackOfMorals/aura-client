package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
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
	client  *http.Client
}

// NewHTTPRequestService creates a new HTTPService with the specified base URL and timeout.
// The service reuses an HTTP client for connection pooling efficiency.
func NewHTTPRequestService(base string, timeout time.Duration) HTTPService {
	return &HTTPRequestsService{
		BaseURL: base,
		Timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
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
func (c *HTTPRequestsService) MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (*HTTPResponse, error) {
	endpointURL := c.BaseURL + endpoint

	// Create request body reader, only if body is not empty
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers if provided
	if header != nil {
		req.Header = toHTTPHeader(header)
	}

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Ensure response body is closed when exiting function
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	// Read the response payload with size limit to prevent memory exhaustion
	payload, err := io.ReadAll(io.LimitReader(resp.Body, int64(DefaultMaxResponseSize)))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Validate HTTP response status code
	if err = checkResponse(resp, payload); err != nil {
		return nil, err
	}

	return &HTTPResponse{
		ResponsePayload: &payload,
		RequestResponse: resp,
	}, nil
}

// checkResponse validates the HTTP response status code.
// It returns an error if the status code is outside the 2xx success range.
func checkResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusNoContent {
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
