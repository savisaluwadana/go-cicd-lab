# Project 22: GitOps with Flux CD - Multi-Tenant & Progressive Delivery

## 🎯 Project Overview

Implement a comprehensive GitOps platform using Flux CD for multi-tenant Kubernetes environments with progressive delivery, automated image updates, policy enforcement, and disaster recovery capabilities.

### What You'll Learn
- Flux CD architecture and components
- Multi-tenant GitOps workflows
- Progressive delivery with Flagger
- Automated image updates and notifications
- GitOps security and RBAC
- Multi-cluster management
- Disaster recovery strategies

### Technologies
- **GitOps:** Flux CD 2.2
- **Progressive Delivery:** Flagger 1.35
- **Service Mesh:** Istio or Linkerd
- **Git Platforms:** GitHub, GitLab
- **Monitoring:** Prometheus, Grafana
- **Notifications:** Slack, Microsoft Teams
- **Policy:** OPA Gatekeeper

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                      GitOps Control Plane                         │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────┐         ┌────────────────┐                  │
│  │  Git Repository │◄───────┤   Flux CD      │                  │
│  │  (Source of     │         │   Controllers  │                  │
│  │   Truth)        │         └────────┬───────┘                  │
│  │                 │                  │                          │
│  │ • infra/        │         ┌────────▼────────┐                 │
│  │ • apps/         │         │ Image Reflector │                 │
│  │ • tenants/      │         │ Image Automation│                 │
│  │ • policies/     │         └────────┬────────┘                 │
│  └────────────────┘                  │                          │
│                                       │                          │
│  ┌────────────────────────────────────▼──────────────────────┐  │
│  │              Kubernetes Clusters (Multi-Cluster)          │  │
│  │                                                             │  │
│  │  Cluster 1 (Production)      Cluster 2 (Staging)          │  │
│  │  ┌─────────────────────┐     ┌─────────────────────┐      │  │
│  │  │ Tenant A Namespace  │     │ Tenant A Namespace  │      │  │
│  │  │ ┌─────┐  ┌─────┐   │     │ ┌─────┐  ┌─────┐   │      │  │
│  │  │ │App 1│  │App 2│   │     │ │App 1│  │App 2│   │      │  │
│  │  │ └─────┘  └─────┘   │     │ └─────┘  └─────┘   │      │  │
│  │  └─────────────────────┘     └─────────────────────┘      │  │
│  │                                                             │  │
│  │  ┌─────────────────────┐     ┌─────────────────────┐      │  │
│  │  │ Tenant B Namespace  │     │ Tenant B Namespace  │      │  │
│  │  │ ┌─────┐  ┌─────┐   │     │ ┌─────┐  ┌─────┐   │      │  │
│  │  │ │App 3│  │App 4│   │     │ │App 3│  │App 4│   │      │  │
│  │  │ └─────┘  └─────┘   │     │ └─────┘  └─────┘   │      │  │
│  │  └─────────────────────┘     └─────────────────────┘      │  │
│  │                                                             │  │
│  │  ┌──────────────────────────────────────────────┐          │  │
│  │  │         Flagger (Progressive Delivery)       │          │  │
│  │  │  • Canary Analysis    • Blue/Green           │          │  │
│  │  │  • A/B Testing        • Traffic Shifting     │          │  │
│  │  └──────────────────────────────────────────────┘          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │              Observability & Notifications                  │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │  │
│  │  │Prometheus│  │ Grafana  │  │  Slack   │  │  GitHub  │  │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- Kubernetes cluster (1.28+)
- kubectl and kubeconfig
- GitHub account with PAT
- Flux CLI 2.2+
- Helm 3.x

---

## 🚀 Implementation

### Step 1: Install Flux CD

