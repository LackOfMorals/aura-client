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

	myAuraClient, err := aura.NewClient(ClientID, ClientSecret)
	if err != nil {
		slog.Error("error creating aura client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	auraTenants, err := myAuraClient.Tenants.List(ctx)
	if err != nil {
		slog.Error("error getting tenant details", slog.String("error", err.Error()))
		os.Exit(1)
	}

	instanceCfg := aura.CreateInstanceConfigData{
		Name:          instanceName,
		Version:       "5",
		Region:        "europe-west1",
		Memory:        "8GB",
		Type:          "enterprise-db",
		TenantId:      auraTenants.Data[0].Id,
		CloudProvider: "gcp",
	}

	response, err := myAuraClient.Instances.Create(ctx, &instanceCfg)
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
