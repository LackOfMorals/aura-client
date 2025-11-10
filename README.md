# Aura API Client

## Overview
There's a few occasions where I have wanted to use the Aura API  with Go based applications.  Rather than re-write the integration each and everytime, it would be nice to have module that can be re-used.  Hence I have put together the Aura API Client. 

My Go knowledge is still being acquired - it's something of a hobby. I've looped around writing Go, trying it, improving, seeing difference directions to take and have had chats with Claude to fill holes in my knowledge.  It's been a great learning experience.   The point I am making - this has not been authored by a professional developer; there are rough edges. 


## Aura API Go Client
The Aura API Go Client allows for working with Instances, Snaphosts, and GDS Sessions.   The following pattern is followed

<Aura feature>.<operation>

For example, after you have instantiated the Aura API Go Client, to list Aura instances

client.Instances.List


## Requirements
* Go 1.23 or newer
* Client ID and Client Secret for Neo4j AuraDB


## Installation

Obtain the Aura API Go Client for your Go application with

```bash
go get github.com/LackOfMorals/aura-api-client
```

## Usage

This will work for the majority of the time.  
```Go
client, err := aura.NewAuraAPIClient(
    aura.WithCredentials("id", "secret"),
)
```

You can also set various configuration options

```Go
client, err := aura.NewAuraAPIClient(
    aura.WithClientID("my-id"),
    aura.WithClientSecret("my-secret"),
    aura.WithBaseURL("https://custom.api.com/"),
    aura.WithTimeout(60 * time.Second),
    aura.WithVersion("v1"),
)
```

