.PHONY: build test lint clean install-deps

BINARY_NAME=ai-dev-cli
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test -v ./...

test-unit:
	go test -v ./internal/...

test-integration:
	SKIP_INTEGRATION=0 go test -v ./tests/integration/...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run ./... --fix

clean:
	rm -rf $(BUILD_DIR)

install-deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...
	gofumpt -w .

run-improve:
	go run . improve $(ARGS)

run-test:
	go run . test $(ARGS)

help:
	@echo "Available commands:"
	@echo "  make build            - Build the binary"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make lint             - Run linter"
	@echo "  make lint-fix         - Run linter with auto-fix"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make install-deps    - Download dependencies"
	@echo "  make fmt             - Format code"
	@echo "  make run-improve     - Run improve command (Usage: make run-improve ARGS=path/to/file.go)"
	@echo "  make run-test        - Run test command (Usage: make run-test ARGS=path/to/file.go)"