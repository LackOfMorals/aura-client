package main

import (
	"fmt"
	"log"
	"os"

	auraAPIClient "github.com/LackOfMorals/aura-api-client/auraAPIClient"
)

func main() {

	fmt.Printf("input the ID of the instance to delete:")
	var instanceID string
	n, err := fmt.Scanln(&instanceID)
	if err != nil {
		log.Println("Error entering instance ID to delete: ", err)
		os.Exit(1)
	}

	if n > 2 {
		log.Println("Only a single value can be entered for the Instance ID. You entered ", n)
		os.Exit(1)
	}

	if len(instanceID) != 8 {
		log.Println("Instance ID is made up of 8 characters. You entered ", len(instanceID))
		os.Exit(1)

	}

	myAuraClient := auraAPIClient.NewAuraAPIActionsService(AuraAPIBaseURL, AuraAPIV1, "120", ClientID, ClientSecret)

	auraToken, err := myAuraClient.GetAuthToken()
	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	delInstance, err := myAuraClient.DeleteInstance(auraToken, instanceID)
	if err != nil {
		log.Println("Error deleting instance: ", err)
		os.Exit(1)
	}

	log.Printf("Instance deleted: ID %s, Name %s", delInstance.Data.Id, delInstance.Data.Name)

}
