# Makefile for Object Storage Gateway

.PHONY: all build test test-coverage clean

# Default target
all: build

# Build the application
build:
	go build -o homework-object-storage

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Start the application
start:
	docker-compose up -d

# Stop the application
stop:
	docker-compose down

# Clean build artifacts
clean:
	rm -f homework-object-storage
	rm -f coverage.out
	rm -f coverage.html

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - Build the application (default)"
	@echo "  build          - Build the application"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage and generate report"
	@echo "  start          - Start the application"
	@echo "  stop           - Stop the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"