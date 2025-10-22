package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/LackOfMorals/aura-client"
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

	fmt.Printf("input the ID of the instance:")
	var instanceID string
	n, err := fmt.Scanln(&instanceID)
	if err != nil {
		log.Println("Error entering instance ID: ", err)
		os.Exit(1)
	}

	if n > 2 {
		log.Println("Only a single value can be entered for the Instance ID. You entered ", n)
		os.Exit(1)
	}
	if len(instanceID) > 8 {
		log.Println("Instance ID can only be 8 characters. You entered ", len(instanceID))
		os.Exit(1)

	}

	myAuraClient, err := aura.NewAuraAPIActionsService(ClientID, ClientSecret)
	if err != nil {
		log.Println("Error creating aura client: ", err)
		os.Exit(1)
	}

	auraToken, err := myAuraClient.Auth.GetAuthToken(ctx)
	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	response, err := myAuraClient.Snapshots.List(ctx, auraToken, instanceID, "2021-08-15")
	if err != nil {
		log.Println("Error reading snapshots: ", err)
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println("Error formatting response: ", err)
		os.Exit(1)
	}

	log.Printf("Details of snapshots: %s", result)

	response1, err1 := myAuraClient.Snapshots.Create(ctx, auraToken, instanceID)
	if err1 != nil {
		log.Println("Error creating snapshots: ", err1)
		os.Exit(1)
	}

	result, err = json.MarshalIndent(response1, "", "  ")
	if err != nil {
		log.Println("Error formatting response: ", err)
		os.Exit(1)
	}

	log.Printf("New snapshots: %s", result)

}

func readClientIDAndSecretFromEnv() (string, string, error) {
	var ClientID, ClientSecret string
	var found bool

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
