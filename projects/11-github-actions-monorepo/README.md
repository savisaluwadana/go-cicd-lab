# Project 11: GitHub Actions - Monorepo CI/CD with Selective Deployment

## 🎯 **Learning Objectives**
- Implement path-based conditional builds
- Manage multiple services in a single repository
- Optimize CI/CD with change detection
- Deploy only affected services
- Use build matrices for multiple languages

## 📋 **Project Overview**
Build a monorepo CI/CD pipeline that intelligently detects changes and only builds/deploys affected services. This is critical for large organizations with multiple teams sharing a repository.

## 🏗️ **Repository Structure**
```
monorepo/
├── services/
│   ├── api-gateway/         # Node.js service
│   ├── user-service/        # Go service
│   ├── payment-service/     # Python service
│   └── notification-service/ # Java service
├── shared/
│   ├── proto/              # Protocol buffers
│   └── config/             # Shared configs
├── .github/
│   └── workflows/
│       ├── detect-changes.yml
│       ├── build-services.yml
│       └── deploy-services.yml
└── scripts/
    ├── detect-changes.sh
    └── get-affected-services.js
```

## 🔧 **Change Detection Workflow**

### `.github/workflows/detect-changes.yml`
```yaml
name: Detect Changes

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      api-gateway: ${{ steps.changes.outputs.api-gateway }}
      user-service: ${{ steps.changes.outputs.user-service }}
      payment-service: ${{ steps.changes.outputs.payment-service }}
      notification-service: ${{ steps.changes.outputs.notification-service }}
      shared: ${{ steps.changes.outputs.shared }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Detect file changes
        uses: dorny/paths-filter@v2
        id: changes
        with:
          filters: |
            api-gateway:
              - 'services/api-gateway/**'
              - 'shared/**'
            user-service:
              - 'services/user-service/**'
              - 'shared/**'
            payment-service:
              - 'services/payment-service/**'
              - 'shared/**'
            notification-service:
              - 'services/notification-service/**'
              - 'shared/**'
            shared:
              - 'shared/**'

  build-api-gateway:
    needs: detect-changes
    if: needs.detect-changes.outputs.api-gateway == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: services/api-gateway/package-lock.json

      - name: Install dependencies
        working-directory: services/api-gateway
        run: npm ci

      - name: Run tests
        working-directory: services/api-gateway
        run: npm test

      - name: Build Docker image
        run: |
          docker build -t api-gateway:${{ github.sha }} \
            services/api-gateway

      - name: Push to registry
        run: |
          echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
          docker tag api-gateway:${{ github.sha }} ghcr.io/${{ github.repository }}/api-gateway:${{ github.sha }}
          docker push ghcr.io/${{ github.repository }}/api-gateway:${{ github.sha }}

  build-user-service:
    needs: detect-changes
    if: needs.detect-changes.outputs.user-service == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache-dependency-path: services/user-service/go.sum

      - name: Run tests
        working-directory: services/user-service
        run: go test -v ./...

      - name: Build Docker image
        run: |
          docker build -t user-service:${{ github.sha }} \
            services/user-service

      - name: Push to registry
        run: |
          docker tag user-service:${{ github.sha }} ghcr.io/${{ github.repository }}/user-service:${{ github.sha }}
          docker push ghcr.io/${{ github.repository }}/user-service:${{ github.sha }}

  build-matrix:
    needs: detect-changes
    if: needs.detect-changes.outputs.shared == 'true'
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service:
          - name: payment-service
            language: python
            version: '3.11'
            test-cmd: 'pytest tests/'
          - name: notification-service
            language: java
            version: '17'
            test-cmd: './gradlew test'
    steps:
      - uses: actions/checkout@v4

      - name: Setup ${{ matrix.service.language }}
        uses: actions/setup-python@v5
        if: matrix.service.language == 'python'
        with:
          python-version: ${{ matrix.service.version }}

      - name: Setup Java
        uses: actions/setup-java@v4
        if: matrix.service.language == 'java'
        with:
          java-version: ${{ matrix.service.version }}
          distribution: 'temurin'

      - name: Run tests
        working-directory: services/${{ matrix.service.name }}
        run: ${{ matrix.service.test-cmd }}

      - name: Build and push
        run: |
          docker build -t ${{ matrix.service.name }}:${{ github.sha }} \
            services/${{ matrix.service.name }}
          docker tag ${{ matrix.service.name }}:${{ github.sha }} \
            ghcr.io/${{ github.repository }}/${{ matrix.service.name }}:${{ github.sha }}
          docker push ghcr.io/${{ github.repository }}/${{ matrix.service.name }}:${{ github.sha }}

  deploy-services:
    needs: [build-api-gateway, build-user-service, build-matrix]
    if: always() && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy changed services
        env:
          KUBECONFIG_DATA: ${{ secrets.KUBECONFIG }}
        run: |
          echo "$KUBECONFIG_DATA" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig
          
          # Deploy only services that were built
          for service in api-gateway user-service payment-service notification-service; do
            if docker manifest inspect ghcr.io/${{ github.repository }}/$service:${{ github.sha }} > /dev/null 2>&1; then
              kubectl set image deployment/$service \
                $service=ghcr.io/${{ github.repository }}/$service:${{ github.sha }} \
                -n production
              echo "✅ Deployed $service"
            else
              echo "⏭️  Skipped $service (no changes)"
            fi
          done
```

