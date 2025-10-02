package httpClient

// Defines the interface for making HTTP Requests
type HTTPRequestExecutor interface {
	MakeRequest(endpoint string, method string, header map[string][]string, body []byte) (*HTTPResponse, error)
}

// HTTP service
type HTTPService interface {
	HTTPRequestExecutor
}
