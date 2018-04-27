package api_test

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"log"
	"net/http"
	"strings"

	"github.com/dnaeon/go-ucs/api"
	"github.com/dnaeon/go-ucs/mo"
)

func Example_configResolveDns() {
	// The following example shows how to retrieve a list of managed objects by specifying their DNs.

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

	// A type to contain the instances of the retrieved DNs.
	type outConfigs struct {
		XMLName xml.Name
		Sys     mo.TopSystem          `xml:"topSystem"`
		Version mo.VersionApplication `xml:"versionApplication"`
	}

	var out outConfigs

	// Retrieve the list of DNs
	log.Println("Retrieving managed objects using configResolveDns query method")
	req := api.ConfigResolveDnsRequest{
		Cookie:         client.Cookie,
		InHierarchical: "false",
		InDns: []api.Dn{
			api.NewDn("sys"),
			api.NewDn("sys/version/application"),
			api.NewDn("no/such/dn"),
		},
	}

	resp, err := client.ConfigResolveDns(ctx, req, &out)
	if err != nil {
		log.Fatalf("Unable to retrieve DNs: %s", err)
	}

	log.Printf("%s is at version %s\n", out.Sys.Name, out.Version.Version)

	unresolved := make([]string, 0)
	for _, dn := range resp.OutUnresolved {
		unresolved = append(unresolved, dn.Value)
	}
	log.Printf("Unresolved DNs: %s", strings.Join(unresolved, ", "))
}
