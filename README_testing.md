Groups-Injector Proxy - Unit Testing
This repository contains a Go-based reverse proxy that intercepts requests, queries the OpenShift API for user group membership based on an access token, and injects those groups into the X-Forwarded-Groups header as a JSON document.

Test Overview
The unit test in main_test.go verifies the core logic of the group extraction and header injection without requiring a live OpenShift cluster or a real upstream service.

How it Works
The test utilizes the net/http/httptest package to create two separate mock servers:

Mock OpenShift API: Simulates the /apis/user.openshift.io/v1/users/~ endpoint, returning a predefined OpenShiftUser JSON object.

Mock Upstream Server: Acts as the final destination for the proxy. It captures the headers sent to it so the test can assert that the JSON formatting is correct.

Prerequisites
Go: version 1.18 or higher.

Environment: Standard Go testing environment (no external dependencies required).

Running the Tests
To run the unit tests with verbose output, execute the following command in your terminal:

Bash

go test -v ./...
Expected Output
If the injection logic is working correctly, you should see:

Plaintext

=== RUN   TestGroupInjection
--- PASS: TestGroupInjection (0.00s)
PASS
ok  	your-repo-name	0.005s
Test Coverage
The TestGroupInjection function specifically validates:

Token Forwarding: That the X-Forwarded-Access-Token is properly passed to the (mock) OpenShift API.

JSON Transformation: That the []string slice of groups is correctly marshaled into a JSON array (e.g., ["admin", "developers"]).

Header Injection: That the resulting JSON string is correctly set in the X-Forwarded-Groups header before the request reaches the upstream service.
