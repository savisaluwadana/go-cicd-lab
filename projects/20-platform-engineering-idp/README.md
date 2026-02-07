# Project 20: Platform Engineering - Internal Developer Platform (IDP)

## 🎯 **Learning Objectives**
- Build complete Internal Developer Platform
- Implement self-service infrastructure provisioning
- Create golden path templates with Backstage
- Automate environment management
- Build developer portals and documentation
- Implement platform metrics and SLOs

## 📋 **Project Overview**
Design and build a comprehensive Internal Developer Platform (IDP) that provides self-service infrastructure, standardized deployment workflows, comprehensive documentation, and observability - enabling developers to ship code faster while maintaining security and compliance.

## 🏗️ **Platform Architecture**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Developer Portal (Backstage)                  │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌─────────────┐   │
│  │ Service  │  │ Templates │  │ Docs     │  │  Metrics    │   │
│  │ Catalog  │  │ (Cookiec.)│  │ (TechD.) │  │ Dashboard   │   │
│  └────┬─────┘  └─────┬─────┘  └────┬─────┘  └──────┬──────┘   │
└───────┼──────────────┼─────────────┼────────────────┼──────────┘
        │              │             │                │
┌───────▼──────────────▼─────────────▼────────────────▼──────────┐
│                Platform Orchestration Layer                      │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌─────────────┐   │
│  │Crossplane│  │  ArgoCD   │  │Terraform │  │   Vault     │   │
│  │ (Infra)  │  │  (Deploy) │  │  (IaC)   │  │  (Secrets)  │   │
│  └────┬─────┘  └─────┬─────┘  └────┬─────┘  └──────┬──────┘   │
└───────┼──────────────┼─────────────┼────────────────┼──────────┘
        │              │             │                │
┌───────▼──────────────▼─────────────▼────────────────▼──────────┐
│                   Kubernetes Platform Layer                      │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌─────────────┐   │
│  │Multi-Tenancy │ Namespace│  │ Network  │  │   Storage   │   │
│  │  (vCluster) │  Quotas  │  │ Policies │  │   Classes   │   │
│  └────┬─────┘  └─────┬─────┘  └────┬─────┘  └──────┬──────┘   │
└───────┼──────────────┼─────────────┼────────────────┼──────────┘
        │              │             │                │
┌───────▼──────────────▼─────────────▼────────────────▼──────────┐
│                    Observability Stack                           │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌─────────────┐   │
│  │Prometheus│  │  Grafana  │  │  Loki    │  │   Jaeger    │   │
│  │ (Metrics)│  │  (Viz)    │  │  (Logs)  │  │  (Traces)   │   │
│  └──────────┘  └───────────┘  └──────────┘  └─────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## 🎨 **Backstage Developer Portal**

### Install Backstage
```bash
# Create Backstage app
npx @backstage/create-app@latest

cd my-backstage-app

# Install dependencies
yarn install

# Configure backend
cat > app-config.production.yaml <<EOF
app:
  title: Platform Engineering Portal
  baseUrl: https://backstage.example.com

organization:
  name: Your Company

backend:
  baseUrl: https://backstage.example.com
  listen:
    port: 7007
  database:
    client: pg
    connection:
      host: \${POSTGRES_HOST}
      port: \${POSTGRES_PORT}
      user: \${POSTGRES_USER}
      password: \${POSTGRES_PASSWORD}

integrations:
  github:
    - host: github.com
      token: \${GITHUB_TOKEN}

catalog:
  providers:
    github:
      organization: 'your-org'
      catalogPath: '/catalog-info.yaml'
      filters:
        branch: 'main'
        repository: '.*'

kubernetes:
  serviceLocatorMethod:
    type: 'multiTenant'
  clusterLocatorMethods:
    - type: 'config'
      clusters:
        - name: production
          url: \${K8S_PRODUCTION_URL}
          authProvider: 'serviceAccount'
          serviceAccountToken: \${K8S_PRODUCTION_TOKEN}
        - name: staging
          url: \${K8S_STAGING_URL}
          authProvider: 'serviceAccount'
          serviceAccountToken: \${K8S_STAGING_TOKEN}
EOF
```

