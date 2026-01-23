BIN := "./bin"
BINARIES := calendar calendar_scheduler calendar_sender
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	for bin in $(BINARIES); do \
		go build -v -o $(BIN)/$$bin -ldflags "$(LDFLAGS)" ./cmd/$$bin; \
	done

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint

generate:
	find api -name "*.proto" | xargs -I {} protoc \
		--proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		{}

up:
	docker compose -f ./docker-compose.yml up -d

down:
	docker compose -f ./docker-compose.yml down -v

integration-tests:
	docker compose -f ./docker-compose.integration.yml up -d
	go test -race -tags=integration ./internal/integration/... || status=$$?; \
	docker compose -f ./docker-compose.integration.yml down -v; \
	exit $${status:-0}
