// package auraAPIClient provides functionality to use the Neo4j Aura API to provision, managed and then destory Aura instances
package auraAPIClient

// These are the interfaces that represent the functions available in this package

type GetAuthTokenExecutor interface {
	GetAuthToken() (*AuthAPIToken, error)
}

type ListTenantsExecutor interface {
	ListTenants(*AuthAPIToken) (*ListTenantsResponse, error)
}

type GetTenantExecutor interface {
	GetTenant(*AuthAPIToken, string) (*GetTenantResponse, error)
}

type ListInstancesExecutor interface {
	ListInstances(*AuthAPIToken) (*ListInstancesResponse, error)
}

type CreateInstanceExecutor interface {
	CreateInstance(*AuthAPIToken, *CreateInstanceConfigData) (*CreateInstanceResponse, error)
}

type DeleteInstanceExecutor interface {
	DeleteInstance(*AuthAPIToken, string) (*GetInstanceResponse, error)
}

type GetInstanceExecutor interface {
	GetInstance(*AuthAPIToken, string) (*GetInstanceResponse, error)
}

// Aura API service
type AuraAPIService interface {
	GetAuthTokenExecutor
	ListTenantsExecutor
	GetTenantExecutor
	ListInstancesExecutor
	CreateInstanceExecutor
	DeleteInstanceExecutor
	GetInstanceExecutor
}

// This is the concrete implementation for Aura API Service
type AuraAPIActionsService struct {
	AuraAPIBaseURL string
	AuraAPIVersion string
	AuraAPITimeout string
	ClientID       string
	ClientSecret   string
}

// NewDriver is the entry point to the auraClient driver to create an instance of a Driver. It is the first function to
// be called in order to establish a connection to a neo4j database. It requires the Aura API base URL, the version of the
// Aura API to use, client id, and client secret.
func NewAuraAPIActionsService(baseurl, ver, timeout, id, sec string) AuraAPIService {
	return &AuraAPIActionsService{
		AuraAPIBaseURL: baseurl,
		AuraAPIVersion: ver,
		AuraAPITimeout: timeout,
		ClientID:       id,
		ClientSecret:   sec,
	}
}
