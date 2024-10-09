FROM golang:1.23-alpine AS builder
ARG VERSION

WORKDIR /builder

RUN apk add --no-cache git

COPY go.* ./ika/
RUN go -C ./ika mod download

COPY ./example/go.* ./example/
RUN echo "replace github.com/alx99/ika => ../ika">>./example/go.mod && go -C ./example mod download

COPY . ./ika/
COPY ./example ./example
RUN echo "replace github.com/alx99/ika => ../ika">>./example/go.mod

RUN --mount=type=cache,target=/root/.cache/go-build \
    cd example && go build -o /bin/ika -ldflags="-X main.version=${VERSION}" ./main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/ika /ika

CMD ["/ika"]
