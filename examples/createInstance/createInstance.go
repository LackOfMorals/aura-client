/*
 This shows how to create an instance in Aura.
 After asking for the name to use for the instance, it looks up the tenants that are accessible
 and then uses the id of the first entry in that list.
 The new instance will be created in GCP in europe-west1 with 8Gb of memory.

 The output shows the details of the new instance that is being created in JSON format.


*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/LackOfMorals/aura-client"
)

func main() {

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		slog.Error("failed to obtain environmental variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Enter the name of the instance to create:")
	var instanceName string
	n, err := fmt.Scanln(&instanceName)
	if err != nil {
		slog.Error("error getting name of instance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if n > 2 {
		slog.Error("only a single value can be entered", slog.String("error", ""))
		os.Exit(1)
	}

	myAuraClient, err := aura.NewClient(aura.WithCredentials(ClientID, ClientSecret))
	if err != nil {
		slog.Error("error creating aura client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	auraTenants, err := myAuraClient.Tenants.List()
	if err != nil {
		slog.Error("error getting tenant details", slog.String("error", err.Error()))
		os.Exit(1)
	}

	instanceCfg := aura.CreateInstanceConfigData{
		Name:          instanceName,
		Version:       "5",
		Region:        "europe-west1",
		Memory:        "16GB",
		Type:          "enterprise-db",
		TenantId:      auraTenants.Data[0].Id,
		CloudProvider: "gcp",
	}

	response, err := myAuraClient.Instances.Create(&instanceCfg)
	if err != nil {
		slog.Error("error creating instance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		slog.Error("error formatting response", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Details of instance being created: %s", result)
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
