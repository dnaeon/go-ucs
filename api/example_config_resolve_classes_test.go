package api_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/dnaeon/go-ucs/api"
	"github.com/dnaeon/go-ucs/mo"
)

func Example_configResolveClasses() {
	// The following example shows how to retrieve managed objects from different classes.
	// In the example below we will retrieve the managed objects of `computeBlade` and `computeRackUnit` classes.

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

	req := api.ConfigResolveClassesRequest{
		Cookie:         client.Cookie,
		InHierarchical: "false",
		InIds: []api.Id{
			api.NewId("computeBlade"),
			api.NewId("computeRackUnit"),
		},
	}

	// ComputeItem is a container for all physical compute items.
	var out mo.ComputeItem

	log.Println("Retrieving managed objects with classes `computeBlade` and `computeRackUnit`")
	if err := client.ConfigResolveClasses(ctx, req, &out); err != nil {
		log.Fatalf("Unable to retrieve `computeBlade` and `computeRackUnit` managed object: %s", err)
	}

	log.Printf("Retrieved %d compute blades\n", len(out.Blades))
	log.Printf("Retrieved %d compute rack units\n", len(out.RackUnits))

	for _, blade := range out.Blades {
		log.Printf("%s:\n", blade.Dn)
		log.Printf("\tNumber of CPUs: %d\n", blade.NumOfCpus)
		log.Printf("\tTotal Memory: %d\n", blade.TotalMemory)
		log.Printf("\tModel: %s\n", blade.Model)
		log.Printf("\tVendor: %s\n", blade.Vendor)
	}

	for _, blade := range out.RackUnits {
		log.Printf("%s:\n", blade.Dn)
		log.Printf("\tNumber of CPUs: %d\n", blade.NumOfCpus)
		log.Printf("\tTotal Memory: %d\n", blade.TotalMemory)
		log.Printf("\tModel: %s\n", blade.Model)
		log.Printf("\tVendor: %s\n", blade.Vendor)
	}
}
