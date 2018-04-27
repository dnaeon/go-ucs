package api_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dnaeon/go-ucs/api"
	"github.com/dnaeon/go-ucs/mo"
)

func Example_rateLimiter() {
	// The following example shows how to rate limit requests again the remote
	// Cisco UCS API endpoint using a token bucket rate limiter.
	// https://en.wikipedia.org/wiki/Token_bucket

	// Skip SSL certificate verification of remote endpoint.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	// Create a new Cisco UCS API client.
	// Set maximum allowed requests per second to 1 with a burst size of 1.
	// A request will wait up to 1 minute for a token.
	config := api.Config{
		Endpoint:   "https://ucs01.example.org/",
		Username:   "admin",
		Password:   "password",
		HttpClient: httpClient,
		RateLimit: &api.RateLimit{
			RequestsPerSecond: 1.0,
			Burst:             1,
			Wait:              time.Duration(1 * time.Minute),
		},
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

	// Start a few concurrent requests to the remote API endpoint.
	// Requests will be executed one at a time, because of how the limiter is configured.
	var wg sync.WaitGroup
	for i := 1; i < 10; i++ {
		wg.Add(1)

		// Our worker function will retrieve the `sys` DN from the remote Cisco UCS API.
		// We will start a few of these in separate goroutines.
		worker := func(id int) {
			defer wg.Done()

			// Retrieve the `sys` DN, which is of type mo.TopSystem
			log.Printf("Worker #%d: Retrieving `sys` managed object\n", id)
			req := api.ConfigResolveDnRequest{
				Cookie:         client.Cookie,
				Dn:             "sys",
				InHierarchical: "false",
			}

			var sys mo.TopSystem
			if err := client.ConfigResolveDn(ctx, req, &sys); err != nil {
				log.Printf("Worker #%d: Unable to retrieve DN: %s\n", id, err)
				return
			}

			log.Printf("Worker #%d: successfully retrieved `sys` managed object\n", id)
		}

		go worker(i)
	}

	wg.Wait()
}
