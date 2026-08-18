BINARY   := baredeploy
PKG      := ./cmd/baredeploy
MODULE   := github.com/ivtkac/baredeploy
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(DATE)

GOFLAGS  := -trimpath
CGO      := 0

.PHONY: all build run test vet fmt lint tidy clean install cover help

all: build

build: ## Build the binary
	@mkdir -p bin
	CGO_ENABLED=$(CGO) go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

run: build ## Run the binary
	./bin/$(BINARY) $(ARGS)

test: ## Run tests
	go test ./...

vet: ## Run vet
	go vet ./...

fmt: ## Format the code
	go fmt ./...

lint: vet ## Run lint
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; skipping"

tidy: ## Run go mod tidy
	go mod tidy

cover: ## Run tests and generate coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

install: ## Install the binary
	CGO_ENABLED=$(CGO) go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(PKG)

clean: ## Clean up build artifacts
	rm -rf bin coverage.out coverage.html

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := all
