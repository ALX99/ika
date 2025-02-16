# Ika

<p align="center">
  <img src="./docs/public/logo.png" width="200" style="display: block; margin: auto;" />
</p>

Ika is a minimalist, plugin-driven API Gateway that puts you in control. It starts with a lean core focused purely on routing, allowing you to add only the features you need through plugins.

## Why Ika?

### 🎯 Focused Design

Everything beyond basic routing is a plugin. This means:

- Cleaner, more maintainable codebase
- Faster deployments with smaller binaries
- Better reliability through reduced complexity

### 🔌 True Extensibility

Need a custom feature? Ika's plugin system gives you the same power as core developers:

- Write plugins in pure Go
- Access the full request/response lifecycle
- Combine plugins to create powerful workflows

### ⚡ Performance First

Starting minimal means:

- Lightning-fast startup times
- Lower memory footprint
- Efficient request processing

### 🧪 Quality Assured

Every component is thoroughly tested:

- Rock-solid core functionality
- Reliable plugin ecosystem
- Stable production performance

## Installation

Choose your preferred installation method:

```bash
# From source
go install github.com/alx99/ika/cmd/ika@latest

# Using Docker
docker pull ghcr.io/alx99/ika:latest

# Binary release
VERSION="v0.0.6"
curl -Lo- "https://github.com/alx99/ika/releases/download/$VERSION/ika_Linux_$(uname -m).tar.gz" | tar -xz --wildcards 'ika'
```

### Basic Configuration

Create a minimal `ika.yaml`:

```yaml
ika:
  gracefulShutdownTimeout: 30s
  logger:
    level: debug
    format: text

server:
  addr: :8888

namespaces:
  api:
    reqModifiers:
      - name: basic-modifier
        config:
          host: https://dummyjson.com
    middlewares:
      - name: access-log
    paths:
      /users: {}
```

For complete documentation and examples, visit our [documentation](https://ika.dozy.dev).
