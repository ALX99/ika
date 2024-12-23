VERSION := 0.0.1

.PHONY: image
image:
	docker build -t ika:$(VERSION) .
	docker tag ika:$(VERSION) ika:latest

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
run:
	CGO_ENABLED=1 go build -tags full -race -o ./bin/ika ./cmd/ika
	./bin/ika -config ./tests/ika.yaml

.PHONY: run-reload
run-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make run

.PHONY: test
test:
	go test ./...

.PHONY: update-test
update-test:
	find . -name '*.snap' | xargs rm -f
	UPDATE_SNAPS=true make test

.PHONY: deps-up
deps-up:
	docker compose up -d httpbun-local

.PHONY: e2e
e2e: deps-up
	k6 run -e HTTPBUN_HOST=http://localhost:8080 ./tests/tests.js

.PHONY: e2e-compose
e2e-compose: upd
	k6 run -e HTTPBUN_HOST=http://httpbun-local ./tests/tests.js

.PHONY: vet
vet:
	cue vet -c ./config/ ./tests/ika.cue

.PHONY: fmt
fmt:
	cue fmt ./...

cfg: cfg-local

cfg-%: vet fmt
	cue export -t env=$* tests/ika.cue --out yaml > tests/ika.yaml

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
