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

build:
	@mkdir -p bin
	CGO_ENABLED=$(CGO) go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

run: build
	./bin/$(BINARY) $(ARGS)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint: vet
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; skipping"

tidy:
	go mod tidy

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

install:
	CGO_ENABLED=$(CGO) go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(PKG)

clean:
	rm -rf bin coverage.out coverage.html

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := all
