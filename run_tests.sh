#!/bin/bash

echo "Running Aura API Client Tests"
echo "=============================="
echo ""

# Run all tests with coverage
echo "Running tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Show coverage summary
echo ""
echo "Coverage Summary:"
go tool cover -func=coverage.out | tail -n 1

# Optionally generate HTML coverage report
# go tool cover -html=coverage.out -o coverage.html
# echo "HTML coverage report generated: coverage.html"
