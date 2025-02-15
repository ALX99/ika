# Fail2ban Plugin

The Fail2ban plugin provides protection against brute force attacks by temporarily banning IPs or identifiers that make too many failed authentication attempts.

## Features

- Configurable retry limits and time windows
- Automatic IP banning after failed attempts
- Custom identifier header support
- Automatic cleanup of expired bans
- Support for reverse proxy headers

## Configuration

| Option        | Type       | Description                                              | Required | Default      |
| ------------- | ---------- | -------------------------------------------------------- | -------- | ------------ |
| `maxRetries`  | `integer`  | Number of failed attempts before banning (must be > 0)   | Yes      | -            |
| `window`      | `duration` | Time window to track failed attempts (e.g., "10m", "1h") | Yes      | -            |
| `banDuration` | `duration` | How long to ban for after exceeding maxRetries           | No       | `window * 2` |
| `idHeader`    | `string`   | Header to use for client identification                  | No       | -            |

::: tip
The `banDuration` defaults to twice the `window` duration if not specified.
:::

::: warning Note
Failed attempts are tracked per IP address (or header value if `idHeader` is set).
:::

### Example

```yaml
middlewares:
  - name: fail2ban
    config:
      maxRetries: 5
      window: 10m
      banDuration: 20m
      idHeader: X-Real-IP
```

## Best Practices

1. Set appropriate retry limits based on your application's security requirements
2. Use `idHeader` when behind a reverse proxy to get the real client IP
3. Configure longer ban durations for stricter security
4. Consider using with the Basic Auth plugin for comprehensive authentication protection

## Common Headers

When using the plugin behind a reverse proxy, common `idHeader` values include:

- `X-Real-IP`: Standard proxy header
- `X-Forwarded-For`: Standard proxy header (first IP)
- `CF-Connecting-IP`: Cloudflare-specific header
- `X-Client-IP`: Alternative proxy header