```bash
# Install Flux CLI
curl -s https://fluxcd.io/install.sh | sudo bash

# Check prerequisites
flux check --pre

# Bootstrap Flux on your cluster
export GITHUB_TOKEN=<your-github-pat>
export GITHUB_USER=<your-github-username>

flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=fleet-infra \
  --branch=main \
  --path=./clusters/production \
  --personal \
  --components-extra=image-reflector-controller,image-automation-controller

# Verify installation
flux check
kubectl get pods -n flux-system
```

### Step 2: Repository Structure

```bash
# Create GitOps repository structure
mkdir -p fleet-infra
cd fleet-infra

# Directory structure
mkdir -p {clusters,infrastructure,apps,tenants,policies}/{production,staging}

# Example structure:
# fleet-infra/
# ├── clusters/
# │   ├── production/
# │   │   └── flux-system/
# │   └── staging/
# ├── infrastructure/
# │   ├── production/
# │   │   ├── ingress-nginx/
# │   │   ├── cert-manager/
# │   │   └── monitoring/
# │   └── staging/
# ├── apps/
# │   ├── production/
# │   │   ├── tenant-a/
# │   │   └── tenant-b/
# │   └── staging/
# ├── tenants/
# │   ├── tenant-a.yaml
# │   └── tenant-b.yaml
# └── policies/
#     └── gatekeeper/
```

### Step 3: Multi-Tenant Configuration

**tenants/tenant-a.yaml**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-a
  labels:
    tenant: tenant-a
    toolkit.fluxcd.io/tenant: tenant-a
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: flux-tenant-a
  namespace: tenant-a
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: flux-tenant-a
  namespace: tenant-a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: flux-tenant-a
  namespace: tenant-a
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: tenant-a-apps
  namespace: tenant-a
spec:
  interval: 1m0s
  url: https://github.com/tenant-a/apps
  ref:
    branch: main
  secretRef:
    name: tenant-a-git-credentials
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: tenant-a-apps
  namespace: tenant-a
spec:
  interval: 10m0s
  serviceAccountName: flux-tenant-a
  sourceRef:
    kind: GitRepository
    name: tenant-a-apps
  path: ./production
  prune: true
  validation: client
  healthChecks:
    - apiVersion: apps/v1
      kind: Deployment
      name: '*'
      namespace: tenant-a
  timeout: 5m
```

### Step 4: Infrastructure Components

**infrastructure/production/ingress-nginx/kustomization.yaml**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - namespace.yaml
  - helmrelease.yaml
  - helmrepository.yaml
```

**infrastructure/production/ingress-nginx/helmrelease.yaml**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
spec:
  interval: 30m
  chart:
    spec:
      chart: ingress-nginx
      version: 4.9.x
      sourceRef:
        kind: HelmRepository
        name: ingress-nginx
        namespace: ingress-nginx
      interval: 12h
  install:
    crds: Create
  upgrade:
    crds: CreateReplace
  values:
    controller:
      replicaCount: 3
      metrics:
        enabled: true
        serviceMonitor:
          enabled: true
      podAnnotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "10254"
  test:
    enable: true
  valuesFrom:
    - kind: ConfigMap
      name: ingress-nginx-values
      optional: true
```

### Step 5: Application Deployment with GitOps

**apps/production/tenant-a/webapp/deployment.yaml**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: tenant-a
  labels:
    app: webapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: ghcr.io/tenant-a/webapp:v1.0.0  # Managed by Flux image automation
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: webapp
  namespace: tenant-a
spec:
  selector:
    app: webapp
  ports:
  - port: 80
    targetPort: 8080
```

**apps/production/tenant-a/webapp/kustomization.yaml**
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml
commonLabels:
  app: webapp
  tenant: tenant-a
images:
  - name: ghcr.io/tenant-a/webapp
    newTag: v1.0.0  # {"$imagepolicy": "tenant-a:webapp"}
```

### Step 6: Automated Image Updates

**apps/production/tenant-a/image-policy.yaml**
```yaml
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImageRepository
metadata:
  name: webapp
  namespace: tenant-a
spec:
  image: ghcr.io/tenant-a/webapp
  interval: 5m0s
  secretRef:
    name: ghcr-credentials
