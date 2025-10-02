package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	auraAPIClient "github.com/LackOfMorals/aura-api-client/auraAPIClient"
)

const (
	AuraAPIBaseURL      = "https://api.neo4j.io/"
	AuraAPIAuthEndpoint = "oauth/token"
	AuraAPIV1           = "v1"
)

func main() {

	ctx := context.Background()

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		log.Println("Unable to obtain values for authentication to Aura API: ", err)
		os.Exit(1)
	}

	myAuraClient := auraAPIClient.NewAuraAPIActionsService(AuraAPIBaseURL, AuraAPIV1, "120", ClientID, ClientSecret)

	auraToken, err := myAuraClient.GetAuthToken()

	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	// Get the list of tenants in the Aura Organisation
	response, err := myAuraClient.ListTenants(ctx, auraToken)

	if err != nil {
		log.Println("Error getting tenants: ", err)
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println("Error formatting response: ", err)
		os.Exit(1)
	}

	log.Printf("Tenants details: %s", result)

}

func readClientIDAndSecretFromEnv() (string, string, error) {
	var ClientID, ClientSecret string
	var found bool

	// Is set by LookupEnv to true if the environmental variable is found
	found = false

	// See if environmantal variables are present and get their value if so
	// set found to true if this is the case
	ClientID, found = os.LookupEnv("ClientID")
	if !found {
		return "", "", errors.New("environmental variable ClientID not set")
	}

	ClientSecret, found = os.LookupEnv("ClientSecret")
	if !found {
		return "", "", errors.New("environmental variable ClientSecret not set")
	}

	return ClientID, ClientSecret, nil
}
