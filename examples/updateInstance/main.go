package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/LackOfMorals/aura-client"
)

const (
	AuraAPIBaseURL      = "https://api.neo4j.io/"
	AuraAPIAuthEndpoint = "oauth/token"
	AuraAPIV1           = "v1"
)

func main() {
	// Enable debug-level logging to stderr
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	ctx := context.Background()

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		slog.Error("failed to obtain environmental variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	myAuraClient, err := aura.NewClient(ClientID, ClientSecret)
	if err != nil {
		slog.Error("error obtaining NewClient", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("input the ID of the instance to get details of:")
	var instanceID string
	n, err := fmt.Scanln(&instanceID)
	if err != nil {
		slog.Error("error entering instance ID", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if n > 2 {
		slog.Error("only a single value can be entered for the Instance ID. You entered ", slog.Int("count: ", n))
		os.Exit(1)
	}

	if len(instanceID) != 8 {
		slog.Error("Instance ID is made up of 8 characters. You entered  ", slog.Int("count: ", len(instanceID)))
		os.Exit(1)

	}

	fmt.Printf("Enter the new name of the instance:")
	var instanceName string
	n, err = fmt.Scanln(&instanceName)
	if err != nil {
		slog.Error("error entering new instance name", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if n > 2 {
		slog.Error("only a single value can be entered for the new nanme. You entered ", slog.Int("count: ", n))
		os.Exit(1)
	}

	instanceNewCfg := aura.UpdateInstanceData{
		Name: instanceName,
	}

	response, err := myAuraClient.Instances.Update(ctx, instanceID, &instanceNewCfg)
	if err != nil {
		slog.Error("error reading instance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		slog.Error("error converting response to JSON", slog.String("error", err.Error()))
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
