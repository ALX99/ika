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

.PHONY: update-test
update-test:
	find . -name '*.snap' | xargs rm -f
	UPDATE_SNAPS=true make test

.PHONY: deps-up
deps-up:
	docker compose up -d httpbun-local

.PHONY: e2e
e2e: deps-up
	k6 run -e HTTPBUN_HOST=http://localhost:8080 $(TEST_FILE)

.PHONY: e2e-compose
e2e-compose: upd
	k6 run -e HTTPBUN_HOST=http://httpbun-local $(TEST_FILE)

.PHONY: vet
vet:
	cue vet -c ./schema/ $(TESTS_DIR)/ika.cue

.PHONY: fmt
fmt:
	cue fmt ./...
	gofmt -l -w .
	pnpm run fmt

cfg: cfg-local

cfg-%: vet fmt
	cue export -t env=$* $(TESTS_DIR)/ika.cue --out yaml > $(CONFIG_FILE)
	cp -f $(CONFIG_FILE) internal/ika/

.PHONY: release-patch
release-patch:
	latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	if [ -z "$$latest_tag" ]; then \
		new_tag="v0.0.1"; \
	else \
		IFS='.' read -r major minor patch <<< "$${latest_tag#v}"; \
		new_patch=$$(($$patch + 1)); \
		new_tag="v$$major.$$minor.$$new_patch"; \
	fi; \
	git tag $$new_tag; \
	echo "Tagged with $$new_tag"

.PHONY: release
release:
	latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	if [ -z "$$latest_tag" ]; then \
		echo "No tags found"; \
	else \
		git push origin $$latest_tag; \
		echo "Pushed tag $$latest_tag"; \
	fi

.PHONY: tidy
tidy:
	./scripts/tidy

.PHONY: clean
clean:
	find . -name "coverage.out" -type f -delete
