package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// Customer Managed Encryption Keys

// GetCmeksResponse contains a list of customer managed encryption keys
type GetCmeksResponse struct {
	Data []GetCmeksData `json:"data"`
}

type GetCmeksData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	TenantId string `json:"tenant_id"`
}

// cmekService handles customer managed encryption key operations
type cmekService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