### Service Catalog Definition
```yaml
# catalog-info.yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: library-manager
  description: Library management system with CRUD operations
  annotations:
    github.com/project-slug: your-org/library-manager
    backstage.io/kubernetes-id: library-manager
    backstage.io/techdocs-ref: dir:.
    pagerduty.com/integration-key: abc123
    grafana/dashboard-selector: 'app=library-manager'
  tags:
    - go
    - api
    - sqlite
  links:
    - url: https://library.example.com
      title: Production
      icon: web
    - url: https://staging.library.example.com
      title: Staging
      icon: dashboard
spec:
  type: service
  lifecycle: production
  owner: platform-team
  system: library-system
  dependsOn:
    - component:database-postgres
    - resource:s3-bucket
  providesApis:
    - library-api
  consumesApis:
    - auth-api

---
apiVersion: backstage.io/v1alpha1
kind: API
metadata:
  name: library-api
  description: REST API for library management
spec:
  type: openapi
  lifecycle: production
  owner: platform-team
  system: library-system
  definition: |
    openapi: 3.0.0
    info:
      title: Library Manager API
      version: 1.0.0
    paths:
      /api/books:
        get:
          summary: List all books
          responses:
            '200':
              description: Success
        post:
          summary: Create a book
          responses:
            '201':
              description: Created
```

### Software Template (Golden Path)
```yaml
# templates/go-service-template.yaml
apiVersion: scaffolder.backstage.io/v1beta3
kind: Template
metadata:
  name: go-service
  title: Go Microservice
  description: Create a new Go microservice with best practices
  tags:
    - go
    - microservice
    - recommended
spec:
  owner: platform-team
  type: service

  parameters:
    - title: Service Information
      required:
        - name
        - owner
      properties:
        name:
          title: Name
          type: string
          description: Unique name for the service
          ui:autofocus: true
          ui:help: 'Use kebab-case (e.g., my-service)'
        description:
          title: Description
          type: string
          description: A brief description of the service
        owner:
          title: Owner
          type: string
          description: Team or person who owns this service
          ui:field: OwnerPicker
          ui:options:
            allowedKinds:
              - Group
              - User

    - title: Infrastructure Configuration
      properties:
        databaseRequired:
          title: Database Required
          type: boolean
          default: true
        databaseType:
          title: Database Type
          type: string
          enum:
            - postgresql
            - mysql
            - mongodb
          default: postgresql
        replicas:
          title: Initial Replicas
          type: integer
          default: 3
          minimum: 1
          maximum: 10

  steps:
    - id: fetch-base
      name: Fetch Base Template
      action: fetch:template
      input:
        url: ./skeleton
        values:
          name: ${{ parameters.name }}
          description: ${{ parameters.description }}
          owner: ${{ parameters.owner }}

    - id: publish-github
      name: Publish to GitHub
      action: publish:github
      input:
        repoUrl: 'github.com?owner=your-org&repo=${{ parameters.name }}'
        description: ${{ parameters.description }}
        repoVisibility: private

    - id: create-argocd-app
      name: Create ArgoCD Application
      action: argocd:create-app
      input:
        name: ${{ parameters.name }}
        namespace: production
        repoUrl: 'https://github.com/your-org/${{ parameters.name }}'
        path: 'k8s/'

    - id: provision-database
      name: Provision Database
      if: ${{ parameters.databaseRequired }}
      action: crossplane:create
      input:
        apiVersion: database.platform.io/v1alpha1
        kind: PostgreSQLInstance
        metadata:
          name: ${{ parameters.name }}-db
        spec:
          parameters:
            size: small
            version: '15'

    - id: register-catalog
      name: Register in Catalog
      action: catalog:register
      input:
        repoContentsUrl: ${{ steps.publish-github.output.repoContentsUrl }}
        catalogInfoPath: '/catalog-info.yaml'

  output:
    links:
      - title: Repository
        url: ${{ steps.publish-github.output.remoteUrl }}
      - title: Open in Backstage
        icon: catalog
        entityRef: ${{ steps.register-catalog.output.entityRef }}
      - title: ArgoCD Application
        url: https://argocd.example.com/applications/${{ parameters.name }}
```

## 🚀 **Self-Service Infrastructure**

