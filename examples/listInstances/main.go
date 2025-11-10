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
	logger := slog.New(handler)

	ctx := context.Background()

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		slog.Error("failed to obtain environmental variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	myAuraClient, err := aura.NewClient(aura.WithCredentials(ClientID, ClientSecret), aura.WithLogger(logger))

	if err != nil {
		slog.Error("error obtaining NewClient", slog.String("error", err.Error()))
		os.Exit(1)
	}

	response, err := myAuraClient.Instances.List(ctx)
	if err != nil {
		slog.Error("error getting instance list", slog.String("error", err.Error()))
		os.Exit(1)
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		slog.Error("error converting response to JSON", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Instances: %s", result)

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
