package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Stores the response from a request. Includes the original response
type HTTPResponse struct {
	ResponsePayload *[]byte
	RequestResponse *http.Response
}

// This is the concrete implementation for HTTP Service
type HTTPRequestsService struct {
	BaseURL string
	Timeout string
}

func NewHTTPRequestService(base, timeout string) HTTPService {
	return &HTTPRequestsService{
		BaseURL: base,
		Timeout: timeout,
	}

}

// Performs a http request, checks status code for ok and returns the response as a http.Response.
func (c *HTTPRequestsService) MakeRequest(endpoint string, method string, header map[string][]string, body []byte) (*HTTPResponse, error) {

	var hClient http.Client

	ctx := context.Background()

	endpointURL := c.BaseURL + endpoint

	// Create a request
	req, err := http.NewRequest(method, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// If we have header, apply it
	if header != nil {
		req.Header = header
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

	// Check on http status code for
	// indication of failed request
	err = checkResponse(resp)
	if err != nil {
		return nil, err
	}

	// Holds the response body as an array of bytes
	var payload []byte

	// Read the response payload into a array of bytes
	payload, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body into array of bytes: \n", err)
		return nil, err
	}

	// ensure response body is closed when exit function
	defer resp.Body.Close()

	return &HTTPResponse{
		ResponsePayload: &payload,
		RequestResponse: resp,
	}, err

}

// Check if the HTTP response to see if there was an error
func checkResponse(resp *http.Response) error {
	if c := resp.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return fmt.Errorf("%s %s: %s", resp.Request.Method, resp.Request.URL, resp.Status)
}
