package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGroupInjection(t *testing.T) {
	// 1. Mock the OpenShift API
	mockOSAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := OpenShiftUser{
			Groups: []string{"admin", "developers"},
		}
		user.Metadata.Name = "test-user"
		json.NewEncoder(w).Encode(user)
	}))
	defer mockOSAPI.Close()

	// 2. Mock the Upstream Server (to verify headers)
	upstreamReceivedJSON := ""
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamReceivedJSON = r.Header.Get("X-Forwarded-Groups")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockUpstream.Close()

	// 3. Setup Proxy Logic
	upstreamURL, _ := url.Parse(mockUpstream.URL)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := http.DefaultClient
		accessToken := r.Header.Get("X-Forwarded-Access-Token")

		if accessToken != "" {
			groups, _ := getUserGroups(client, mockOSAPI.URL, accessToken)
			if len(groups) > 0 {
				jsonGroups, _ := json.Marshal(groups)
				r.Header.Set("X-Forwarded-Groups", string(jsonGroups))
			}
		}

		// Forward to mock upstream
		r.Host = upstreamURL.Host
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simplified proxying for the test
			req, _ := http.NewRequest(r.Method, mockUpstream.URL, nil)
			req.Header = r.Header
			client.Do(req)
		}).ServeHTTP(w, r)
	}))
	defer proxy.Close()

	// 4. Execute Request
	req, _ := http.NewRequest("GET", proxy.URL, nil)
	req.Header.Set("X-Forwarded-Access-Token", "fake-token")
	http.DefaultClient.Do(req)

	// 5. Assertions
	expectedJSON := `["admin","developers"]`
	if upstreamReceivedJSON != expectedJSON {
		t.Errorf("Expected JSON header %s, got %s", expectedJSON, upstreamReceivedJSON)
	}
}
