package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type OpenShiftUser struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Groups []string `json:"groups"`
}

func main() {
	// Configuration
	upstreamURL := getEnv("UPSTREAM_URL", "http://localhost:3000")
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	openshiftAPIURL := getEnv("OPENSHIFT_API_URL", "https://kubernetes.default.svc")
	
	log.Printf("Starting groups-injector on %s", listenAddr)
	log.Printf("Upstream: %s", upstreamURL)
	log.Printf("OpenShift API: %s", openshiftAPIURL)

	// Parse upstream URL
	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		log.Fatalf("Failed to parse upstream URL: %v", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	
	// Create HTTP client for OpenShift API (skip TLS verification for in-cluster)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Handler that injects groups
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the access token from oauth-proxy
		accessToken := r.Header.Get("X-Forwarded-Access-Token")
		
		if accessToken != "" {
			// Query OpenShift API for user info
			groups, err := getUserGroups(client, openshiftAPIURL, accessToken)
			if err != nil {
				log.Printf("Error fetching user groups: %v", err)
			} else if len(groups) > 0 {
				// Inject groups as comma-separated header
				groupsHeader := strings.Join(groups, ",")
				r.Header.Set("X-Forwarded-Groups", groupsHeader)
				log.Printf("Injected groups: %s", groupsHeader)
			}
		}

		// Forward to upstream
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func getUserGroups(client *http.Client, apiURL, token string) ([]string, error) {
	// Query the OpenShift API for user info
	req, err := http.NewRequest("GET", apiURL+"/apis/user.openshift.io/v1/users/~", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var user OpenShiftUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	
	return user.Groups, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
