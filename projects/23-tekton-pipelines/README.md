# Project 23: Cloud-Native CI/CD with Tekton Pipelines

## 🎯 Project Overview

Build a Kubernetes-native CI/CD platform using Tekton Pipelines for cloud-native application delivery. Implement reusable tasks, parallel execution, artifact management, security scanning, and multi-cloud deployments.

### What You'll Learn
- Tekton architecture and CRDs
- Cloud-native pipeline design
- Reusable Tasks and ClusterTasks
- Parallel and sequential task execution
- Workspaces and volume management
- Tekton Triggers for event-driven pipelines
- Integration with artifact registries
- Multi-cloud deployment strategies

### Technologies
- **CI/CD:** Tekton Pipelines 0.53, Tekton Triggers
- **Kubernetes:** 1.28+
- **Artifact Storage:** Harbor, JFrog Artifactory
- **Security:** Trivy, Cosign, SBOM generation
- **Observability:** Tekton Dashboard, Prometheus
- **Git:** GitHub, GitLab

---

## 🏗️ Architecture

```
┌───────────────────────────────────────────────────────────────────┐
│                    Tekton CI/CD Platform                          │
├───────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌─────────────┐          ┌──────────────┐                       │
│  │  Git Event  │─────────►│    Tekton    │                       │
│  │  (Webhook)  │          │   Triggers   │                       │
│  └─────────────┘          └──────┬───────┘                       │
│                                   │                               │
│                          ┌────────▼────────┐                      │
│                          │  EventListener  │                      │
│                          │  TriggerBinding │                      │
│                          │  TriggerTemplate│                      │
│                          └────────┬────────┘                      │
│                                   │                               │
│                          ┌────────▼────────┐                      │
│                          │   PipelineRun   │                      │
│                          └────────┬────────┘                      │
│                                   │                               │
│  ┌────────────────────────────────▼────────────────────────────┐ │
│  │                    Tekton Pipeline                          │ │
│  │                                                              │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │ │
│  │  │  Clone   │─►│  Build   │─►│   Test   │─►│  Scan    │   │ │
│  │  │   Repo   │  │  Image   │  │   Unit   │  │  Security│   │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │ │
│  │       │                                           │         │ │
│  │       │         ┌──────────┐  ┌──────────┐       │         │ │
│  │       └────────►│  Publish │─►│  Deploy  │◄──────┘         │ │
│  │                 │  Artifact│  │   K8s    │                 │ │
│  │                 └──────────┘  └──────────┘                 │ │
│  │                                                              │ │
│  │  Workspaces:                                                │ │
│  │  • source-code    • cache       • artifacts                │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │             Shared Resources & Storage                      │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐           │  │
│  │  │   Harbor   │  │    PVC     │  │   Secrets  │           │  │
│  │  │  Registry  │  │  Workspace │  │   Manager  │           │  │
│  │  └────────────┘  └────────────┘  └────────────┘           │  │
│  └─────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- Kubernetes cluster (1.28+)
- kubectl and kubeconfig
- Tekton CLI (tkn) installed
- Container registry access
- 4GB+ RAM available

---

## 🚀 Implementation

### Step 1: Install Tekton Pipelines

```bash
# Install Tekton Pipelines
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

# Verify installation
kubectl get pods -n tekton-pipelines

# Install Tekton Triggers
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml

# Install Tekton Dashboard
kubectl apply -f https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml

# Access dashboard
kubectl port-forward -n tekton-pipelines svc/tekton-dashboard 9097:9097

