VERSION := 0.0.1
BIN_DIR := ./bin
CMD_DIR := ./cmd/ika
TESTS_DIR := ./tests
CONFIG_FILE := $(TESTS_DIR)/ika.yaml
TEST_FILE := $(TESTS_DIR)/tests.js
BUILD_TAGS ?=

.PHONY: image
image:
	docker build --build-arg BUILD_TAGS="$(BUILD_TAGS)" -t ika:$(VERSION) .
	docker tag ika:$(VERSION) ika:latest

.PHONY: up
up:
	BUILD_TAGS=$(BUILD_TAGS) docker compose up --build

.PHONY: up-reload
up-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make up

.PHONY: upd
upd:
	BUILD_TAGS=$(BUILD_TAGS) docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: run
run:
	CGO_ENABLED=1 go build -tags "$(BUILD_TAGS)" -race -o $(BIN_DIR)/ika $(CMD_DIR)
	$(BIN_DIR)/ika -config $(CONFIG_FILE)

.PHONY: run-reload
run-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make run

.PHONY: test
test:
	./scripts/test-all $(ARGS)

.PHONY: fmt
fmt:
	cue fmt ./...
	gofmt -l -w .
	pnpm run fmt

.PHONY: tidy
tidy:
	./scripts/tidy

.PHONY: clean
clean:
	find . -name "coverage.out" -type f -delete