## 📜 **Change Detection Script**

### `scripts/detect-changes.sh`
```bash
#!/bin/bash

# Get the base branch (main or develop)
BASE_BRANCH="${GITHUB_BASE_REF:-main}"
HEAD_SHA="${GITHUB_SHA}"

# Fetch the base branch
git fetch origin "$BASE_BRANCH"

# Get changed files
CHANGED_FILES=$(git diff --name-only "origin/$BASE_BRANCH" "$HEAD_SHA")

# Initialize services array
declare -A CHANGED_SERVICES

# Check each service directory
for service in api-gateway user-service payment-service notification-service; do
  if echo "$CHANGED_FILES" | grep -q "services/$service/"; then
    CHANGED_SERVICES[$service]=true
    echo "🔄 Detected changes in $service"
  fi
done

# Check shared directory (affects all services)
if echo "$CHANGED_FILES" | grep -q "shared/"; then
  echo "🔄 Detected changes in shared/ - marking all services for rebuild"
  for service in api-gateway user-service payment-service notification-service; do
    CHANGED_SERVICES[$service]=true
  done
fi

# Output results
echo "changed_services=${!CHANGED_SERVICES[@]}" >> "$GITHUB_OUTPUT"
```

## 🧪 **Advanced Features**

### 1. **Dependency Graph**
```javascript
// scripts/get-affected-services.js
const dependencyGraph = {
  'api-gateway': ['user-service', 'payment-service', 'notification-service'],
  'user-service': [],
  'payment-service': ['user-service'],
  'notification-service': ['user-service']
};

function getAffectedServices(changedServices) {
  const affected = new Set(changedServices);
  
  // Find services that depend on changed services
  for (const [service, deps] of Object.entries(dependencyGraph)) {
    if (deps.some(dep => changedServices.includes(dep))) {
      affected.add(service);
    }
  }
  
  return Array.from(affected);
}

// Usage in workflow
const changed = process.env.CHANGED_SERVICES.split(',');
const affected = getAffectedServices(changed);
console.log(JSON.stringify(affected));
```

### 2. **Parallel Testing Matrix**
```yaml
test-matrix:
  runs-on: ubuntu-latest
  strategy:
    fail-fast: false
    matrix:
      include:
        - service: api-gateway
          language: node
          version: ['18', '20']
        - service: user-service
          language: go
          version: ['1.21', '1.22']
        - service: payment-service
          language: python
          version: ['3.10', '3.11', '3.12']
  steps:
    - name: Test ${{ matrix.service }} with ${{ matrix.language }} ${{ matrix.version }}
      run: echo "Testing..."
```

## 🎯 **Key Learnings**
- ✅ Path-based change detection with `dorny/paths-filter`
- ✅ Conditional job execution with `if` expressions
- ✅ Build matrices for multiple languages
- ✅ Service dependency management
- ✅ Optimized CI/CD for large codebases

## 📊 **Performance Optimization**
- **Before**: Build all 4 services (~20 minutes)
- **After**: Build only changed services (~5-8 minutes average)
- **Savings**: 60-70% reduction in CI time

## 🚀 **Deployment Strategy**
1. Detect changes on PR creation
2. Build only affected services in parallel
3. Run integration tests for dependent services
4. Deploy to staging automatically
5. Promote to production on merge to main

## 🔍 **Troubleshooting**
- **Issue**: All services rebuild on every commit
  - **Solution**: Check `fetch-depth: 0` to get full git history
- **Issue**: Shared folder changes trigger all builds
  - **Solution**: This is intentional - shared code affects all services
- **Issue**: Dependency detection not working
  - **Solution**: Verify dependency graph in `get-affected-services.js`

## 📚 **Additional Resources**
- [Monorepo Patterns](https://monorepo.tools/)
- [GitHub Actions Path Filtering](https://github.com/dorny/paths-filter)
- [Turborepo Documentation](https://turbo.build/repo)