### Crossplane Compositions for IDP
```yaml
# platform/compositions/complete-app-stack.yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xapplications.platform.example.com
spec:
  group: platform.example.com
  names:
    kind: XApplication
    plural: xapplications
  claimNames:
    kind: Application
    plural: applications
  versions:
    - name: v1alpha1
      served: true
      referenceable: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                parameters:
                  type: object
                  properties:
                    name:
                      type: string
                    environment:
                      type: string
                      enum: [dev, staging, production]
                    database:
                      type: object
                      properties:
                        enabled:
                          type: boolean
                        size:
                          type: string
                          enum: [small, medium, large]
                    storage:
                      type: object
                      properties:
                        enabled:
                          type: boolean
                        sizeGB:
                          type: integer
                    observability:
                      type: object
                      properties:
                        enabled:
                          type: boolean
                        slo:
                          type: object
                          properties:
                            availability:
                              type: number
                            latencyP95:
                              type: integer

---
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: application-aws
  labels:
    provider: aws
spec:
  compositeTypeRef:
    apiVersion: platform.example.com/v1alpha1
    kind: XApplication
  
  resources:
    # Kubernetes Namespace
    - name: namespace
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha1
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: v1
              kind: Namespace
              metadata:
                name: ""  # Patched
                labels:
                  managed-by: crossplane

    # Database
    - name: rds-instance
      base:
        apiVersion: rds.aws.upbound.io/v1beta1
        kind: Instance
        spec:
          forProvider:
            region: us-east-1
            engine: postgres
            instanceClass: db.t3.micro
            allocatedStorage: 20
      patches:
        - fromFieldPath: spec.parameters.database.size
          toFieldPath: spec.forProvider.instanceClass
          transforms:
            - type: map
              map:
                small: db.t3.micro
                medium: db.t3.medium
                large: db.m5.large

    # S3 Bucket
    - name: s3-bucket
      base:
        apiVersion: s3.aws.upbound.io/v1beta1
        kind: Bucket
        spec:
          forProvider:
            region: us-east-1
            versioning:
              - enabled: true

    # ServiceMonitor for Prometheus
    - name: service-monitor
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha1
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: monitoring.coreos.com/v1
              kind: ServiceMonitor
              metadata:
                name: ""  # Patched
              spec:
                selector:
                  matchLabels:
                    app: ""  # Patched
                endpoints:
                  - port: metrics
                    interval: 30s

    # SLO Definition
    - name: slo
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha1
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: sloth.slok.dev/v1
              kind: PrometheusServiceLevel
              metadata:
                name: ""  # Patched
              spec:
                service: ""  # Patched
                slos:
                  - name: availability
                    objective: 99.9
                    sli:
                      events:
                        errorQuery: sum(rate(http_requests_total{status=~"5.."}[5m]))
                        totalQuery: sum(rate(http_requests_total[5m]))
```

## 📊 **Platform Metrics & SLOs**

### Platform Dashboards
```yaml
# platform/dashboards/platform-metrics.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: platform-dashboards
  namespace: monitoring
data:
  platform-health.json: |
    {
      "dashboard": {
        "title": "Platform Engineering Metrics",
        "panels": [
          {
            "title": "Lead Time for Changes",
            "targets": [{
              "expr": "histogram_quantile(0.95, sum(rate(deploy_duration_seconds_bucket[7d])) by (le))"
            }],
            "description": "DORA metric: Time from commit to production"
          },
          {
            "title": "Deployment Frequency",
            "targets": [{
              "expr": "sum(increase(deployments_total[7d]))"
            }],
            "description": "DORA metric: Deployments per week"
          },
          {
            "title": "Change Failure Rate",
            "targets": [{
              "expr": "sum(increase(deployments_failed_total[7d])) / sum(increase(deployments_total[7d]))"
            }],
            "description": "DORA metric: Percentage of failed deployments"
          },
          {
            "title": "Mean Time to Recovery",
            "targets": [{
              "expr": "avg(incident_duration_seconds)"
            }],
            "description": "DORA metric: Average time to resolve incidents"
          },
          {
            "title": "Developer Onboarding Time",
            "targets": [{
              "expr": "histogram_quantile(0.95, sum(rate(onboarding_duration_seconds_bucket[30d])) by (le))"
            }]
          },
          {
            "title": "Self-Service Success Rate",
            "targets": [{
              "expr": "sum(increase(selfservice_requests_success[7d])) / sum(increase(selfservice_requests_total[7d]))"
            }]
          }
        ]
      }
    }
```

### Platform SLOs
```yaml
# platform/slos/platform-slo.yaml
apiVersion: sloth.slok.dev/v1
kind: PrometheusServiceLevel
metadata:
  name: platform-slos
  namespace: monitoring
spec:
  service: "platform-engineering"
  labels:
    team: "platform"
  
  slos:
    # API Availability
    - name: "api-availability"
      objective: 99.95
      description: "Platform APIs must be available 99.95% of the time"
      sli:
        events:
          errorQuery: sum(rate(http_requests_total{job="platform-api",code=~"5.."}[5m]))
          totalQuery: sum(rate(http_requests_total{job="platform-api"}[5m]))
      alerting:
        name: PlatformAPIHighErrorRate
        labels:
          severity: critical
        annotations:
          summary: "Platform API error rate is above SLO"

    # Self-Service Provisioning Success
    - name: "provisioning-success"
      objective: 99.5
      description: "Self-service resource provisioning must succeed 99.5% of the time"
      sli:
        events:
          errorQuery: sum(rate(crossplane_managed_resource_exists{state="False"}[10m]))
          totalQuery: sum(rate(crossplane_managed_resource_exists[10m]))

    # Developer Portal Availability
    - name: "portal-availability"
      objective: 99.9
      description: "Developer portal must be available 99.9% of the time"
      sli:
        events:
          errorQuery: sum(rate(backstage_requests_total{code=~"5.."}[5m]))
          totalQuery: sum(rate(backstage_requests_total[5m]))
```

