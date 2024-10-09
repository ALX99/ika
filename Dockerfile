FROM golang:1.23-alpine AS builder

COPY go.* ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/ika ./cmd/ika

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/ika /ika

CMD ["/ika"]