---
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImagePolicy
metadata:
  name: webapp
  namespace: tenant-a
spec:
  imageRepositoryRef:
    name: webapp
  policy:
    semver:
      range: 1.0.x  # Only patch updates
  filterTags:
    pattern: '^v(?P<version>.*)$'
    extract: '$version'
---
apiVersion: image.toolkit.fluxcd.io/v1beta1
kind: ImageUpdateAutomation
metadata:
  name: webapp
  namespace: tenant-a
spec:
  interval: 30m
  sourceRef:
    kind: GitRepository
    name: tenant-a-apps
  git:
    checkout:
      ref:
        branch: main
    commit:
      author:
        email: fluxcdbot@users.noreply.github.com
        name: fluxcdbot
      messageTemplate: |
        Automated image update
        
        Automation name: {{ .AutomationObject }}
        
        Files:
        {{ range $filename, $_ := .Updated.Files -}}
        - {{ $filename }}
        {{ end -}}
        
        Objects:
        {{ range $resource, $_ := .Updated.Objects -}}
        - {{ $resource.Kind }} {{ $resource.Name }}
        {{ end -}}
        
        Images:
        {{ range .Updated.Images -}}
        - {{.}}
        {{ end -}}
    push:
      branch: main
  update:
    path: ./production/tenant-a
    strategy: Setters
```

### Step 7: Progressive Delivery with Flagger

**Install Flagger**
```bash
# Add Flagger Helm repository
kubectl apply -f https://raw.githubusercontent.com/fluxcd/flagger/main/artifacts/flagger/crd.yaml

# Create Flagger HelmRelease
cat <<EOF | kubectl apply -f -
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: flagger
  namespace: flux-system
spec:
  interval: 24h
  url: https://flagger.app
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: flagger
  namespace: flux-system
spec:
  interval: 30m
  chart:
    spec:
      chart: flagger
      version: 1.35.x
      sourceRef:
        kind: HelmRepository
        name: flagger
      interval: 12h
  values:
    meshProvider: istio
    metricsServer: http://prometheus.monitoring:9090
EOF
```

**apps/production/tenant-a/webapp/canary.yaml**
```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: webapp
  namespace: tenant-a
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: webapp
  progressDeadlineSeconds: 600
  service:
    port: 80
    targetPort: 8080
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
      - name: request-success-rate
        thresholdRange:
          min: 99
        interval: 1m
      - name: request-duration
        thresholdRange:
          max: 500
        interval: 1m
    webhooks:
      - name: acceptance-test
        type: pre-rollout
        url: http://flagger-loadtester.test/
        timeout: 30s
        metadata:
          type: bash
          cmd: "curl -sd 'test' http://webapp-canary.tenant-a:80/token | grep token"
      - name: load-test
        url: http://flagger-loadtester.test/
        timeout: 5s
        metadata:
          cmd: "hey -z 1m -q 10 -c 2 http://webapp-canary.tenant-a:80/"
      - name: slack-notification
        type: post-rollout
        url: ${{ secrets.SLACK_WEBHOOK_URL }}
        metadata:
          text: "Canary deployment completed for webapp in tenant-a"
```

### Step 8: Multi-Cluster GitOps

**clusters/production/flux-system/gotk-sync.yaml**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: main
  secretRef:
    name: flux-system
  url: https://github.com/your-org/fleet-infra
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./clusters/production
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: infrastructure
  namespace: flux-system
spec:
  interval: 10m0s
  sourceRef:
    kind: GitRepository
    name: flux-system
  path: ./infrastructure/production
  prune: true
  wait: true
  timeout: 5m
  dependsOn:
    - name: flux-system
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: apps
  namespace: flux-system
spec:
  interval: 10m0s
  sourceRef:
    kind: GitRepository
    name: flux-system
  path: ./apps/production
  prune: true
  dependsOn:
    - name: infrastructure
```

