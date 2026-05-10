.PHONY: fmt lint test ci

fmt:
	golangci-lint fmt

lint:
	golangci-lint run ./...

test:
	go test ./...

ci:
	golangci-lint fmt --diff
	golangci-lint run ./...
	go test ./...
