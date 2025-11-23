BINARY?=pr-review
PKG?=./...

.PHONY: all help build run test fmt vet lint deps clean

all: build

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build       Build the binary"
	@echo "  run         Build and run the service"
	@echo "  test        Run unit tests"
	@echo "  fmt         Run gofmt (in-place)"
	@echo "  vet         Run go vet"
	@echo "  lint        Run golangci-lint (if installed)"
	@echo "  deps        Download Go modules"
	@echo "  clean       Remove generated binaries"

build:
	@echo "Building $(BINARY)..."
	@mkdir -p bin
	go build -v -o bin/$(BINARY) ./cmd

run: build
	@echo "Running $(BINARY)..."
	./bin/$(BINARY)

test:
	@echo "Running tests in ./test/..."
	go test ./test/... -v

fmt:
	@echo "Formatting code..."
	gofmt -s -w .

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running golangci-lint (if installed)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

deps:
	@echo "Downloading Go module dependencies..."
	go mod download

clean:
	@echo "Cleaning..."
	rm -rf bin
