# Basic Auth Plugin

The Basic Auth plugin provides HTTP Basic Authentication support for both incoming and outgoing requests. It can validate incoming requests against configured credentials and add authentication to outgoing requests.

## Features

- Incoming request authentication
- Outgoing request authentication
- Environment variable support for credentials
- Configurable for both client and server roles

## Configuration

| Option              | Type     | Description                                                            | Required | Default  |
| ------------------- | -------- | ---------------------------------------------------------------------- | -------- | -------- |
| `incoming.type`     | `string` | Credential lookup type for incoming requests. Options: `static`, `env` | No       | `static` |
| `incoming.username` | `string` | Username or environment variable name for incoming requests            | Yes\*    | -        |
| `incoming.password` | `string` | Password or environment variable name for incoming requests            | Yes\*    | -        |
| `outgoing.type`     | `string` | Credential lookup type for outgoing requests. Options: `static`, `env` | No       | `static` |
| `outgoing.username` | `string` | Username or environment variable name for outgoing requests            | Yes\*    | -        |
| `outgoing.password` | `string` | Password or environment variable name for outgoing requests            | Yes\*    | -        |

::: warning Note
At least one of `incoming` or `outgoing` must be configured.  
\*Required when the respective section (`incoming` or `outgoing`) is used.
:::

### Example

```yaml
middlewares:
  - name: basic-auth
    config:
      incoming:
        type: static
        username: admin
        password: secret
      outgoing:
        type: env
        username: AUTH_USER
        password: AUTH_PASS
```

## Best Practices

1. Use environment variables for sensitive credentials
2. Ensure strong passwords in production
3. Use HTTPS to protect credential transmission
4. Consider using more secure authentication methods for sensitive operations
