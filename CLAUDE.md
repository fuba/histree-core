# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `make` or `go build -ldflags "-X main.Version=$(VERSION)" -o bin/histree-core ./cmd/histree-core`
- Test all: `make test` or `go test -v ./...`
- Test single file: `go test -v ./cmd/histree-core/main_test.go`
- Test specific test: `go test -v -run TestFormatVerboseWithTimezone ./cmd/histree-core`
- Clean: `make clean`

## Code Style Guidelines
- Go version: 1.18+
- Error handling: Use `fmt.Errorf("context: %w", err)` with error wrapping
- Function naming: Use camelCase
- Indentation: Tabs, not spaces
- Return early pattern for error handling
- Always close resources (db connections, file handles) with defer
- Use type definitions for constants (OutputFormat)
- Document exported functions and types
- SQL queries should be properly indented and use parametrized queries
- Time values should be stored in UTC, but displayed in local timezone