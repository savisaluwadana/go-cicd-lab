# Project 2: Docker Build & Push with GitHub Actions

## Overview
Build multi-architecture Docker images, implement layer caching, and push to Docker Hub and GitHub Container Registry (GHCR).

## What You'll Learn
- Docker buildx for multi-arch builds
- Layer caching strategies
- Push to multiple registries
- Image tagging strategies
- Security scanning

## Project Structure
```
02-github-actions-docker/
├── .github/
│   └── workflows/
│       ├── docker-build.yaml
│       └── docker-scan.yaml
├── app/
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── main.go
│   └── go.mod
└── README.md
```

## Implementation

### 1. Optimized Dockerfile

**app/Dockerfile:**
```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.19-alpine AS builder

WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o app .

# Final stage - minimal runtime image
FROM alpine:3.18

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app .

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./app"]
```

**app/.dockerignore:**
```
.git
.github
*.md
.env
.env.*
coverage.out
bin/
tmp/
```

### 2. Docker Build Workflow

**.github/workflows/docker-build.yaml:**
```yaml
name: Docker Build & Push

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY_DOCKERHUB: docker.io
  REGISTRY_GHCR: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      
      - name: Set up QEMU
        uses: docker/setup-qemu-action@v2
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Log in to Docker Hub
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Log in to GHCR
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: |
            ${{ env.REGISTRY_DOCKERHUB }}/${{ env.IMAGE_NAME }}
            ${{ env.REGISTRY_GHCR }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: ./app
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
      
      - name: Image digest
        run: echo ${{ steps.build.outputs.digest }}
```

### 3. Security Scanning Workflow

**.github/workflows/docker-scan.yaml:**
```yaml
name: Docker Security Scan

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 0'  # Weekly scan

jobs:
  scan:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      
      - name: Build image for scanning
        run: |
          docker build -t local-scan:latest ./app
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'local-scan:latest'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
      
      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
      
      - name: Run Snyk scan
        uses: snyk/actions/docker@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          image: local-scan:latest
          args: --severity-threshold=high
```

## Advanced Features

### Multi-Registry Push with Cosign Signing

```yaml
- name: Install Cosign
  uses: sigstore/cosign-installer@v3

- name: Sign image
  env:
    COSIGN_EXPERIMENTAL: 1
  run: |
    cosign sign --yes \
      ${{ env.REGISTRY_GHCR }}/${{ env.IMAGE_NAME }}@${{ steps.build.outputs.digest }}
```

### Build Matrix for Multiple Images

```yaml
strategy:
  matrix:
    include:
      - dockerfile: Dockerfile
        image: app
      - dockerfile: Dockerfile.worker
        image: worker
```

### Custom Build Args

```yaml
build-args: |
  GO_VERSION=1.19
  BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  VCS_REF=${{ github.sha }}
```

## Image Tagging Strategy

```
main branch        → latest, main-{sha}
develop branch     → develop, develop-{sha}
v1.2.3 tag        → 1.2.3, 1.2, 1, latest
PR #123           → pr-123
feature/auth      → feature-auth-{sha}
```

## Testing

### Test locally:
```bash
# Build multi-arch image
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 -t test:latest ./app

# Test the image
docker run -p 8080:8080 test:latest

# Scan for vulnerabilities
trivy image test:latest
```

### Test in CI:
```bash
# Trigger workflow
git tag v1.0.0
git push origin v1.0.0

# Check GitHub Actions tab
# Verify images in Docker Hub and GHCR
```

## Optimization Tips

### 1. Layer Caching
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

### 2. .dockerignore
Exclude unnecessary files to speed up context upload

### 3. Multi-stage Builds
Keep final image minimal (Alpine ~5MB vs Ubuntu ~70MB)

### 4. Build-time Variables
```dockerfile
ARG VERSION=dev
LABEL version=$VERSION
```

## Common Issues

**Issue:** Multi-arch build fails  
**Solution:** Ensure QEMU and buildx are set up

**Issue:** Push denied  
**Solution:** Check registry credentials and permissions

**Issue:** Large image size  
**Solution:** Use multi-stage builds and .dockerignore

**Issue:** Slow builds  
**Solution:** Enable GHA caching and optimize Dockerfile layers

## Security Best Practices

✅ Scan images for vulnerabilities  
✅ Use specific base image tags (not `latest`)  
✅ Run containers as non-root user  
✅ Sign images with Cosign  
✅ Use secrets for registry credentials  
✅ Enable Docker Content Trust  
✅ Regularly update base images  

## Next Steps

- Implement image promotion pipeline
- Add SBOM (Software Bill of Materials) generation
- Set up automated security patching
- Integrate with Harbor or Quay registry
- Add performance benchmarking

## Resources

- [Docker Buildx Documentation](https://docs.docker.com/buildx/working-with-buildx/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Trivy Scanner](https://aquasecurity.github.io/trivy/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
