package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	fmt.Printf("input the ID of the instance to delete:")
	var instanceID string
	n, err := fmt.Scanln(&instanceID)
	if err != nil {
		log.Println("Error entering instance ID to delete: ", err)
		os.Exit(1)
	}

	if n > 2 {
		log.Println("Only a single value can be entered for the Instance ID. You entered ", n)
		os.Exit(1)
	}

	if len(instanceID) != 8 {
		log.Println("Instance ID is made up of 8 characters. You entered ", len(instanceID))
		os.Exit(1)

	}

	myAuraClient := auraAPIClient.NewAuraAPIActionsService(AuraAPIBaseURL, AuraAPIV1, "120", ClientID, ClientSecret)

	auraToken, err := myAuraClient.Auth.GetAuthToken(ctx)
	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	response, err := myAuraClient.Instances.Delete(ctx, auraToken, instanceID)
	if err != nil {
		log.Println("Error deleting instance: ", err)
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println("Error formatting response: ", err)
		os.Exit(1)
	}

	log.Printf("Details of instance being deleted: %s", result)

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