## 🎓 **Developer Onboarding Automation**

### Onboarding Workflow
```yaml
# .github/workflows/developer-onboarding.yml
name: Developer Onboarding

on:
  workflow_dispatch:
    inputs:
      developer_username:
        description: 'GitHub username'
        required: true
      developer_email:
        description: 'Email address'
        required: true
      team:
        description: 'Team assignment'
        required: true
        type: choice
        options:
          - platform
          - backend
          - frontend

jobs:
  onboard-developer:
    runs-on: ubuntu-latest
    steps:
      - name: Create Namespace
        run: |
          kubectl create namespace dev-${{ github.event.inputs.developer_username }}
          kubectl label namespace dev-${{ github.event.inputs.developer_username }} \
            owner=${{ github.event.inputs.developer_username }} \
            team=${{ github.event.inputs.team }}

      - name: Setup RBAC
        run: |
          cat <<EOF | kubectl apply -f -
          apiVersion: rbac.authorization.k8s.io/v1
          kind: RoleBinding
          metadata:
            name: ${{ github.event.inputs.developer_username }}-admin
            namespace: dev-${{ github.event.inputs.developer_username }}
          roleRef:
            apiGroup: rbac.authorization.k8s.io
            kind: ClusterRole
            name: admin
          subjects:
            - kind: User
              name: ${{ github.event.inputs.developer_username }}
              apiGroup: rbac.authorization.k8s.io
          EOF

      - name: Provision Development Database
        run: |
          kubectl apply -f - <<EOF
          apiVersion: database.platform.io/v1alpha1
          kind: PostgreSQLInstance
          metadata:
            name: ${{ github.event.inputs.developer_username }}-dev-db
            namespace: dev-${{ github.event.inputs.developer_username }}
          spec:
            parameters:
              size: small
              version: "15"
          EOF

      - name: Create Backstage User
        run: |
          # Add user to Backstage catalog
          cat <<EOF > catalog-${{ github.event.inputs.developer_username }}.yaml
          apiVersion: backstage.io/v1alpha1
          kind: User
          metadata:
            name: ${{ github.event.inputs.developer_username }}
          spec:
            profile:
              displayName: ${{ github.event.inputs.developer_username }}
              email: ${{ github.event.inputs.developer_email }}
            memberOf: [team-${{ github.event.inputs.team }}]
          EOF
          
          # Commit to catalog repo
          git add catalog-${{ github.event.inputs.developer_username }}.yaml
          git commit -m "Add user: ${{ github.event.inputs.developer_username }}"
          git push

      - name: Send Welcome Email
        uses: dawidd6/action-send-mail@v3
        with:
          server_address: smtp.gmail.com
          server_port: 465
          username: ${{ secrets.EMAIL_USERNAME }}
          password: ${{ secrets.EMAIL_PASSWORD }}
          subject: Welcome to Platform Engineering!
          to: ${{ github.event.inputs.developer_email }}
          from: Platform Team
          body: |
            Welcome to the team!
            
            Your development environment is ready:
            - Namespace: dev-${{ github.event.inputs.developer_username }}
            - Database: Provisioned and ready
            - Portal: https://backstage.example.com
            
            Getting Started:
            1. Login to Backstage
            2. Create your first service using our templates
            3. Join #platform-engineering on Slack
            
            Documentation: https://docs.platform.example.com
```

## 🎯 **Key Learnings**
- ✅ Internal Developer Platform architecture
- ✅ Self-service infrastructure provisioning
- ✅ Developer portal with Backstage
- ✅ Golden path templates
- ✅ Platform observability and SLOs
- ✅ Automated developer onboarding

## 📊 **Platform Metrics**
- **Developer Onboarding Time**: 1 hour (from 2 weeks)
- **Self-Service Success Rate**: 98%
- **Deployment Frequency**: 50+ per day
- **Platform Availability**: 99.95%
- **MTTR**: < 30 minutes

## 📚 **Additional Resources**
- [Backstage Documentation](https://backstage.io/docs/)
- [Platform Engineering Guide](https://platformengineering.org/)
- [Team Topologies](https://teamtopologies.com/)
- [DORA Metrics](https://dora.dev/)
