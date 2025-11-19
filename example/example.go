package main

import (
	"fmt"
	"log"
	"os"
	"time"

	aura "github.com/LackOfMorals/aura-client"
)

func main() {
	// Load credentials from environment
	clientID := os.Getenv("AURA_CLIENT_ID")
	clientSecret := os.Getenv("AURA_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Missing required environment variables")
	}

	// Create client
	client, err := aura.NewClient(
		aura.WithCredentials(clientID, clientSecret),
		aura.WithTimeout(120*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	/*
			// List existing instances

			fmt.Println("=== Current Instances ===")
			instances, err := client.Instances.List()
			if err != nil {
				log.Fatalf("Failed to list instances: %v", err)
			}

			for _, inst := range instances.Data {
				fmt.Printf("- %s: %s (%s)   \n",
					inst.Name,
					inst.Id,
					inst.CloudProvider,
				)

			}

			// get the details of an instance
			for _, inst := range instances.Data {
				instanceDetails, err := client.Instances.Get(string(inst.Id))
				if err != nil {
					log.Fatalf("Failed to get instance details: %v", err)
				}
				fmt.Printf("- %s: %s (%s)   \n",
					instanceDetails.Data.Id,
					instanceDetails.Data.Status,
					instanceDetails.Data.Name,
				)
			}


		// Delete an instance
		changeInstance, err := client.Instances.Delete("b12d2087")
		if err != nil {
			log.Fatalf("Failed to resume instance : %v", err)
		}
		fmt.Printf("Instance is resuming %s %s", changeInstance.Data.Name, changeInstance.Data.Status)
	*/

	auraTenants, err := client.Tenants.List()
	if err != nil {
		log.Fatalf("Failed to get tenant : %v", err)
		os.Exit(1)
	}

	instanceCfg := aura.CreateInstanceConfigData{
		Name:          "jg-myfirst-instance",
		Version:       "5",
		Region:        "europe-west1",
		Memory:        "16GB",
		Type:          "enterprise-db",
		TenantId:      auraTenants.Data[0].Id,
		CloudProvider: "gcp",
	}

	newInst, err := client.Instances.Create(&instanceCfg)
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Instance creating \n name %s \n id %s \n ", newInst.Data.Name, newInst.Data.Id)

	fmt.Println("\n✓ Client is working correctly!")
}
