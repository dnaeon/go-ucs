package api_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/dnaeon/go-ucs/api"
	"github.com/dnaeon/go-ucs/mo"
)

func Example_configResolveDn() {
	// The following example shows how to retrieve a single managed object for a specified DN.

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

	log.Printf("Logging in to %s\n", config.Endpoint)
	if _, err := client.AaaLogin(ctx); err != nil {
		log.Fatalf("Unable to login: %s\n", err)
	}
	defer client.AaaLogout(ctx)

	log.Printf("Got authentication cookie: %s\n", client.Cookie)

	// Retrieve the `sys` DN, which is of type mo.TopSystem
	log.Println("Retrieving `sys` managed object")
	req := api.ConfigResolveDnRequest{
		Cookie:         client.Cookie,
		Dn:             "sys",
		InHierarchical: "false",
	}

	var sys mo.TopSystem
	if err := client.ConfigResolveDn(ctx, req, &sys); err != nil {
		log.Fatalf("Unable to retrieve DN: %s", err)
	}

	log.Printf("Address: %s\n", sys.Address)
	log.Printf("Current time: %s\n", sys.CurrentTime)
	log.Printf("Dn: %s\n", sys.Dn)
	log.Printf("Mode: %s\n", sys.Mode)
	log.Printf("Uptime: %s\n", sys.SystemUptime)
}
