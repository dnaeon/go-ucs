package api_test

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"log"
	"net/http"

	"github.com/dnaeon/go-ucs/api"
	"github.com/dnaeon/go-ucs/mo"
)

func Example_configResolveClass() {
	// The following example shows how to retrieve all compute blades.

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

	// The type into which we unmarshal the result data
	type blades struct {
		XMLName xml.Name
		Blades  []mo.ComputeBlade `xml:"computeBlade"`
	}

	req := api.ConfigResolveClassRequest{
		Cookie:         client.Cookie,
		ClassId:        "computeBlade",
		InHierarchical: "false",
	}

	var out blades

	log.Println("Retrieving managed objects with class `computeBlade`")
	if err := client.ConfigResolveClass(ctx, req, &out); err != nil {
		log.Fatalf("Unable to retrieve `computeBlade` managed object: %s", err)
	}

	log.Printf("Retrieved %d compute blades\n", len(out.Blades))
	for _, blade := range out.Blades {
		log.Printf("%s:\n", blade.Dn)
		log.Printf("\tNumber of CPUs: %d\n", blade.NumOfCpus)
		log.Printf("\tTotal Memory: %d\n", blade.TotalMemory)
		log.Printf("\tModel: %s\n", blade.Model)
		log.Printf("\tVendor: %s\n", blade.Vendor)
	}
}
