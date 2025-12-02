#!/bin/bash

# Pre-commit hook for GoConfig Guardian
# Place this in .git/hooks/pre-commit or use with pre-commit framework

set -e

echo "🔍 Running pre-commit checks..."

# Format check
echo "📝 Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files are not formatted:"
    echo "$UNFORMATTED"
    echo "Run 'make format' to fix formatting issues"
    exit 1
fi

# Run linters
echo "🔎 Running linters..."
if ! golangci-lint run ./...; then
    echo "❌ Linting failed"
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
if ! go test -short ./...; then
    echo "❌ Tests failed"
    exit 1
fi

# Check go mod tidy
echo "📦 Checking go.mod..."
go mod tidy
if ! git diff --exit-code go.mod go.sum; then
    echo "❌ go.mod or go.sum is not tidy. Run 'go mod tidy'"
    exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0

