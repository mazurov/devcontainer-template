# Makefile for github.com/mazurov/devcontainer-template

# Variables
MODULE := github.com/mazurov/devcontainer-template
CLI_BIN := bin/devctmpl
CLI_DIR := ./cmd/cli
PKG_DIR := ./pkg
GO_FILES := $(shell find . -name '*.go' -type f)

# Default target
.PHONY: all
all: build

# Build the CLI binary
.PHONY: build
build: fmt vet
	go build -o $(CLI_BIN) $(CLI_DIR)

# Run the CLI
.PHONY: run
run: build
	$(CLI_BIN)

# Format all Go files
.PHONY: fmt
fmt:
	go fmt ./...

# Lint and vet Go code
.PHONY: vet
vet:
	go vet ./...

# Run tests
.PHONY: test
test:
	go test -race -cover ./...

# Tidy go.mod dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Clean binary files
.PHONY: clean
clean:
	rm -rf bin/*

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build   - Build CLI binary"
	@echo "  make run     - Run CLI binary"
	@echo "  make fmt     - Format code"
	@echo "  make vet     - Vet (lint) code"
	@echo "  make test    - Run tests with coverage"
	@echo "  make tidy    - Cleanup dependencies"
	@echo "  make clean   - Clean up binaries"
