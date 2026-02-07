# Project 1: Basic CI Pipeline with GitHub Actions

## Overview
Build a complete CI pipeline that runs tests, linting, and builds a Go application on every push and pull request.

## What You'll Learn
- GitHub Actions workflow syntax
- Running tests in CI
- Caching dependencies
- Status badges
- Branch protection rules

## Project Structure
```
01-github-actions-basic/
├── .github/
│   └── workflows/
│       └── ci.yaml
├── app/
│   ├── main.go
│   ├── main_test.go
│   └── go.mod
├── scripts/
│   └── lint.sh
└── README.md
```

## Step-by-Step Implementation

### 1. Create the Application

**app/main.go:**
```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, CI/CD World!")
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

**app/main_test.go:**
```go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()
    handler(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
    
    expected := "Hello, CI/CD World!"
    if w.Body.String() != expected {
        t.Errorf("expected %q, got %q", expected, w.Body.String())
    }
}
```

**app/go.mod:**
```
module github.com/savisaluwadana/go-cicd-lab/project01

go 1.19
```

### 2. Create CI Workflow

**.github/workflows/ci.yaml:**
```yaml
name: CI Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.19'
          cache: true
      
      - name: Download dependencies
        run: go mod download
        working-directory: ./app
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
        working-directory: ./app
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./app/coverage.out
          flags: unittests
  
  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.19'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: ./app
  
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.19'
      
      - name: Build binary
        run: |
          go build -v -o bin/app .
          chmod +x bin/app
        working-directory: ./app
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: app-binary
          path: ./app/bin/app
          retention-days: 30
```

### 3. Linting Script

**scripts/lint.sh:**
```bash
#!/bin/bash
set -e

echo "Running gofmt..."
gofmt -s -w .

echo "Running go vet..."
go vet ./...

echo "Running golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
fi

golangci-lint run ./...
echo "✅ Linting passed!"
```

## Advanced Features

### Add Status Badge
Add to your README.md:
```markdown
![CI Status](https://github.com/savisaluwadana/go-cicd-lab/workflows/CI%20Pipeline/badge.svg)
```

### Matrix Testing (Multiple Go Versions)
```yaml
test:
  strategy:
    matrix:
      go-version: ['1.19', '1.20', '1.21']
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    - run: go test ./...
```

### Conditional Steps
```yaml
- name: Deploy to staging
  if: github.ref == 'refs/heads/develop'
  run: echo "Deploying to staging..."
```

## Testing the Pipeline

1. Push code to GitHub:
```bash
git add .
git commit -m "Add CI pipeline"
git push origin main
```

2. Check Actions tab in GitHub
3. Create a pull request to see PR checks
4. View test results and coverage reports

## Common Issues & Solutions

**Issue:** Tests fail due to missing dependencies  
**Solution:** Ensure `go mod download` runs before tests

**Issue:** Cache not working  
**Solution:** Verify `go-version` matches and use `cache: true`

**Issue:** Lint errors  
**Solution:** Run `./scripts/lint.sh` locally first

## Best Practices

✅ Always run tests before build  
✅ Use caching for dependencies  
✅ Set up branch protection rules  
✅ Require status checks before merge  
✅ Keep workflows DRY with reusable workflows  
✅ Use secrets for sensitive data  
✅ Set appropriate timeouts  

## Next Steps

- Add code coverage reporting (Codecov)
- Implement semantic versioning
- Add automated changelog generation
- Set up dependency updates (Dependabot)
- Integrate with Slack/Discord notifications

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [golangci-lint](https://golangci-lint.run/)
