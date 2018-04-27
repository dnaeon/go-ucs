package api_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/dnaeon/go-ucs/api"
)

func Example_aaaRefresh() {
	// The following example shows how to refresh an existing session.

	// Skip SSL certificate verification of remote endpoint.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	// Create a new Cisco UCS API client
	config := api.Config{
		Endpoint:   "https://ucs01.example.org/",
		Username:   "admin",
		Password:   "password",
		HttpClient: httpClient,
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Unable to create API client: %s", err)
	}

	ctx := context.Background()

	// Authenticate to the remote API endpoint and obtain authentication cookie
	log.Printf("Logging in to %s\n", config.Endpoint)
	if _, err := client.AaaLogin(ctx); err != nil {
		log.Fatalf("Unable to authenticate: %s", err)
	}
	defer client.AaaLogout(ctx)

	log.Printf("Got authentication cookie: %s\n", client.Cookie)

	log.Println("Refreshing session")
	if _, err := client.AaaRefresh(ctx); err != nil {
		log.Fatalf("Unable to refresh session: %s\n", err)
	}

	log.Printf("New authentication cookie is: %s\n", client.Cookie)
}
