# Request ID Plugin

The Request ID plugin adds unique identifiers to HTTP requests, helping with request tracing and debugging. It supports multiple ID generation algorithms and flexible header management.

## Features

- Multiple ID generation algorithms (UUIDv4, UUIDv7, KSUID, XID)
- Configurable header name
- Flexible handling of existing request IDs
- Optional response header inclusion
- Cryptographically secure random number generation

## Configuration

| Option     | Type      | Description                                                                 | Required | Default        |
| ---------- | --------- | --------------------------------------------------------------------------- | -------- | -------------- |
| `header`   | `string`  | The HTTP header name used for the request ID                                | No       | `X-Request-ID` |
| `variant`  | `string`  | ID generation algorithm to use. Options: `UUIDv4`, `UUIDv7`, `KSUID`, `XID` | No       | `XID`          |
| `override` | `boolean` | Replace existing header value with new ID                                   | No       | `true`         |
| `append`   | `boolean` | Add new ID while preserving existing value                                  | No       | `false`        |
| `expose`   | `boolean` | Copy the request ID to response headers                                     | No       | `true`         |

### Example

```yaml
plugins:
  request-id:
    header: X-Request-ID
    variant: XID
    override: true
    append: false
    expose: true
```
