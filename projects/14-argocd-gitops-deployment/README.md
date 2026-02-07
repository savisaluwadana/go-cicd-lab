# Project 14: ArgoCD GitOps Deployment - Continuous Delivery for Kubernetes

## 🎯 **Learning Objectives**
- Implement GitOps with ArgoCD
- Automated Kubernetes deployments
- Multi-cluster management
- Progressive delivery with Argo Rollouts
- Automated sync and health checks

## 📋 **Project Overview**
Build a complete GitOps deployment pipeline using ArgoCD for continuous delivery to Kubernetes clusters. Implement automated deployments, rollbacks, and progressive delivery strategies like canary and blue-green deployments.

## 🏗️ **Repository Structure**
```
gitops-infrastructure/
├── apps/
│   ├── dev/
│   │   ├── application.yaml
│   │   └── kustomization.yaml
│   ├── staging/
│   │   ├── application.yaml
│   │   └── kustomization.yaml
│   └── production/
│       ├── application.yaml
│       └── kustomization.yaml
├── base/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   └── kustomization.yaml
├── overlays/
│   ├── dev/
│   ├── staging/
│   └── production/
├── argocd/
│   ├── projects/
│   │   └── app-project.yaml
│   ├── applications/
│   │   ├── app-dev.yaml
│   │   ├── app-staging.yaml
│   │   └── app-production.yaml
│   └── rollouts/
│       ├── canary-rollout.yaml
│       └── bluegreen-rollout.yaml
└── .github/
    └── workflows/
        ├── validate-manifests.yml
        └── promote-image.yml
```

## 🔧 **ArgoCD Application Configuration**

### `argocd/applications/app-production.yaml`
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: library-manager-production
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: library-manager
  
  source:
    repoURL: https://github.com/your-org/gitops-infrastructure.git
    targetRevision: main
    path: overlays/production
    
    # Kustomize specific config
    kustomize:
      images:
        - library-manager=ghcr.io/your-org/library-manager:v1.2.3
      
      # Common labels for all resources
      commonLabels:
        environment: production
        managed-by: argocd
  
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  
  syncPolicy:
    automated:
      prune: true        # Delete resources not in git
      selfHeal: true     # Force sync if cluster state differs
      allowEmpty: false  # Prevent deletion of all resources
    
    syncOptions:
      - CreateNamespace=true
      - PruneLast=true   # Delete resources last
      - RespectIgnoreDifferences=true
    
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  
  # Health check configuration
  ignoreDifferences:
    - group: apps
      kind: Deployment
      jsonPointers:
        - /spec/replicas  # Ignore HPA-managed replicas
  
  # Notifications
  revisionHistoryLimit: 10
```

### `argocd/projects/app-project.yaml`
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: library-manager
  namespace: argocd
spec:
  description: Library Manager Application Project
  
  # Source repositories
  sourceRepos:
    - 'https://github.com/your-org/gitops-infrastructure.git'
    - 'https://charts.bitnami.com/bitnami'
  
  # Destination clusters and namespaces
  destinations:
    - namespace: 'dev'
      server: https://kubernetes.default.svc
    - namespace: 'staging'
      server: https://kubernetes.default.svc
    - namespace: 'production'
      server: https://kubernetes.default.svc
  
  # Cluster resource whitelist
  clusterResourceWhitelist:
    - group: ''
      kind: Namespace
    - group: rbac.authorization.k8s.io
      kind: ClusterRole
    - group: rbac.authorization.k8s.io
      kind: ClusterRoleBinding
  
  # Namespace resource whitelist
  namespaceResourceWhitelist:
    - group: ''
      kind: Service
    - group: ''
      kind: ConfigMap
    - group: ''
      kind: Secret
    - group: apps
      kind: Deployment
    - group: apps
      kind: StatefulSet
    - group: networking.k8s.io
      kind: Ingress
    - group: autoscaling
      kind: HorizontalPodAutoscaler
  
  roles:
    - name: developer
      description: Developers can sync dev environment
      policies:
        - p, proj:library-manager:developer, applications, sync, library-manager/dev, allow
        - p, proj:library-manager:developer, applications, get, library-manager/*, allow
      groups:
        - developers
    
    - name: operator
      description: Operators can sync all environments
      policies:
        - p, proj:library-manager:operator, applications, *, library-manager/*, allow
      groups:
        - platform-team
```

## 🎯 **Canary Deployment with Argo Rollouts**

### `argocd/rollouts/canary-rollout.yaml`
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: library-manager
  namespace: production
spec:
  replicas: 10
  revisionHistoryLimit: 3
  
  selector:
    matchLabels:
      app: library-manager
  
  template:
    metadata:
      labels:
        app: library-manager
        version: stable
    spec:
      containers:
        - name: library-manager
          image: ghcr.io/your-org/library-manager:v1.2.3
          ports:
            - containerPort: 8080
              name: http
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
  
  strategy:
    canary:
      # Canary service (receives canary traffic)
      canaryService: library-manager-canary
      # Stable service (receives stable traffic)
      stableService: library-manager-stable
      
      trafficRouting:
        istio:
          virtualService:
            name: library-manager-vsvc
            routes:
              - primary
      
      steps:
        # Step 1: Deploy canary with 10% traffic
        - setWeight: 10
        - pause:
            duration: 5m
        
        # Step 2: Run analysis
        - analysis:
            templates:
              - templateName: success-rate
              - templateName: latency
            args:
              - name: service-name
                value: library-manager-canary
        
        # Step 3: Increase to 25% traffic
        - setWeight: 25
        - pause:
            duration: 5m
        
        # Step 4: Run analysis again
        - analysis:
            templates:
              - templateName: success-rate
              - templateName: latency
        
        # Step 5: Increase to 50% traffic
        - setWeight: 50
        - pause:
            duration: 5m
        
        # Step 6: Increase to 75% traffic
        - setWeight: 75
        - pause:
            duration: 5m
        
        # Step 7: Full rollout
        - setWeight: 100
      
      # Anti-affinity for canary and stable pods
      antiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution: {}
      
      # Automatic rollback on metrics failure
      analysis:
        successfulRunHistoryLimit: 5
        unsuccessfulRunHistoryLimit: 5
