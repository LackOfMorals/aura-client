// types.go - Exported error types for use by consumers
package aura

import (
	"github.com/LackOfMorals/aura-client/internal/api"
)

// ============================================================================
// Error Types
// ============================================================================

// Error represents an error response from the Aura API
type Error = api.Error

// ErrorDetail represents individual error details
type ErrorDetail = api.ErrorDetail
