# OpenShift Groups-Injector Proxy

A lightweight Go reverse proxy designed to run behind an OAuth proxy (like `oauth2-proxy`). It extracts the user's access token, queries the OpenShift User API, and injects the user's group memberships into an HTTP header as a JSON document.

## Feature: JSON Group Injection
Unlike traditional implementations that use comma-separated strings, this version injects group data as a valid JSON array.
* **Header Name:** `X-Forwarded-Groups`
* **Format:** `["group-a", "group-b", "group-c"]`

---

## Unit Testing

The repository includes a comprehensive test suite (`main_test.go`) that validates the injection logic without requiring a live OpenShift cluster or an active upstream service.



### How the Test Works
The test suite utilizes `net/http/httptest` to create a fully isolated lifecycle:
1. **Mock OpenShift API**: Simulates the OpenShift User API (`/apis/user.openshift.io/v1/users/~`) and returns a mock user object with specific groups.
2. **Mock Upstream**: A backend server that receives the proxied request and captures the headers for verification.
3. **Internal Client**: Executes a request through the proxy logic and asserts that the `X-Forwarded-Groups` header reaching the upstream is a correctly formatted JSON string.

### Prerequisites
* Go 1.18+
* `net/http` and `net/http/httptest` (standard library)

### Running the Test
To execute the tests and verify the JSON transformation logic, run:

```bash
go test -v ./...
