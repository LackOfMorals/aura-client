package main

import (
	"errors"
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

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		slog.Error("failed to obtain environmental variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	myAuraClientConfig, _ := aura.NewClient().WithCredentials(ClientID, ClientSecret).WithURL(AuraAPIBaseURL)

	//myAuraClientConfig := aura.NewClientConfig().WithDefaults(ClientID, ClientSecret)

	slog.Error("", slog.String("client id", myAuraClientConfig.ClientID))

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
