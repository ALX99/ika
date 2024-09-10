FROM ika-builder:latest AS builder


COPY . /user

RUN --mount=type=cache,target=/root/.cache/go-build \
  /ika/builder --package example --package-directory /user --output-directory /bin

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/ika /ika
COPY ika.yaml /ika.yaml

CMD ["/ika"]
