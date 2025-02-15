# Access Log Plugin

The Access Log plugin provides detailed HTTP access logging with configurable output formats and header tracking.

## Features

- Structured logging using slog
- Request and response metrics
- Configurable header logging
- Selective query parameter logging
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
      # Optional: Include remote address in logs
      remoteAddr: true
      # Optional: List of query parameters to include in logs
      # If not specified, no query parameters will be logged
      queryParams:
        - page
        - limit
        - sort
```

### Configuration Options

| Option        | Type       | Description                         | Required | Default            |
| ------------- | ---------- | ----------------------------------- | -------- | ------------------ |
| `headers`     | `string[]` | List of headers to include in logs  | No       | `["X-Request-ID"]` |
| `remoteAddr`  | `boolean`  | Include remote address in logs      | No       | `false`            |
| `queryParams` | `string[]` | List of query parameters to include | No       | `[]`               |

## Log Output

The plugin logs the following information for each request:

```json
{
  "request": {
    "method": "GET",
    "path": "/api/users",
    "headers": {
      "User-Agent": "curl/7.88.1"
    },
    "query": {
      "page": "1",
      "limit": "10"
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
- `query`: Selected query parameters (if configured)

#### Response

- `duration`: Request processing time in milliseconds
- `status`: HTTP status code
- `bytesWritten`: Number of bytes written in response

#### Ika

- `pattern`: The Ika route pattern that matched this request

## Best Practices

1. **Header Selection**: Only log headers that provide value for debugging or monitoring
2. **Query Parameter Selection**: Only log query parameters needed for debugging or analytics
3. **Performance**: Be mindful of logging volume in high-traffic environments
4. **Security**: Avoid logging sensitive headers or query parameters
5. **Storage**: Consider log rotation and storage requirements
