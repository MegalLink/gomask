#!/bin/bash

# Create a coverage directory if it doesn't exist
mkdir -p coverage

# Run tests with coverage for all packages
go test ./... -coverprofile=coverage/coverage.out

# Display coverage statistics
go tool cover -func=coverage/coverage.out

# Generate HTML report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

echo "Coverage report generated at coverage/coverage.html"
