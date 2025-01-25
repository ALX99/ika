# Ika

<p align="center">
  <img src="https://github.com/ALX99/ika/blob/main/logo.png" />
</p>

## What is Ika?

Ika is a simple, modular, programmable API Gateway written in Go. It is designed to serve as a base for building a custom API gateway and have **ZERO** *gotchas*.
In fact, it is so simple, that by default, it is closer to a reverse proxy than a full-fledged API Gateway. It even boasts an impressive external dependency count of **2**;
One which is a yaml parser used to read the configuration file, and one other to provide colored text output for [slog](https://pkg.go.dev/log/slog).

### Why Ika?

Ika is designed for people that value the following:

- **Zero gotchas**: The original path (and path parameters) are always perfectly preserved and no extra headers will be added to the request/response. You will find no surprises when using Ika.
- **Simple**: Ika is designed to be simple and easy to understand. It is not a full-fledged API Gateway, but more of a stable base to extend upon.
- **Programmable**: Ika is designed to be programmable. You can easily write your own Go middleware to handle your business logic and compile it into Ika to extend its functionality.
- **Future-proof**: Because Ika basically has no external dependencies, it is very future-proof. As long as Go is around and working, Ika will be too.

### Features

- **Namespace support**: Ika supports configuring multiple namespaces, each with its own isolated configuration which does not interfere with other namespaces.
- **Path matching**: Ika can match paths and capture parameters or wildcards. It supports all the exact same patterns as [`http.ServeMux`](https://pkg.go.dev/net/http#hdr-Patterns).
- **Path rewriting**: Ika is able to rewrite the path of a request before it is sent to the backend.
- **Virtual hosting**: Need to handle traffic for multiple domains? No problem, Ika supports virtual hosting.
- **Middleware support**: Middleware can be applied on a namespace or per-path level. Users of Ika can write their own middleware and compile it into Ika.

### What Ika is not

- Ika is not, and never will be a full-fledged API Gateway like [Kong](https://konghq.com/products/kong-gateway), [Tyk](https://tyk.io) or [KrakenD](https://www.krakend.io).
- Mature. It is a new project, not yet been battle-tested in production, and has yet to see its first 1.0 release. Why don't you help us get there ;)

### Performance

As of now, Ika has not been benchmarked, and in fact there is little reason to do so.
Because Ika is so simple, the performance is expected to be that identical of [`http.Server`](https://pkg.go.dev/net/http#Server) and [`http.ServeMux`](https://pkg.go.dev/net/http#ServeMux).

## Getting started

### Installation

To install Ika, you can use `go get`:

```bash
go install github.com/alx99/ika@latest
```

### Configuration

Ika is configured using a YAML file. The most basic configuration file might look something like this:

```yaml
ika:
  gracefulShutdownTimeout: 30s
  logger:
    level: debug
    format: text

server:
  addr: :8888  # The server will listen on port 8080 for incoming traffic

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

For a full configuration reference please see [ika.example.yaml](./ika.example.yaml) and the [JSON schema](./config/schema.json).
