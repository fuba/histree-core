# Define the Go binary name and source directory
BINARY_NAME=histree
SOURCE_DIR=./cmd/histree

# Default target: build the binary and run tests
all: build test

# Build the Go binary
build:
	go build -o bin/$(BINARY_NAME) $(SOURCE_DIR)/main.go

# Run tests
test:
	go test -v $(SOURCE_DIR)

# Install the binary and setup configuration
install: build
	./scripts/install.sh

# Clean up generated files
clean:
	rm -f bin/$(BINARY_NAME)
	rm -f ./test_histree.db

.PHONY: all build test install clean

.PHONY: all test clean install release

VERSION := $(shell git describe --tags --always --dirty)

all: bin/histree

bin/histree: cmd/histree/*.go
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/histree ./cmd/histree

test:
	go test -v ./...

clean:
	rm -f bin/histree

install: all
	@echo "Installing histree..."
	@./scripts/install.sh

release:
	@if [ -z "$$VERSION" ]; then \
		echo "Usage: make release VERSION=v0.2.0"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working directory is not clean"; \
		exit 1; \
	fi
	@echo "Creating release $$VERSION..."
	@git tag -a "$$VERSION" -m "Release $$VERSION"
	@echo "Push the tag with: git push origin $$VERSION"
