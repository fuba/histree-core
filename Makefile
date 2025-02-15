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
