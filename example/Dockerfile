FROM golang:1.24-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /bin/example ./main.go

FROM scratch
COPY --from=build \
  /etc/ssl/certs/ca-certificates.crt \
  /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/example /example
COPY ika.yaml /ika.yaml

CMD ["/example"]
