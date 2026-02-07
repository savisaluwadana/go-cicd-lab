# Projects 5-10: Quick Reference Guide

## Project 5: Jenkins Multibranch Pipeline
**Focus:** Automatic branch-based builds and PR validation

**Key Features:**
- Automatic Jenkinsfile discovery per branch
- Pull request validation
- Branch-specific deployments
- Automatic cleanup of stale branches

**Jenkinsfile Example:**
```groovy
pipeline {
    agent any
    stages {
        stage('Branch Detection') {
            steps {
                script {
                    if (env.BRANCH_NAME == 'main') {
                        env.DEPLOY_ENV = 'production'
                    } else if (env.BRANCH_NAME == 'develop') {
                        env.DEPLOY_ENV = 'staging'
                    } else if (env.CHANGE_ID) {
                        env.DEPLOY_ENV = 'pr-${CHANGE_ID}'
                    } else {
                        env.DEPLOY_ENV = 'feature'
                    }
                }
            }
        }
    }
}
```

---

## Project 6: Jenkins Blue-Green Deployment
**Focus:** Zero-downtime deployments with traffic switching

**Key Concepts:**
- Maintain two identical production environments (Blue/Green)
- Deploy to inactive environment
- Run tests on new deployment
- Switch traffic instantly
- Rollback capability

**Implementation:**
```groovy
stage('Deploy Green') {
    steps {
        sh 'kubectl apply -f k8s/green-deployment.yaml'
        sh 'kubectl wait --for=condition=ready pod -l version=green --timeout=300s'
    }
}
stage('Switch Traffic') {
    steps {
        input 'Switch traffic to Green?'
        sh 'kubectl patch service app -p \'{"spec":{"selector":{"version":"green"}}}\''
    }
}
```

---

## Project 7: GitLab Auto DevOps
**Focus:** Fully automated CI/CD with GitLab's built-in features

**Features:**
- Auto Test
- Auto Build
- Auto Review Apps
- Auto Deploy
- Auto Monitoring

**.gitlab-ci.yml:**
```yaml
include:
  - template: Auto-DevOps.gitlab-ci.yml

variables:
  AUTO_DEVOPS_DOMAIN: example.com
  POSTGRES_ENABLED: "true"
  POSTGRES_VERSION: "14"

production:
  extends: .auto-deploy
  environment:
    name: production
    url: https://example.com
  when: manual
  only:
    - main
```

**Custom Stages:**
```yaml
test:
  stage: test
  script:
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  coverage: '/coverage: \d+.\d+% of statements/'

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
```

---

## Project 8: GitLab Kubernetes Deployment
**Focus:** Deploy containerized apps to Kubernetes clusters

**Implementation:**
```yaml
.deploy_template:
  image: bitnami/kubectl:latest
  script:
    - kubectl config set-cluster k8s --server="${KUBE_URL}"
    - kubectl config set-credentials deployer --token="${KUBE_TOKEN}"
    - kubectl config set-context default --cluster=k8s --user=deployer
    - kubectl config use-context default
    - |
      cat <<EOF | kubectl apply -f -
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: ${CI_PROJECT_NAME}
        namespace: ${KUBE_NAMESPACE}
      spec:
        replicas: 3
        selector:
          matchLabels:
            app: ${CI_PROJECT_NAME}
        template:
          metadata:
            labels:
              app: ${CI_PROJECT_NAME}
          spec:
            containers:
            - name: app
              image: ${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}
              ports:
              - containerPort: 8080
      EOF
    - kubectl rollout status deployment/${CI_PROJECT_NAME} -n ${KUBE_NAMESPACE}

deploy:staging:
  extends: .deploy_template
  environment:
    name: staging
    url: https://staging.example.com
  variables:
    KUBE_NAMESPACE: staging
  only:
    - develop

deploy:production:
  extends: .deploy_template
  environment:
    name: production
    url: https://example.com
  variables:
    KUBE_NAMESPACE: production
  when: manual
  only:
    - main
```

---

## Project 9: GitLab Security Scanning
**Focus:** Comprehensive security testing in CI/CD

**Scanning Types:**
1. **SAST** (Static Application Security Testing)
2. **DAST** (Dynamic Application Security Testing)
3. **Dependency Scanning**
4. **Container Scanning**
5. **Secret Detection**

**.gitlab-ci.yml:**
```yaml
include:
  - template: Security/SAST.gitlab-ci.yml
  - template: Security/Dependency-Scanning.gitlab-ci.yml
  - template: Security/Container-Scanning.gitlab-ci.yml
  - template: Security/Secret-Detection.gitlab-ci.yml

variables:
  SAST_EXCLUDED_PATHS: "spec, test, tests, tmp"
  SECURE_LOG_LEVEL: "debug"

gosec-sast:
  variables:
    SAST_GOSEC_LEVEL: 2

container_scanning:
  variables:
    CS_IMAGE: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    CS_SEVERITY_THRESHOLD: high

dependency_scanning:
  variables:
    DS_EXCLUDED_PATHS: "tests/"

secret_detection:
  variables:
    SECRET_DETECTION_EXCLUDED_PATHS: "test/"

security:report:
  stage: deploy
  script:
    - |
      echo "Security Scan Summary:"
      echo "SAST: $(cat gl-sast-report.json | jq '.vulnerabilities | length') vulnerabilities"
      echo "Dependency: $(cat gl-dependency-scanning-report.json | jq '.vulnerabilities | length') vulnerabilities"
  artifacts:
    reports:
      sast: gl-sast-report.json
      dependency_scanning: gl-dependency-scanning-report.json
      container_scanning: gl-container-scanning-report.json
```

