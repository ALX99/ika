# Showcase

This showcase demonstrates Ika's capabilities in managing multiple services in a real-world scenario. You'll see how to:

- Route different services through isolated namespaces
- Apply service-specific plugin configurations
- Handle various traffic patterns and security requirements

## Complete Configuration

<<<@/example/configs/complex.yaml

## Key Features Demonstrated

### 1. Public API (`api.example.com`)

- Multiple API versions (v1, v2) with different backends
- Version-specific request ID formats (XID for v2, UUIDv4 for v1)
- Automatic host rewriting with internal service routing
- Structured logging with request tracking

### 2. Admin Dashboard (`admin.example.com`, `dashboard.example.com`)

- Multiple domain support for the same service
- Environment-based authentication for admin routes
- Public and protected route separation
- Detailed access logging with IP tracking

### 3. Internal Services (`internal.example.com`)

- Dynamic service and version-based routing
- Tenant-aware request tracking
- Structured logging with business context
- Rate limiting and security controls

## Testing the Configuration

Here are some example requests to test different aspects of the configuration:

```bash
# Public API v2 request
curl -H "Host: api.example.com" "http://localhost:8080/v2/users"
# → Routed to api-v2.internal with request tracking

# Public API v1 (legacy) request
curl -H "Host: api.example.com" "http://localhost:8080/v1/users"
# → Routed to legacy-api.internal/api/users with backward compatibility

# Protected admin route
curl -H "Host: admin.example.com" -u "$ADMIN_USER:$ADMIN_PASSWORD" "http://localhost:8080/admin/metrics"
# → Routed to admin-panel.internal/admin/metrics with auth

# Public dashboard route (works on both domains)
curl -H "Host: dashboard.example.com" "http://localhost:8080/public/status"
# → Routed to public-dashboard.internal/public/status

# Internal service with version and tenant context
curl -H "Host: internal.example.com" "http://localhost:8080/billing/v2/invoices?tenant=acme&region=eu-west"
# → Routed to billing-v2.internal/invoices with tenant context
```

## Key Takeaways

This showcase demonstrates several powerful features of Ika:

1. **Namespace Isolation**: Each service group operates independently with its own configuration
2. **Plugin Flexibility**: Mix and match plugins to create the exact functionality needed
3. **Security in Depth**: Authentication, rate limiting, and request tracking working together
4. **Dynamic Routing**: Pattern-based routing with variable capture and rewriting
5. **Observability**: Comprehensive logging with business context preservation
