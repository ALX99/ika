VERSION := 0.0.1

.PHONY: image
image:
	docker build -t ika:$(VERSION) .
	docker tag ika:$(VERSION) ika:latest

.PHONY: docker-config
docker-config:
	export HTTPBUN_HOST="http://httpbun-local" && \
		envsubst < ./tests/ika.tpl.yaml > ./tests/ika.yaml

.PHONY: local-config
local-config:
	export HTTPBUN_HOST="http://localhost:8080" && \
		envsubst < ./tests/ika.tpl.yaml > ./tests/ika.yaml

.PHONY: up
up: docker-config
	docker compose up --build

.PHONY: up-reload
up-reload: docker-config
	find . -name '*.go' -o -name '*.yaml' | entr -rc -- make up

.PHONY: upd
upd: docker-config
	docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: run
run: local-config
	CGO_ENABLED=1 go build -race -o ./bin/ika ./cmd/ika
	./bin/ika -config ./tests/ika.yaml

.PHONY: test
test:
	go test ./...

.PHONY: e2e
e2e:
	docker compose up -d httpbun-local
	k6 run ./tests/tests.js
