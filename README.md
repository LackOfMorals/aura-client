# Aura API Client

## Overview
There's a few occasions where I have wanted to use the Aura API  with Go based applications.  Rather than re-write the integration each and everytime, it would be nice to have module that can be re-used.  Hence I have put together the Aura API Client. 

My Go knowledge is still being acquired - it's something of a hobby. I've looped around writing Go, trying it, improving, seeing difference directions to take and have had chats with Claude to fill holes in my knowledge.  It's been a great learning experience.   The point I am making - this has not been authored by a professional developer; there are rough edges. 

## Installation

Get the module
```bash
go get github.com/LackOfMorals/aura-api-client
```

## Usage
The modules follows a pattern of 

- Instantiate a client with .NewAuraAPIActionsService
- With the client, get a token to use with the aura api with .Auth
- Use the token with methods to work with the Aura API

The methods are organised by capabilities in Aura e.g everything for working with instances are found under client.Instances

Currently there are

.Auth.Get
.Instances.Create
.Instances.Delete
.Instances.List
.Instances.Read
.Tenants.List
.Tenants.Read

You will need the following to instantiate a client
 - A Client ID and Client Secret to get a token.  These are obtained using the Neo4j Aura Console


The Aura API Client is designed to work with v1, its beta iterations and with a timeout of 120 seconds.  These are defined as constants in auraClient.go.  For reference, the values for these are

const (
	BaseURL    = "https://api.neo4j.io/"
	ApiVersion = "v1"
	ApiTimeout = "120"
)

There is no support at the moment to make use of any other version of the Aura API including the beta releases.  That may change as time goes on.   
	
### Hellow world

This lists the tenants in an Aura Organisation. Output is exactly what Tenants.List returns. 

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/LackOfMorals/aura-api-client/auraAPIClient"
)

func main() {

	ctx := context.Background()

	// Set the values of ClientID and ClientSecret to match your own

	ClientID := ""
	ClientSecret := ""
	
	// Instantiate an auraAPIClient, supplying ClientID and ClientSecret
	client := auraAPIClient.NewAuraAPIActionsService(ClientID, ClientSecret)

	// Obtain a token to use with the Aura API
	auraToken, err := client.Auth.GetAuthToken(ctx)
	if err != nil {
		log.Println("Error getting token: ", err)
		os.Exit(1)
	}

	// Get the list of tenants in the Aura Organisation
	response, err := client.Tenants.List(ctx, auraToken)
	if err != nil {
		log.Println("Error getting tenants: ", err)
		os.Exit(1)
	}

	log.Printf("Tenants details: %s", response)

}
```
