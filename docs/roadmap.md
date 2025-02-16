# Roadmap

This page outlines the planned features and improvements for Ika Gateway. The roadmap items are organized in no particular order, with a focus on maintaining Ika's core principles of minimalism, flexibility, and performance.

## Core

::: info Configuration

- Configuration validation <Badge type="tip">Complete</Badge>
- Configuration variable support <Badge type="info">Idea</Badge>
- Remote configuration reference <Badge type="info">Idea</Badge>
- Live configuration reloading <Badge type="danger">Planned</Badge>
- Configuration templating <Badge type="info">Idea</Badge>
- Error response customization <Badge type="warning">Ongoing</Badge>
- TLS support <Badge type="danger">Planned</Badge>
- H2C support <Badge type="danger">Planned</Badge>
- Global plugins <Badge type="danger">Planned</Badge>
- Configuration policy support <Badge type="info">Idea</Badge>

:::

## Plugins

::: info Plugins

- Rate limiter <Badge type="danger">Planned</Badge>
- Request validator <Badge type="danger">Planned</Badge>
  - JSON Schema <Badge type="info">Idea</Badge>
  - Dynamic validation <Badge type="info">Idea</Badge>
- Circuit breaker <Badge type="danger">Planned</Badge>
- Request robuster
  - Retry mechanism <Badge type="danger">Planned</Badge>
  - Timeout handling <Badge type="danger">Planned</Badge>
  - Request/Response body buffer control <Badge type="danger">Planned</Badge>
  - Bulkhead pattern <Badge type="danger">Planned</Badge>
- Cache system <Badge type="danger">Planned</Badge>
  - Auto cache function <Badge type="info">Idea</Badge>
- Request introspection (debug) <Badge type="danger">Planned</Badge>
- JWT Auth <Badge type="danger">Planned</Badge>
- Security <Badge type="info">Idea</Badge>
    - CORS <Badge type="info">Idea</Badge>
    - CSP <Badge type="info">Idea</Badge>
- Static file server <Badge type="danger">Planned</Badge>
- Whitelist (IP/CIDR) <Badge type="info">Idea</Badge>
- Query allowlist <Badge type="info">Idea</Badge>
- Load balancer <Badge type="danger">Planned</Badge>

:::

## Plugin System

::: info Plugin System

- Plugin dependency management <Badge type="info">Idea</Badge>
- Plugin communication <Badge type="info">Idea</Badge>

:::

## Documentation

::: info Documentation

- Expand all documentation <Badge type="warning">Ongoing</Badge>
- Expanded plugin development guide <Badge type="danger">Planned</Badge>
- Deploying to production guide <Badge type="danger">Planned</Badge>
- More real-world examples <Badge type="danger">Planned</Badge>
- API reference documentation <Badge type="danger">Planned</Badge>

:::

---

::: tip Contributing
This roadmap evolves based on community feedback and changing requirements. Feel free to suggest new features or improvements through our [GitHub repository](https://github.com/alx99/ika)!
:::

::: details Status Key

- <Badge type="tip">Complete</Badge> - Feature is implemented and available
- <Badge type="warning">In Progress</Badge> - Currently being worked on
- <Badge type="danger">Planned</Badge> - Planned but not yet started
- <Badge type="info">Idea</Badge> - Feature is in the discussion stage

:::
