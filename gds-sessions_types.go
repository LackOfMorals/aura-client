package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// GetGDSSessionListResponse contains a list of GDS sessions
type GetGDSSessionListResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

// GetGDSSessionResponse contains information about a single GDS Session
type GetGDSSessionResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

type GetGDSSessionData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Memory        string `json:"memory"`
	InstanceId    string `json:"instance_id"`
	DatabaseId    string `json:"database_uuid"`
	Status        string `json:"status"`
	Create        string `json:"created_at"`
	Host          string `json:"host"`
	Expiry        string `json:"expiry_date"`
	Ttl           string `json:"ttl"`
	UserId        string `json:"user_id"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
}

type CreateGDSSessionConfigData struct {
	Name          string `json:"name"`
	Ttl           string `json:"ttl"`
	TenantId      string `json:"tenant_id"`
	InstanceId    string `json:"instance_id"`
	DatabaseId    string `json:"database_uuid"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Memory        string `json:"memory"`
}

type GetGDSSessionSizeEstimation struct {
	NodeCount                 int      `json:"node_count"`
	NodePropertyCount         int      `json:"node_property_count"`
	NodeLabelCount            int      `json:"node_label_count"`
	RelationshipCount         int      `json:"relationship_count"`
	RelationshipPropertyCount int      `json:"relationship_property_count"`
	AlgorithmCategories       []string `json:"algorithm_categories"`
}

type GDSSessionSizeEstimationResponse struct {
	Data GDSSessionSizeEstimationData `json:"data"`
}

type GDSSessionSizeEstimationData struct {
	EstimatedMemory string `json:"estimated_memory"`
	RecommendedSize string `json:"recommended_size"`
}

type DeleteGDSSessionResponse struct {
	Data DeleteGDSSession `json:"data"`
}

type DeleteGDSSession struct {
	ID string `json:"id"`
}

// gDSSessionService handles GDS Session operations
type gDSSessionService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
