package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/aura-client/resources"
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

	fmt.Printf("Enter the name of the instance to create:")
	var instanceName string
	n, err := fmt.Scanln(&instanceName)
	if err != nil {
		log.Println("Error entering instance name: ", err)
		os.Exit(1)
	}

	if n > 2 {
		log.Println("Only a single value can be entered for the instance name. You entered ", n)
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

	auraTenants, err := myAuraClient.Tenants.List(ctx, auraToken)
	if err != nil {
		log.Println("Error getting tenant details: ", err)
		os.Exit(1)
	}

	instanceCfg := resources.CreateInstanceConfigData{
		Name:          instanceName,
		Version:       "5",
		Region:        "europe-west1",
		Memory:        "8GB",
		Type:          "enterprise-db",
		TenantId:      auraTenants.Data[0].Id,
		CloudProvider: "gcp",
	}

	response, err := myAuraClient.Instances.Create(ctx, auraToken, &instanceCfg)
	if err != nil {
		log.Println("Error creating instance: ", err)
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println("Error formatting response: ", err)
		os.Exit(1)
	}

	log.Printf("Details of instance being created: %s", result)
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
