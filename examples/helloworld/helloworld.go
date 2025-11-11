package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/LackOfMorals/aura-client"
)

func main() {

	// Read ClientID, ClientSecret from env vars of the same name
	ClientID, ClientSecret, err := readClientIDAndSecretFromEnv()
	if err != nil {
		slog.Error("failed to obtain environmental variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	myAuraClient, err := aura.NewClient(aura.WithCredentials(ClientID, ClientSecret))

	if err != nil {
		slog.Error("error obtaining NewClient", slog.String("error", err.Error()))
		os.Exit(1)
	}

	response, err := myAuraClient.Instances.List()
	if err != nil {
		slog.Error("error getting instance list", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Print out results
	// In a table

	tw := new(tabwriter.Writer)
	tw.Init(os.Stdout, 16, 8, 4, '\t', 0)
	// Header
	fmt.Fprintln(tw, "Name\tId   \tTenant Id\tCloud Provider\tCreated\t")
	for _, r := range response.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t \n", r.Name, r.Id, r.TenantId, r.CloudProvider, r.Created)
	}
	tw.Flush()

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
