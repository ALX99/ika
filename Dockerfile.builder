FROM golang:1.23-alpine

WORKDIR /ika

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  go mod download && go build -o tmp ./cmd/ika/main.go && rm tmp

ENTRYPOINT ["/ika/builder"]
