# Access Log Plugin

The Access Log plugin provides detailed HTTP access logging with configurable output formats and header tracking.

## Features

- Structured logging using slog
- Request and response metrics
- Configurable header logging
- Response timing information
- Pattern matching details

## Configuration

```yaml
hooks:
  - name: access-log
    config:
      # Optional: List of headers to include in logs
      # If not specified, defaults to ["X-Request-ID"]
      headers:
        - User-Agent
        - Referer
        - X-Request-ID
```

### Configuration Options

| Option    | Type       | Description                        | Default            |
| --------- | ---------- | ---------------------------------- | ------------------ |
| `headers` | `string[]` | List of headers to include in logs | `["X-Request-ID"]` |

## Log Output

The plugin logs the following information for each request:

```json
{
  "request": {
    "method": "GET",
    "path": "/api/users",
    "headers": {
      "User-Agent": "curl/7.88.1"
    }
  },
  "response": {
    "duration": 45,
    "status": 200,
    "bytesWritten": 1234
  },
  "ika": {
    "pattern": "/api/users"
  }
}
```

### Log Fields

#### Request

- `method`: HTTP method used
- `path`: Request path
- `headers`: Selected request headers (if configured)

#### Response

- `duration`: Request processing time in milliseconds
- `status`: HTTP status code
- `bytesWritten`: Number of bytes written in response

#### Ika

- `pattern`: The Ika route pattern that matched this request

## Best Practices

1. **Header Selection**: Only log headers that provide value for debugging or monitoring
2. **Performance**: Be mindful of logging volume in high-traffic environments
3. **Security**: Avoid logging sensitive headers like Authorization tokens
4. **Storage**: Consider log rotation and storage requirements
