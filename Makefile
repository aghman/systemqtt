VERSION ?= 0.0.0-dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "-X github.com/systemqtt/systemqtt/internal/version.Version=$(VERSION) -X github.com/systemqtt/systemqtt/internal/version.GitCommit=$(GIT_COMMIT) -X github.com/systemqtt/systemqtt/internal/version.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint clean
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/systemqtt ./cmd/systemqtt

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin
