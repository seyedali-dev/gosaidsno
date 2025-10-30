.PHONY: test bench lint fmt clean example help

# Default target
help:
	@echo "gosaidsno - AOP without annotations"
	@echo ""
	@echo "Available targets:"
	@echo "  test      - Run all tests"
	@echo "  bench     - Run benchmarks"
	@echo "  lint      - Run golangci-lint"
	@echo "  fmt       - Format all code"
	@echo "  example   - Run basic usage example"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help"

# Run all tests
test:
	@echo "Running tests..."
	go test ./aspect/... -v -race -cover

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test ./aspect/... -bench=. -benchmem

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Run example
example:
	@echo "Running basic usage example..."
	go run examples/basic_usage.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f coverage.out

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test ./aspect/... -coverprofile=coverage.out
	go tool cover -html=coverage.out