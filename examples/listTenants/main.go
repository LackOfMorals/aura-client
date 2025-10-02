package main

import (
	"log"
	"os"

	auraAPIClient "github.com/LackOfMorals/aura-api-client/auraAPIClient"
)

func main() {

	myAuraClient := auraAPIClient.NewAuraAPIActionsService(AuraAPIBaseURL, AuraAPIV1, "120", ClientID, ClientSecret)

	auraToken, err := myAuraClient.GetAuthToken()

	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	// Get the list of tenants in the Aura Organisation
	auraTenants, err := myAuraClient.ListTenants(auraToken)

	if err != nil {
		log.Println("Error getting tenants: ", err)
		os.Exit(1)
	}

	// Display them in a table
	log.Println("List of tenants \n", auraTenants)
}
