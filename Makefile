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
	CGO_ENABLED=1 go build -race -o ./bin/ika ./cmd/ika
	./bin/ika -config ./tests/ika.yaml

.PHONY: run-reload
run-reload:
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make run

.PHONY: test
test:
	go test ./...

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

cfg-%: vet fmt
	cue export -t env=$* tests/ika.cue --out yaml > tests/ika.yaml