### Step 9: Notifications and Alerts

**flux-system/notification.yaml**
```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta3
kind: Provider
metadata:
  name: slack
  namespace: flux-system
spec:
  type: slack
  channel: gitops-alerts
  secretRef:
    name: slack-url
---
apiVersion: notification.toolkit.fluxcd.io/v1beta3
kind: Alert
metadata:
  name: on-call-webapp
  namespace: flux-system
spec:
  providerRef:
    name: slack
  eventSeverity: info
  eventSources:
    - kind: GitRepository
      name: '*'
    - kind: Kustomization
      name: '*'
    - kind: HelmRelease
      name: '*'
    - kind: Canary
      name: '*'
      namespace: tenant-a
  summary: "Cluster production: {{ .InvolvedObject.kind }}/{{ .InvolvedObject.name }} status changed"
---
apiVersion: v1
kind: Secret
metadata:
  name: slack-url
  namespace: flux-system
stringData:
  address: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
```

### Step 10: Disaster Recovery

**backup-restore.sh**
```bash
#!/bin/bash
# Flux disaster recovery script

# Backup Flux state
backup_flux() {
  echo "Backing up Flux resources..."
  
  kubectl get gitrepositories -A -o yaml > flux-backup-gitrepos.yaml
  kubectl get kustomizations -A -o yaml > flux-backup-kustomizations.yaml
  kubectl get helmreleases -A -o yaml > flux-backup-helmreleases.yaml
  kubectl get imagepolicies -A -o yaml > flux-backup-imagepolicies.yaml
  kubectl get canaries -A -o yaml > flux-backup-canaries.yaml
  
  echo "Backup completed!"
}

# Restore Flux state
restore_flux() {
  echo "Restoring Flux resources..."
  
  # Reinstall Flux
  flux install
  
  # Wait for Flux to be ready
  kubectl wait --for=condition=Ready pods --all -n flux-system --timeout=300s
  
  # Restore resources
  kubectl apply -f flux-backup-gitrepos.yaml
  kubectl apply -f flux-backup-kustomizations.yaml
  kubectl apply -f flux-backup-helmreleases.yaml
  kubectl apply -f flux-backup-imagepolicies.yaml
  kubectl apply -f flux-backup-canaries.yaml
  
  echo "Restore completed!"
}

# Usage
case "$1" in
  backup)
    backup_flux
    ;;
  restore)
    restore_flux
    ;;
  *)
    echo "Usage: $0 {backup|restore}"
    exit 1
esac
```

---

## 🧪 Testing

```bash
# Test Flux reconciliation
flux reconcile source git flux-system
flux reconcile kustomization apps

# Check Flux status
flux get all

# View deployment status
kubectl get gitrepositories -A
kubectl get kustomizations -A
kubectl get helmreleases -A

# Monitor canary deployment
watch kubectl get canary -A

# Test image automation
# Push new image tag and watch Flux update the deployment
```

---

## 📊 Success Metrics

- **Deployment Frequency:** Multiple deployments per day
- **Lead Time for Changes:** <30 minutes from commit to production
- **Mean Time to Recovery (MTTR):** <15 minutes
- **Change Failure Rate:** <5%
- **Automated Rollbacks:** 100% via canary analysis

---

## 🎓 Best Practices

1. **Git as Single Source of Truth:** Never make manual changes to clusters
2. **Environment Parity:** Keep staging and production configurations similar
3. **Progressive Rollouts:** Always use canary or blue/green deployments
4. **Automated Testing:** Include smoke tests in Flagger webhooks
5. **Secret Management:** Use Sealed Secrets or External Secrets Operator
6. **RBAC:** Implement strict tenant isolation
7. **Monitoring:** Always monitor Flux reconciliation metrics

---

## 📚 Additional Resources

- [Flux CD Documentation](https://fluxcd.io/docs/)
- [Flagger Documentation](https://docs.flagger.app/)
- [GitOps Principles](https://opengitops.dev/)
