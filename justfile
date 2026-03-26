
# Commands for clip4llm
default:
  @just --list
# Build clip4llm with Go
build:
  go build ./...

# Run tests for clip4llm with Go
test:
  go clean -testcache
  go test ./...