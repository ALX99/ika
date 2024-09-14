VERSION := 0.0.1

.PHONY: build
build:
	go build -o ./bin/ika ./cmd/ika

.PHONY: image
image:
	docker build -t ika:$(VERSION) .
	docker build -t ika-builder:$(VERSION) -f Dockerfile.builder .
	docker tag ika:$(VERSION) ika:latest
	docker tag ika-builder:$(VERSION) ika-builder:latest

.PHONY: up
up:
	docker compose up --build

.PHONY: up-reload
up-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make up

.PHONY: upd
upd:
	docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: run
run: build
	./bin/ika -config ./tests/ika.yaml -log-format text

.PHONY: test
test:
	go test -v ./...

.PHONY: e2e
e2e: upd
	k6 run ./tests/tests.js
