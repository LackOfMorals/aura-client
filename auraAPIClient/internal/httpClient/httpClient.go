package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Defines the interface for making HTTP Requests
type HTTPRequestExecutor interface {
	MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (*HTTPResponse, error)
}

// HTTP service
type HTTPService interface {
	HTTPRequestExecutor
}

// Stores the response from a request. Includes the original response
type HTTPResponse struct {
	ResponsePayload *[]byte
	RequestResponse *http.Response
}

// This is the concrete implementation for HTTP Service
type HTTPRequestsService struct {
	BaseURL string
	Timeout time.Duration
}

func NewHTTPRequestService(base string, timeout time.Duration) HTTPService {
	return &HTTPRequestsService{
		BaseURL: base,
		Timeout: timeout,
	}

}

// Convert map[string]string to http.Header (map[string][]string)
func toHTTPHeader(input map[string]string) http.Header {
	h := http.Header{}
	for k, v := range input {
		h[k] = []string{v}
	}
	return h
}

// Performs a http request, checks status code for ok and returns the response as a http.Response.
func (c *HTTPRequestsService) MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (response *HTTPResponse, err error) {

	// http client with timeout
	hClient := http.Client{Timeout: c.Timeout}

	endpointURL := c.BaseURL + endpoint

	// http new request requires the body to be in bytes.
	// convert string to bytes
	bodyBytes := []byte(body)

	// Create a request
	req, err := http.NewRequest(method, endpointURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// If we have header, apply it
	if header != nil {
		req.Header = toHTTPHeader(header)
	}

	req = req.WithContext(ctx)

	// Make the request
	resp, err := hClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}

	// ensure response body is closed when exit function
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Read the response payload into an array of bytes
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check HTTP status code and return an error including body on failure
	if err = checkResponse(resp, payload); err != nil {
		return nil, err
	}

	return &HTTPResponse{
		ResponsePayload: &payload,
		RequestResponse: resp,
	}, nil

}

// Check if the HTTP response to see if there was an error
func checkResponse(resp *http.Response, body []byte) error {
	if c := resp.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return fmt.Errorf("%s %s: %s - body: %s", resp.Request.Method, resp.Request.URL, resp.Status, string(body))
}
