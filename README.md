# OpenShift Groups Injector

A lightweight HTTP proxy sidecar that injects OpenShift user group information into HTTP headers for applications using oauth-proxy authentication.

## Problem Statement

OpenShift's `oauth-proxy` v4.4 (distributed with OpenShift 4.16) successfully authenticates users and retrieves their group memberships from OpenShift, but **does not inject group information into HTTP headers** when using the `--set-xauthrequest` flag. This limitation prevents downstream applications like Grafana from implementing group-based role mapping and access control.

## Solution

This groups-injector acts as a transparent proxy between oauth-proxy and your application:

1. Receives requests from oauth-proxy with `X-Forwarded-Access-Token` header
2. Uses the access token to query OpenShift API for user's groups
3. Injects `X-Forwarded-Groups` header containing comma-separated group list
4. Forwards the enriched request to the downstream application

## Architecture

```
Browser → OpenShift Route → oauth-proxy:8443 → groups-injector:8080 → Application:3000
```

The groups-injector runs as a sidecar container in the same pod as your application.

## Features

- Lightweight: Single binary, minimal dependencies
- Fast: Simple reverse proxy with minimal overhead
- Configurable: Environment variables for all settings
- Secure: Uses OpenShift service account for API authentication
- Transparent: No changes required to your application

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | Address to listen on |
| `UPSTREAM_URL` | `http://localhost:3000` | Upstream application URL |
| `OPENSHIFT_API_URL` | `https://kubernetes.default.svc` | OpenShift API endpoint |

## Building

### Prerequisites
- Go 1.21+
- Podman or Docker

### Build the container

```bash
# Clone the repository
git clone https://github.com/yourusername/groups-injector.git
cd groups-injector

# Build with podman
podman build -t groups-injector:latest .

# Or build with docker
docker build -t groups-injector:latest .
```

### For disconnected/air-gapped environments

```bash
# Build locally
podman build -t groups-injector:latest .

# Tag for your registry
podman tag groups-injector:latest your-registry.example.com/groups-injector:latest

# Push to your registry
podman push your-registry.example.com/groups-injector:latest
```

## Deployment

### Example: Grafana with OAuth-Proxy

Deploy as a sidecar container in your application pod:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
spec:
  template:
    spec:
      serviceAccount: oauth-proxy-sa
      containers:
        # OAuth Proxy (entry point)
        - name: oauth-proxy
          image: openshift/oauth-proxy:v4.4
          args:
            - '--upstream=http://localhost:8080'  # Point to groups-injector
            - '--pass-access-token=true'
            # ... other oauth-proxy args
          ports:
            - containerPort: 8443
        
        # Groups Injector (middleware)
        - name: groups-injector
          image: your-registry.example.com/groups-injector:latest
          env:
            - name: UPSTREAM_URL
              value: "http://localhost:3000"
            - name: LISTEN_ADDR
              value: ":8080"
          ports:
            - containerPort: 8080
        
        # Your Application (e.g., Grafana)
        - name: grafana
          image: grafana/grafana:latest
          ports:
            - containerPort: 3000
```

### Required RBAC

The service account needs permission to read user information:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: grafana-auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: oauth-proxy-sa
  namespace: your-namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: grafana-oauth-proxy-cluster-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-reader
subjects:
- kind: ServiceAccount
  name: oauth-proxy-sa
  namespace: your-namespace
```

## Usage with Grafana

Configure Grafana to use the injected groups header for role mapping:

```ini
[auth.proxy]
enabled = true
header_name = X-Forwarded-User
header_property = username
auto_sign_up = true
headers = Groups:X-Forwarded-Groups Email:X-Forwarded-Email
enable_role_attribute_mapping = true
role_attribute_path = contains(Groups[*], 'ocp-cluster-admins') && 'Admin' || 'Editor'
```

## Logging

The groups-injector logs each request with the injected groups:

```
2025/12/29 14:40:24 Injected groups: ocp-cluster-admins,system:authenticated,system:authenticated:oauth
```

## Security Considerations

- The groups-injector runs with the pod's service account credentials
- It uses the OpenShift API within the cluster (no external access required)
- TLS verification is skipped for in-cluster API calls (standard for service-to-service communication)
- Only the `X-Forwarded-Access-Token` header is processed; all other headers pass through unchanged

## Troubleshooting

### No groups appearing in application

1. Verify oauth-proxy is passing the access token:
   ```bash
   oc logs <pod> -c oauth-proxy | grep "X-Forwarded-Access-Token"
   ```

2. Check groups-injector logs:
   ```bash
   oc logs <pod> -c groups-injector
   ```

3. Verify service account has proper RBAC:
   ```bash
   oc auth can-i get users --as system:serviceaccount:your-namespace:oauth-proxy-sa
   ```

### API errors

If you see API errors in the logs, verify:
- Service account exists
- ClusterRoleBindings are applied
- OpenShift API is accessible from the pod

## Use Cases

- **Grafana**: Group-based role mapping and team synchronization
- **Custom applications**: Any app needing OpenShift group information for authorization
- **Multi-tenant systems**: Namespace-based access control using OpenShift groups

## Limitations

- Adds minimal latency (one API call per request with `X-Forwarded-Access-Token`)
- Requires OpenShift 4.x (uses OpenShift-specific user API)
- Designed for oauth-proxy v4.4; newer versions may have native group support

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - See LICENSE file for details

## Acknowledgments

Created to solve group-based access control limitations in OpenShift 4.16's oauth-proxy v4.4 when integrating with Grafana and other applications requiring group membership information.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.
