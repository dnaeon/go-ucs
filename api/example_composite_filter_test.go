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

func Example_compositeFilter() {
	// The following example shows how to use a composite AND property filter.

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

	// Create a composite AND property filter, which will find all blades,
	// which have total memory greater than or equal to 2048 MB and are in chassis 3.
	filter := api.FilterAnd{
		Filters: []api.FilterAny{
			api.FilterGe{
				FilterProperty: api.FilterProperty{
					Class:    "computeBlade",
					Property: "totalMemory",
					Value:    "2048",
				},
			},
			api.FilterEq{
				FilterProperty: api.FilterProperty{
					Class:    "computeBlade",
					Property: "chassisId",
					Value:    "3",
				},
			},
		},
	}

	req := api.ConfigResolveClassRequest{
		Cookie:         client.Cookie,
		InHierarchical: "false",
		ClassId:        "computeBlade",
		InFilter:       filter,
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
		log.Printf("\tChassis ID: %d\n", blade.ChassisId)
		log.Printf("\tVendor: %s\n", blade.Vendor)
	}
}
