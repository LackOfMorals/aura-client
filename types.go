// types.go - Exported error types for use by consumers
package aura

import (
	"github.com/LackOfMorals/aura-client/internal/api"
)

// ============================================================================
// Error Types
// ============================================================================

// APIError represents an error response from the Aura API
type APIError = api.APIError

// APIErrorDetail represents individual error details
type APIErrorDetail = api.APIErrorDetail
