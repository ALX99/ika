FROM golang:1.23-alpine AS build

RUN apk add --no-cache ca-certificates

COPY go.* ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -tags full -o /bin/ika ./cmd/ika

FROM scratch
COPY --from=build \
  /etc/ssl/certs/ca-certificates.crt \
  /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/ika /ika

CMD ["/ika"]
