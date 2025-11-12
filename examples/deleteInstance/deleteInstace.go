package main

import (
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

	fmt.Printf("input the ID of the instance to delete:")
	var instanceID string
	n, err := fmt.Scanln(&instanceID)
	if err != nil {
		slog.Error("error entering instance id ", slog.String("error", err.Error()))
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

	myAuraClient, err := aura.NewClient(aura.WithCredentials(ClientID, ClientSecret))
	if err != nil {
		slog.Error("error creating aura client: ", slog.String("error", err.Error()))
		os.Exit(1)
	}

	response, err := myAuraClient.Instances.Delete(instanceID)

	if err != nil {
		slog.Error("error deleting instance: ", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Details of instance being deleted: %+v \n", response.Data)

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
