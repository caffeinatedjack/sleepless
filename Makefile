.PHONY: build test install clean fmt lint

# Build variables
VERSION?=1.0.0-go
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## build: Build both binaries
build:
	go build $(LDFLAGS) -o regimen ./cmd/regimen
	go build $(LDFLAGS) -o nightwatch ./cmd/nightwatch

## test: Run tests
test:
	go test -v -race ./...

## install: Install to ~/.local/bin
install:
	@go build $(LDFLAGS) -o regimen ./cmd/regimen
	@go build $(LDFLAGS) -o nightwatch ./cmd/nightwatch
	@rm -f ~/.local/bin/regimen ~/.local/bin/nightwatch
	@cp regimen ~/.local/bin/
	@cp nightwatch ~/.local/bin/
	@rm regimen nightwatch
	@go clean


## clean: Remove build artifacts
clean:
	rm -f regimen nightwatch
	go clean

## fmt: Format code
fmt:
	go fmt ./...

## lint: Run linter
lint:
	go vet ./...

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
