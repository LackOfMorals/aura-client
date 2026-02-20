package api

// Response represents a response from the Aura API
type Response struct {
	StatusCode int
	Body       []byte
}

// Error represents an error response from the Aura API
type Error struct {
	StatusCode int           `json:"status_code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail represents individual error details
type ErrorDetail struct {
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
	Field   string `json:"field,omitempty"`
}
