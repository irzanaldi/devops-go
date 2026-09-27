.PHONY: run build test clean

# Run server
run:
	go run cmd/api/main.go

# Build binary
build:
	go build -o bin/api cmd/api/main.go

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint (kalau sudah install golangci-lint)
lint:
	golangci-lint run
