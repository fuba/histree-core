.PHONY: all build test clean install release

VERSION := $(shell git describe --tags --always --dirty)

all: bin/histree

bin/histree: cmd/histree/*.go
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/histree ./cmd/histree

test:
	go test -v ./...

clean:
	rm -f bin/histree
	rm -f ./test_histree.db

install: all

release:
	@if [ -z "$$VERSION" ]; then \
		echo "Usage: make release VERSION=v0.2.1"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working directory is not clean"; \
		exit 1; \
	fi
	@echo "Creating release $$VERSION..."
	@git tag -a "$$VERSION" -m "Release $$VERSION"
	@echo "Push the tag with: git push origin $$VERSION"