# Install Tekton CLI
brew install tektoncd-cli  # macOS
# or
curl -LO https://github.com/tektoncd/cli/releases/download/v0.33.0/tkn_0.33.0_Linux_x86_64.tar.gz
tar xvzf tkn_0.33.0_Linux_x86_64.tar.gz -C /usr/local/bin/ tkn
```

### Step 2: Create Reusable Tasks

**tasks/git-clone.yaml**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: git-clone
  labels:
    app.kubernetes.io/version: "0.9"
spec:
  description: Clone a git repository
  params:
    - name: url
      description: Repository URL to clone
      type: string
    - name: revision
      description: Revision to checkout (branch, tag, sha, ref)
      type: string
      default: main
    - name: subdirectory
      description: Subdirectory inside workspace to clone the repo
      type: string
      default: ""
  workspaces:
    - name: output
      description: The git repo will be cloned here
  results:
    - name: commit
      description: The precise commit SHA that was fetched
    - name: url
      description: The URL that was fetched
  steps:
    - name: clone
      image: gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:latest
      env:
        - name: PARAM_URL
          value: $(params.url)
        - name: PARAM_REVISION
          value: $(params.revision)
        - name: PARAM_SUBDIR
          value: $(params.subdirectory)
        - name: WORKSPACE_OUTPUT_PATH
          value: $(workspaces.output.path)
      script: |
        #!/bin/sh
        set -eu
        
        CHECKOUT_DIR="${WORKSPACE_OUTPUT_PATH}/${PARAM_SUBDIR}"
        
        /ko-app/git-init \
          -url="${PARAM_URL}" \
          -revision="${PARAM_REVISION}" \
          -path="${CHECKOUT_DIR}"
        
        cd "${CHECKOUT_DIR}"
        RESULT_SHA="$(git rev-parse HEAD)"
        
        printf "%s" "${RESULT_SHA}" > $(results.commit.path)
        printf "%s" "${PARAM_URL}" > $(results.url.path)
```

**tasks/buildah-build.yaml**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: buildah-build
  labels:
    app.kubernetes.io/version: "0.6"
spec:
  description: Build and push container image using Buildah
  params:
    - name: IMAGE
      description: Reference of the image to build
    - name: DOCKERFILE
      description: Path to Dockerfile
      default: ./Dockerfile
    - name: CONTEXT
      description: Build context directory
      default: .
    - name: TLSVERIFY
      description: Verify TLS
      default: "true"
    - name: FORMAT
      description: Image manifest format (oci or docker)
      default: oci
  workspaces:
    - name: source
    - name: dockerconfig
      optional: true
  results:
    - name: IMAGE_DIGEST
      description: Digest of the built image
    - name: IMAGE_URL
      description: URL of the built image
  steps:
    - name: build
      image: quay.io/buildah/stable:latest
      workingDir: $(workspaces.source.path)
      script: |
        #!/bin/bash
        set -e
        
        buildah --storage-driver=vfs bud \
          --format=$(params.FORMAT) \
          --tls-verify=$(params.TLSVERIFY) \
          --no-cache \
          -f $(params.DOCKERFILE) \
          -t $(params.IMAGE) \
          $(params.CONTEXT)
      volumeMounts:
        - name: varlibcontainers
          mountPath: /var/lib/containers
      securityContext:
        privileged: true
    
    - name: push
      image: quay.io/buildah/stable:latest
      script: |
        #!/bin/bash
        set -e
        
        if [ -f $(workspaces.dockerconfig.path)/config.json ]; then
          export DOCKER_CONFIG="$(workspaces.dockerconfig.path)"
        fi
        
        buildah --storage-driver=vfs push \
          --tls-verify=$(params.TLSVERIFY) \
          --digestfile /tmp/image-digest \
          $(params.IMAGE) \
          docker://$(params.IMAGE)
        
        cat /tmp/image-digest | tee $(results.IMAGE_DIGEST.path)
        echo -n "$(params.IMAGE)" | tee $(results.IMAGE_URL.path)
      volumeMounts:
        - name: varlibcontainers
          mountPath: /var/lib/containers
      securityContext:
        privileged: true
  
  volumes:
    - name: varlibcontainers
      emptyDir: {}
```

**tasks/trivy-scan.yaml**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: trivy-scanner
spec:
  description: Scan container images for vulnerabilities using Trivy
  params:
    - name: IMAGE
      description: Image to scan
    - name: SEVERITY
      description: Severities to scan for
      default: "CRITICAL,HIGH"
    - name: EXIT_CODE
      description: Exit code when vulnerabilities are found
      default: "0"
  results:
    - name: SCAN_RESULT
      description: Summary of scan results
  steps:
    - name: scan
      image: aquasec/trivy:latest
      script: |
        #!/bin/sh
        set -e
        
        trivy image \
          --severity $(params.SEVERITY) \
          --exit-code $(params.EXIT_CODE) \
          --format json \
          --output /tmp/trivy-results.json \
          $(params.IMAGE)
        
        # Extract summary
        CRITICAL=$(cat /tmp/trivy-results.json | jq '[.Results[].Vulnerabilities[]? | select(.Severity=="CRITICAL")] | length')
        HIGH=$(cat /tmp/trivy-results.json | jq '[.Results[].Vulnerabilities[]? | select(.Severity=="HIGH")] | length')
        
        echo "Critical: ${CRITICAL}, High: ${HIGH}" | tee $(results.SCAN_RESULT.path)
        
        # Fail if critical vulnerabilities found
        if [ "$CRITICAL" -gt "0" ]; then
          echo "❌ Critical vulnerabilities found!"
          exit 1
        fi
```

