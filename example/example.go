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
	*/

	// Delete an instance
	changeInstance, err := client.Instances.Delete("b12d2087")
	if err != nil {
		log.Fatalf("Failed to resume instance : %v", err)
	}
	fmt.Printf("Instance is resuming %s %s", changeInstance.Data.Name, changeInstance.Data.Status)

	fmt.Println("\n✓ Client is working correctly!")
}
