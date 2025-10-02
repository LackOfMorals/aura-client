package main

import (
	"fmt"
	"log"
	"os"

	auraAPIClient "github.com/LackOfMorals/aura-api-client/auraAPIClient"
)

func main() {

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

	myAuraClient := auraAPIClient.NewAuraAPIActionsService(AuraAPIBaseURL, AuraAPIV1, "120", ClientID, ClientSecret)

	auraToken, err := myAuraClient.GetAuthToken()
	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	auraTenants, err := myAuraClient.ListTenants(auraToken)
	if err != nil {
		log.Println("Error getting tenant details: ", err)
		os.Exit(1)
	}

	instanceCfg := auraAPIClient.CreateInstanceConfigData{
		Name:          instanceName,
		Version:       "5",
		Region:        "europe-west1",
		Memory:        "8GB",
		Type:          "enterprise-db",
		TenantId:      auraTenants.Data[0].Id,
		CloudProvider: "gcp",
	}

	newInstance, err := myAuraClient.CreateInstance(auraToken, &instanceCfg)
	if err != nil {
		log.Println("Error creating instance: ", err)
		os.Exit(1)
	}

	log.Printf("Instance created: \n ID %s \n Name %s \n URI %s \n User: %s \n Pwd: %s \n", newInstance.Data.Id, newInstance.Data.Name, newInstance.Data.ConnectionUrl, newInstance.Data.Username, newInstance.Data.Password)

}