---

## Project 10: Advanced Hybrid CI/CD Platform
**Focus:** Multi-platform orchestration with Kubernetes

**Architecture:**
- GitHub Actions for code quality and tests
- Jenkins for complex builds and deployments
- GitLab CI for security scanning and monitoring
- Kubernetes for deployment target
- ArgoCD for GitOps

**Workflow:**
```
GitHub (PR) → Run tests & linting
    ↓
Jenkins (main) → Build Docker images, run integration tests
    ↓
GitLab → Security scanning (SAST, DAST, dependency scan)
    ↓
ArgoCD → Deploy to Kubernetes (GitOps)
    ↓
Monitoring (Prometheus/Grafana) → Continuous monitoring
```

**GitHub Actions (.github/workflows/pr-check.yaml):**
```yaml
name: PR Quality Check
on: [pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.19'
      - run: go test -v -race ./...
      - run: golangci-lint run
      
      # Trigger Jenkins build via API
      - name: Trigger Jenkins
        run: |
          curl -X POST https://jenkins.example.com/job/pr-build/buildWithParameters \
            --user ${{ secrets.JENKINS_USER }}:${{ secrets.JENKINS_TOKEN }} \
            --data PR_NUMBER=${{ github.event.number }}
```

**Jenkinsfile (builds and pushes to registry):**
```groovy
pipeline {
    agent { kubernetes { yaml kubernetesPodYaml() } }
    stages {
        stage('Build Multi-Arch') {
            steps {
                sh 'docker buildx build --platform linux/amd64,linux/arm64 -t image:${BUILD_TAG} --push .'
            }
        }
        stage('Integration Tests') {
            steps {
                sh 'docker-compose -f docker-compose.test.yml up --abort-on-container-exit'
            }
        }
        stage('Trigger GitLab Security Scan') {
            steps {
                sh '''
                    curl -X POST https://gitlab.com/api/v4/projects/${GITLAB_PROJECT_ID}/trigger/pipeline \
                        --form token=${GITLAB_TRIGGER_TOKEN} \
                        --form ref=main \
                        --form "variables[IMAGE_TAG]=${BUILD_TAG}"
                '''
            }
        }
    }
}
```

**GitLab CI (.gitlab-ci.yml - security & ArgoCD update):**
```yaml
stages:
  - security
  - deploy

security:scan:
  stage: security
  include:
    - template: Security/SAST.gitlab-ci.yml
    - template: Security/Container-Scanning.gitlab-ci.yml
  variables:
    CS_IMAGE: registry.example.com/app:${IMAGE_TAG}

update:argocd:
  stage: deploy
  image: bitnami/git:latest
  script:
    - git clone https://oauth2:${GITLAB_TOKEN}@gitlab.com/your-org/k8s-manifests.git
    - cd k8s-manifests
    - yq eval ".spec.template.spec.containers[0].image = \"registry.example.com/app:${IMAGE_TAG}\"" -i overlays/production/deployment.yaml
    - git config user.email "ci@example.com"
    - git config user.name "GitLab CI"
    - git add overlays/production/deployment.yaml
    - git commit -m "Update image to ${IMAGE_TAG}"
    - git push origin main
```

**ArgoCD Application (k8s-manifests/argocd-app.yaml):**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: library-manager
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://gitlab.com/your-org/k8s-manifests.git
    targetRevision: main
    path: overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

**Monitoring Integration:**
```yaml
# Prometheus alerts
groups:
  - name: deployment
    rules:
      - alert: DeploymentFailed
        expr: kube_deployment_status_replicas_available{deployment="library-manager"} < 1
        for: 5m
        annotations:
          summary: "Deployment {{ $labels.deployment }} has no available replicas"
```

---

## Comparison Matrix

| Feature | GitHub Actions | Jenkins | GitLab CI |
|---------|---------------|---------|-----------|
| Ease of Setup | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| Flexibility | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Built-in Security | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| Kubernetes Integration | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Cost (Self-hosted) | Free | Free | Free |
| Cost (Cloud) | Free tier | N/A | Free tier |
| Matrix Builds | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Marketplace/Plugins | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

---

## Common Patterns Across All Projects

### 1. Testing Strategy
```
Unit Tests → Integration Tests → E2E Tests → Security Scans → Smoke Tests
```

### 2. Deployment Flow
```
Build → Test → Scan → Deploy Dev → Deploy Staging → Manual Approval → Deploy Production
```

### 3. Rollback Strategy
```
Monitor Metrics → Detect Issues → Auto/Manual Rollback → Alert Team → Post-mortem
```

### 4. Security Checkpoints
- Secret scanning in code
- Dependency vulnerability scanning
- Container image scanning
- SAST before merge
- DAST in staging
- Runtime security in production

---

## Next Steps After Completing All Projects

1. **Implement GitOps with ArgoCD/Flux**
2. **Add Chaos Engineering (Chaos Mesh)**
3. **Implement Progressive Delivery (Flagger)**
4. **Set up Observability Stack (Prometheus, Grafana, Loki, Tempo)**
5. **Add Cost Optimization (Kubecost)**
6. **Implement Policy as Code (OPA/Kyverno)**
7. **Set up Disaster Recovery procedures**
8. **Add Performance Testing (k6, Locust)**
9. **Implement Feature Flags (LaunchDarkly, Unleash)**
10. **Create Runbooks for common scenarios**
