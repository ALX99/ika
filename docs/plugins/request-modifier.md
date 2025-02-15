# Request Modifier Plugin

The Request Modifier plugin allows dynamic modification of HTTP requests, including URL path rewriting and host rewriting. It's particularly useful for URL normalization, API versioning, and request routing.

## Features

- Dynamic path rewriting with segment capture
- Host and scheme rewriting
- Wildcard path segment support
- Configurable host header retention
- Pattern-based URL transformation

## Configuration

| Option             | Type      | Description                                                       | Required | Default |
| ------------------ | --------- | ----------------------------------------------------------------- | -------- | ------- |
| `path`             | `string`  | New path pattern that can include segments from the original path | No\*     | -       |
| `host`             | `string`  | Target host URL including scheme (e.g., "https://example.com")    | No\*     | -       |
| `retainHostHeader` | `boolean` | Whether to preserve the original Host header                      | No       | `false` |

::: warning Note
\*At least one of `path` or `host` must be configured.
:::

### Example

```yaml
reqModifiers:
  - name: req-modifier
    config:
      path: /new/{id}/path/{wildcard...}
      host: https://api.example.com
      retainHostHeader: false
```

### Path Rewriting

The path rewriting feature uses a pattern-based syntax:

- `{segment}`: Captures a single path segment
- `{wildcard...}`: Captures all remaining path segments
- `{$}`: Special token for static segments

## Best Practices

1. Use path rewriting for API versioning and normalization
2. Enable `retainHostHeader` when required by upstream services
3. Use wildcards carefully to maintain URL structure
4. Test rewrite patterns with various path combinations
