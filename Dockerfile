FROM golang:1.24-alpine AS build

ARG BUILD_TAGS

RUN apk add --no-cache ca-certificates

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -tags "$BUILD_TAGS" -ldflags '-w -s' -o /bin/ika ./cmd/ika-full/main.go

FROM scratch
COPY --from=build \
  /etc/ssl/certs/ca-certificates.crt \
  /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/ika /usr/bin/ika

ENTRYPOINT ["/usr/bin/ika"]