```

### `argocd/rollouts/analysis-template.yaml`
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
  namespace: production
spec:
  metrics:
    - name: success-rate
      interval: 1m
      successCondition: result >= 0.95
      failureLimit: 3
      provider:
        prometheus:
          address: http://prometheus.monitoring:9090
          query: |
            sum(rate(http_requests_total{
              service="{{args.service-name}}",
              status=~"2.."
            }[5m]))
            /
            sum(rate(http_requests_total{
              service="{{args.service-name}}"
            }[5m]))

---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: latency
  namespace: production
spec:
  metrics:
    - name: p95-latency
      interval: 1m
      successCondition: result <= 500
      failureLimit: 3
      provider:
        prometheus:
          address: http://prometheus.monitoring:9090
          query: |
            histogram_quantile(0.95,
              sum(rate(http_request_duration_seconds_bucket{
                service="{{args.service-name}}"
              }[5m])) by (le)
            ) * 1000
```

## 🔵🟢 **Blue-Green Deployment**

### `argocd/rollouts/bluegreen-rollout.yaml`
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: library-manager-bg
  namespace: production
spec:
  replicas: 5
  
  selector:
    matchLabels:
      app: library-manager
  
  template:
    metadata:
      labels:
        app: library-manager
    spec:
      containers:
        - name: library-manager
          image: ghcr.io/your-org/library-manager:latest
          ports:
            - containerPort: 8080
  
  strategy:
    blueGreen:
      # Service for active (blue) version
      activeService: library-manager-active
      # Service for preview (green) version
      previewService: library-manager-preview
      
      # Automatically promote after successful preview
      autoPromotionEnabled: false
      autoPromotionSeconds: 30
      
      # Scale down old version after promotion
      scaleDownDelaySeconds: 30
      scaleDownDelayRevisionLimit: 2
      
      # Anti-affinity between blue and green
      antiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution: {}
      
      # Pre-promotion analysis
      prePromotionAnalysis:
        templates:
          - templateName: smoke-test
        args:
          - name: service-name
            value: library-manager-preview
      
      # Post-promotion analysis
      postPromotionAnalysis:
        templates:
          - templateName: success-rate
          - templateName: latency
        args:
          - name: service-name
            value: library-manager-active
```

## 🔄 **GitHub Actions Integration**

### `.github/workflows/promote-image.yml`
```yaml
name: Promote Image

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to promote to'
        required: true
        type: choice
        options:
          - dev
          - staging
          - production
      image_tag:
        description: 'Image tag to promote'
        required: true

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout GitOps repo
        uses: actions/checkout@v4
        with:
          repository: your-org/gitops-infrastructure
          token: ${{ secrets.GITOPS_TOKEN }}

      - name: Update image tag
        run: |
          cd overlays/${{ github.event.inputs.environment }}
          kustomize edit set image \
            library-manager=ghcr.io/your-org/library-manager:${{ github.event.inputs.image_tag }}

      - name: Commit and push
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add .
          git commit -m "Promote image ${{ github.event.inputs.image_tag }} to ${{ github.event.inputs.environment }}"
          git push

      - name: Wait for ArgoCD sync
        run: |
          # Install ArgoCD CLI
          curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
          chmod +x argocd
          
          # Login to ArgoCD
          ./argocd login ${{ secrets.ARGOCD_SERVER }} \
            --username admin \
            --password ${{ secrets.ARGOCD_PASSWORD }} \
            --insecure
          
          # Wait for sync
          ./argocd app wait library-manager-${{ github.event.inputs.environment }} \
            --timeout 600 \
            --health

      - name: Notify deployment
        uses: slackapi/slack-github-action@v1
        with:
          channel-id: 'deployments'
          slack-message: |
            ✅ Successfully promoted image to ${{ github.event.inputs.environment }}
            Image: ghcr.io/your-org/library-manager:${{ github.event.inputs.image_tag }}
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

## 🎯 **Key Learnings**
- ✅ GitOps principles with ArgoCD
- ✅ Automated sync and self-healing
- ✅ Progressive delivery (canary, blue-green)
- ✅ Analysis templates with Prometheus
- ✅ Multi-environment management
- ✅ Automated rollbacks

## 📊 **Benefits**
- **Declarative**: Git as single source of truth
- **Automated**: Self-healing and auto-sync
- **Auditable**: All changes tracked in Git
- **Secure**: RBAC and policy enforcement
- **Reliable**: Progressive rollouts with metrics

## 🚀 **ArgoCD CLI Commands**
```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Create application
argocd app create library-manager \
  --repo https://github.com/your-org/gitops-infrastructure.git \
  --path overlays/production \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace production

# Sync application
argocd app sync library-manager

# Rollback to previous version
argocd app rollback library-manager

# View application status
argocd app get library-manager
```

## 📚 **Additional Resources**
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Argo Rollouts](https://argoproj.github.io/argo-rollouts/)
- [GitOps Principles](https://www.gitops.tech/)
