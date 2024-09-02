# Ika

<p align="center">
  <img src="https://github.com/ALX99/ika/blob/main/logo.png" />
</p>

Ika is a simple, modular, programmable API Gateway written in Go. It is designed to serve as a base for building custom API gateway and have **ZERO** *gotchas*.
In fact, it is so simple, that by default, it is closer to a reverse proxy than a full-fledged API Gateway. It even boasts an impressive external dependency count of **1**.
This single dependency is yaml parser used to read the configuration file.


## Why Ika?

Ika is designed for people that value the following:

- **Zero gotchas**: The original path (and path parameters) are always perfectly preserved, and no extra headers will be added to the request/response. You will find no surprises when using Ika.
- **Simple**: Ika is designed to be simple and easy to understand. It is not a full-fledged API Gateway, but more of a stable base to extend upon.
- **Programmable**: Ika is designed to be programmable. You can easily write your own Go middleware and compile it into Ika to extend its functionality.
- **Future-proof**: Because Ika basically has no external dependencies, it is very future-proof. As long as Go is around and working, Ika will be too.

## Features

- **Namespace support**: Ika supports configuring multiple namespaces, each with its own isolated configuration which does not interfere with other namespaces.
- **Path matching**: Ika can match paths and capture parameters or wildcards. It supports all the exact same patterns as [`http.ServeMux`](https://pkg.go.dev/net/http#hdr-Patterns).
- **Path rewriting**: Ika is able to rewrite the path of a request before it is sent to the backend.
- **Middleware support**: Ika supports middleware that can be applied on a namespace or path level. Users of Ika can write their own middleware and compile it into Ika.

## What Ika is not

- Ika is not a full-fledged API Gateway like [Kong](https://konghq.com/products/kong-gateway), [Tyk](https://tyk.io) or [KrakenD](https://www.krakend.io).
- Mature. It is a new project and has not yet been battle-tested in production, and has yet to see its first 1.0 release.


## Performance

As of now, Ika has not been benchmarked, and in fact there is little reason to do so.
Because Ika is so simple, the performance is expected to be that identical of [`http.Server`](https://pkg.go.dev/net/http#Server), [`http.ServeMux`](https://pkg.go.dev/net/http#ServeMux).
