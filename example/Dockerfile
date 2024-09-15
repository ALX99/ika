FROM golang:1.23-alpine AS builder
ARG VERSION

RUN apk add --no-cache git
COPY go.* ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/ika -ldflags="-X main.version=${VERSION}" ./main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/ika /ika
COPY ika.yaml /ika.yaml

CMD ["/ika"]