**tasks/kubectl-deploy.yaml**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: kubectl-deploy
spec:
  description: Deploy to Kubernetes using kubectl
  params:
    - name: MANIFEST_DIR
      description: Directory containing Kubernetes manifests
      default: ./k8s
    - name: NAMESPACE
      description: Kubernetes namespace
      default: default
    - name: IMAGE
      description: New image to deploy
  workspaces:
    - name: source
  steps:
    - name: deploy
      image: bitnami/kubectl:latest
      script: |
        #!/bin/bash
        set -e
        
        cd $(workspaces.source.path)/$(params.MANIFEST_DIR)
        
        # Update image in manifests
        find . -name "*.yaml" -exec sed -i "s|image:.*|image: $(params.IMAGE)|g" {} \;
        
        # Apply manifests
        kubectl apply -n $(params.NAMESPACE) -f .
        
        # Wait for rollout
        kubectl rollout status deployment -n $(params.NAMESPACE) --timeout=5m
```

### Step 3: Create CI/CD Pipeline

**pipelines/build-deploy-pipeline.yaml**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: build-deploy-pipeline
spec:
  params:
    - name: repo-url
      type: string
    - name: revision
      type: string
      default: main
    - name: image-name
      type: string
    - name: image-tag
      type: string
      default: latest
    - name: deploy-namespace
      type: string
      default: default
  
  workspaces:
    - name: shared-workspace
    - name: docker-credentials
  
  tasks:
    - name: fetch-repository
      taskRef:
        name: git-clone
      params:
        - name: url
          value: $(params.repo-url)
        - name: revision
          value: $(params.revision)
      workspaces:
        - name: output
          workspace: shared-workspace
    
    - name: run-tests
      runAfter:
        - fetch-repository
      taskRef:
        name: golang-test
      workspaces:
        - name: source
          workspace: shared-workspace
    
    - name: build-image
      runAfter:
        - run-tests
      taskRef:
        name: buildah-build
      params:
        - name: IMAGE
          value: "$(params.image-name):$(params.image-tag)"
        - name: DOCKERFILE
          value: ./Dockerfile
      workspaces:
        - name: source
          workspace: shared-workspace
        - name: dockerconfig
          workspace: docker-credentials
    
    - name: scan-image
      runAfter:
        - build-image
      taskRef:
        name: trivy-scanner
      params:
        - name: IMAGE
          value: "$(params.image-name):$(params.image-tag)"
        - name: SEVERITY
          value: "CRITICAL,HIGH"
    
    - name: deploy-to-cluster
      runAfter:
        - scan-image
      taskRef:
        name: kubectl-deploy
      params:
        - name: MANIFEST_DIR
          value: ./k8s
        - name: NAMESPACE
          value: $(params.deploy-namespace)
        - name: IMAGE
          value: "$(params.image-name):$(params.image-tag)"
      workspaces:
        - name: source
          workspace: shared-workspace
  
  finally:
    - name: cleanup
      taskRef:
        name: cleanup-workspace
      workspaces:
        - name: source
          workspace: shared-workspace
```

### Step 4: Configure Tekton Triggers

**triggers/eventlistener.yaml**
```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: github-listener
spec:
  serviceAccountName: tekton-triggers-sa
  triggers:
    - name: github-push
      interceptors:
        - ref:
            name: "github"
          params:
            - name: "secretRef"
              value:
                secretName: github-secret
                secretKey: secretToken
            - name: "eventTypes"
              value: ["push"]
        - ref:
            name: "cel"
          params:
            - name: "filter"
              value: "body.ref.startsWith('refs/heads/main')"
      bindings:
        - ref: github-push-binding
      template:
        ref: build-deploy-template
---
apiVersion: v1
kind: Service
metadata:
  name: el-github-listener
spec:
  selector:
    eventlistener: github-listener
  ports:
    - port: 8080
      targetPort: 8080
  type: LoadBalancer
```

**triggers/triggerbinding.yaml**
```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: github-push-binding
spec:
  params:
    - name: git-repo-url
      value: $(body.repository.clone_url)
    - name: git-revision
      value: $(body.after)
    - name: git-repo-name
      value: $(body.repository.name)
    - name: git-commit-message
      value: $(body.head_commit.message)
    - name: git-committer
      value: $(body.head_commit.committer.name)
```

