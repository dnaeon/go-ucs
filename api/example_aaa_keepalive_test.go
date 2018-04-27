package api_test

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnaeon/go-ucs/api"
)

func Example_aaaKeepAlive() {
	// The following example shows how to keep a session alive.

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

	// Channel on which the shutdown signal is sent
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Send a KeepAlive request every minute
	ticker := time.NewTicker(1 * time.Minute)

L:
	for {
		select {
		case <-quit:
			log.Println("Keyboard interrupt detected, terminating.")
			break L
		case <-ticker.C:
			log.Println("Sending KeepAlive request ...")
			resp, err := client.AaaKeepAlive(ctx)
			if err != nil {
				log.Printf("Unable to keep session alive: %s\n", err)
				break L
			}

			log.Printf("Got response with cookie: %s\n", resp.Cookie)
		}
	}
}
