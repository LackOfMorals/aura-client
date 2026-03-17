package aura

// Customer Managed Encryption Keys
// service structs are in the service go file

// GetCmeksResponse contains a list of customer managed encryption keys
type GetCmeksResponse struct {
	Data []GetCmeksData `json:"data"`
}

type GetCmeksData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}
