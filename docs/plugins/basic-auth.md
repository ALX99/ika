# Basic Auth Plugin

The Basic Auth plugin provides HTTP Basic Authentication support for both incoming and outgoing requests. It can validate incoming requests against configured credentials and add authentication to outgoing requests.

## Features

- Multiple incoming credentials support
- Outgoing request authentication
- Environment variable support for credentials
- Configurable for both client and server roles
- Named credentials for better organization

## Configuration

| Option     | Type     | Description                         | Required | Default |
| ---------- | -------- | ----------------------------------- | -------- | ------- |
| `incoming` | `object` | Configuration for incoming requests | Yes\*    | -       |
| `outgoing` | `object` | Configuration for outgoing requests | Yes\*    | -       |

### `incoming` Configuration

The `incoming` section supports multiple named credentials. Each credential can be configured independently.

| Option                            | Type      | Description                                      | Required | Default  |
| --------------------------------- | --------- | ------------------------------------------------ | -------- | -------- |
| `incoming.credentials[].name`     | `string`  | Unique identifier for the credential             | Yes      | -        |
| `incoming.credentials[].type`     | `string`  | Credential lookup type. Options: `static`, `env` | No       | `static` |
| `incoming.credentials[].username` | `string`  | Username or environment variable name            | Yes      | -        |
| `incoming.credentials[].password` | `string`  | Password or environment variable name            | Yes      | -        |
| `incoming.strip`                  | `boolean` | Remove credentials after successful authentication| No       | `false`  |

### `outgoing` Configuration

| Option              | Type     | Description                                                            | Required | Default  |
| ------------------- | -------- | ---------------------------------------------------------------------- | -------- | -------- |
| `outgoing.type`     | `string` | Credential lookup type for outgoing requests. Options: `static`, `env` | No       | `static` |
| `outgoing.username` | `string` | Username or environment variable name for outgoing requests            | Yes      | -        |
| `outgoing.password` | `string` | Password or environment variable name for outgoing requests            | Yes      | -        |

::: warning Note
At least one of `incoming` or `outgoing` must be configured.  
:::

### Example

```yaml
middlewares:
  - name: basic-auth
    config:
      incoming:
        strip: false
        credentials:
          - name: admin
            type: env
            username: ADMIN_USERNAME
            password: ADMIN_PASSWORD
          - name: user
            type: static
            username: regularuser
            password: userpass123
      outgoing:
        type: env
        username: AUTH_USER
        password: AUTH_PASS
```

In this example:

- Two incoming credentials are configured:
  1. An admin user with credentials from environment variables
  2. A regular user with static credentials
- Outgoing requests will use credentials from environment variables

## Best Practices

1. Use environment variables for sensitive credentials
2. Ensure strong passwords in production
3. Use HTTPS to protect credential transmission
4. Consider using more secure authentication methods for sensitive operations
5. Use meaningful names for credentials to aid in organization and monitoring
6. Avoid duplicate credential names as they will cause validation errors
7. Enable `strip` when you want to prevent credentials from being forwarded to upstream services