**triggers/triggertemplate.yaml**
```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: build-deploy-template
spec:
  params:
    - name: git-repo-url
    - name: git-revision
    - name: git-repo-name
    - name: git-commit-message
    - name: git-committer
  resourcetemplates:
    - apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
        generateName: build-deploy-run-
        labels:
          app: $(tt.params.git-repo-name)
          commit: $(tt.params.git-revision)
      spec:
        serviceAccountName: pipeline-sa
        pipelineRef:
          name: build-deploy-pipeline
        params:
          - name: repo-url
            value: $(tt.params.git-repo-url)
          - name: revision
            value: $(tt.params.git-revision)
          - name: image-name
            value: "harbor.example.com/library/$(tt.params.git-repo-name)"
          - name: image-tag
            value: "$(tt.params.git-revision)"
        workspaces:
          - name: shared-workspace
            volumeClaimTemplate:
              spec:
                accessModes:
                  - ReadWriteOnce
                resources:
                  requests:
                    storage: 1Gi
          - name: docker-credentials
            secret:
              secretName: docker-credentials
```

### Step 5: RBAC and Service Accounts

**rbac/serviceaccount.yaml**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline-sa
  namespace: default
secrets:
  - name: docker-credentials
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pipeline-role
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps", "secrets"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: ["tekton.dev"]
    resources: ["tasks", "taskruns", "pipelines", "pipelineruns"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pipeline-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pipeline-role
subjects:
  - kind: ServiceAccount
    name: pipeline-sa
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tekton-triggers-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: tekton-triggers-role
rules:
  - apiGroups: ["triggers.tekton.dev"]
    resources: ["eventlisteners", "triggerbindings", "triggertemplates"]
    verbs: ["get", "list"]
  - apiGroups: ["tekton.dev"]
    resources: ["pipelineruns", "pipelineresources"]
    verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tekton-triggers-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: tekton-triggers-role
subjects:
  - kind: ServiceAccount
    name: tekton-triggers-sa
```

### Step 6: Run Pipeline

```bash
# Create PipelineRun manually
tkn pipeline start build-deploy-pipeline \
  --param repo-url=https://github.com/your-org/your-app \
  --param revision=main \
  --param image-name=harbor.example.com/library/your-app \
  --param image-tag=v1.0.0 \
  --workspace name=shared-workspace,volumeClaimTemplateFile=workspace-template.yaml \
  --workspace name=docker-credentials,secret=docker-credentials \
  --serviceaccount=pipeline-sa \
  --showlog

# List pipeline runs
tkn pipelinerun list

# View logs
tkn pipelinerun logs <pipelinerun-name> -f

# Describe pipeline run
tkn pipelinerun describe <pipelinerun-name>

# Delete old pipeline runs (keep last 10)
tkn pipelinerun delete --keep 10
```

---

## 🧪 Testing

```bash
# Test individual task
tkn task start git-clone \
  --param url=https://github.com/your-org/your-app \
  --param revision=main \
  --workspace name=output,emptyDir="" \
  --showlog

# Trigger pipeline via webhook (simulate GitHub push)
curl -X POST http://el-github-listener:8080 \
  -H 'Content-Type: application/json' \
  -H 'X-GitHub-Event: push' \
  -d @test-payload.json

# Monitor pipeline execution
watch tkn pipelinerun list
```

---

## 📊 Success Metrics

- **Pipeline Success Rate:** >95%
- **Average Build Time:** <10 minutes
- **Parallel Task Execution:** 3+ tasks concurrently
- **Resource Utilization:** CPU <70%, Memory <80%
- **Pipeline Reusability:** 80% shared tasks across pipelines

---

## 🎓 Best Practices

1. **Reusable Tasks:** Create ClusterTasks for organization-wide reuse
2. **Workspace Management:** Use PVCs or emptyDir efficiently
3. **Parallel Execution:** Leverage `runAfter` for task dependencies
4. **Secret Management:** Use Kubernetes secrets, never hardcode
5. **Resource Limits:** Set CPU/memory limits on all tasks
6. **Idempotency:** Ensure tasks can be re-run safely
7. **Logging:** Stream logs to external systems for persistence

---

## 📚 Additional Resources

- [Tekton Documentation](https://tekton.dev/docs/)
- [Tekton Catalog](https://hub.tekton.dev/)
- [Cloud Native CI/CD with Tekton](https://www.manning.com/books/continuous-delivery-for-kubernetes)
