GO=go
MIN_GO_VERSION=1.26

BASE_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
COMMIT_COUNT=$(shell git rev-list --count HEAD 2>/dev/null || echo "0")
COMMIT_HASH=$(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
VERSION?=$(BASE_VERSION)+git$(COMMIT_COUNT).$(COMMIT_HASH)

.PHONY: all build test cover fmt vet lint deps mocks check-go version help

all: build

build:
	CGO_ENABLED=0 $(GO) build ./...

test:
	CGO_ENABLED=0 $(GO) test ./...

cover:
	CGO_ENABLED=0 $(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

fmt:
	$(GO) fmt ./...

vet:
	CGO_ENABLED=0 $(GO) vet ./...

lint:
	golangci-lint run

deps:
	$(GO) mod tidy
	$(GO) mod download

mocks:
	mockery

check-go:
	@$(GO) version | grep -qE "go($(MIN_GO_VERSION)|1\.2[7-9]|[2-9]\.)" || { echo "Go $(MIN_GO_VERSION)+ required"; exit 1; }

version:
	@echo "$(VERSION)"

help:
	@echo "Targets: build test cover fmt vet lint deps mocks check-go version"
