.PHONY: test coverage coverage-html test-benchmark

# Default target
all: test

# Run all tests
test:
	go test ./...

# Run benchmark tests
test-benchmark:
	go test -bench=. -benchmem ./...

# Run tests with coverage and display summary
coverage:
	mkdir -p coverage
	go test ./... -coverprofile=coverage/coverage.out
	go tool cover -func=coverage/coverage.out

# Generate HTML coverage report and open it in the default browser
coverage-html: coverage
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	open coverage/coverage.html
