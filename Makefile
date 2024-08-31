VERSION := 0.0.1

.PHONY: build
build:
	go build -o ./bin/ika ./cmd/ika

.PHONY: up
up:
	docker compose up --build

.PHONY: up-reload
up-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make up

.PHONY: upd
upd:
	docker-compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: run
run: build
	./bin/ika -config ./tests/ika.yaml

.PHONY: test
test:
	go test -v ./...

.PHONY: e2e
e2e: down upd
	k6 run ./tests/tests.js
