package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// instance status values that we're interested in
const (
	StatusRunning   = "running"
	StatusStopped   = "stopped"
	StatusPaused    = "paused"
	StatusAvailable = "available"
)

// ListInstancesResponse contains a list of instances in a tenant
type ListInstancesResponse struct {
	Data []ListInstanceData `json:"data"`
}

type ListInstanceData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Created       string `json:"created_at"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
}

type CreateInstanceConfigData struct {
	Name          string `json:"name"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Version       string `json:"version"`
	Memory        string `json:"memory"`
}

type CreateInstanceResponse struct {
	Data CreateInstanceData `json:"data"`
}

type CreateInstanceData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	ConnectionUrl string `json:"connection_url"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type UpdateInstanceData struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

type GetInstanceResponse struct {
	Data InstanceData `json:"data"`
}

type DeleteInstanceResponse struct {
	Data InstanceData `json:"data"`
}

type InstanceData struct {
	Id              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	TenantId        string  `json:"tenant_id"`
	CloudProvider   string  `json:"cloud_provider"`
	ConnectionUrl   string  `json:"connection_url"`
	Region          string  `json:"region"`
	Type            string  `json:"type"`
	Memory          string  `json:"memory"`
	Storage         *string `json:"storage"`
	CDCEnrichment   string  `json:"cdc_enrichment_mode"`
	GDSPlugin       bool    `json:"graph_analytics_plugin"`
	MetricsURL      string  `json:"metrics_integration_url"`
	Secondaries     int     `json:"secondaries_count"`
	VectorOptimized bool    `json:"vector_optimized"`
}

type overwriteInstanceRequest struct {
	SourceInstanceId string `json:"source_instance_id,omitempty"`
	SourceSnapshotId string `json:"source_snapshot_id,omitempty"`
}

type OverwriteInstanceResponse struct {
	Data string `json:"data"`
}

// instanceService handles instance operations
type instanceService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
