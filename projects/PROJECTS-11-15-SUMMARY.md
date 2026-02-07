# Projects 11-15: Advanced CI/CD & Infrastructure

This document provides a comprehensive overview of the advanced CI/CD projects (11-15), focusing on modern DevOps practices including monorepo management, reusable pipeline components, infrastructure automation, and GitOps.

---

## 📋 Quick Reference

| Project | Focus Area | Difficulty | Time Investment |
|---------|-----------|------------|-----------------|
| 11 - Monorepo CI/CD | Multi-service orchestration | ⭐⭐⭐ Advanced | 8-12 hours |
| 12 - Pipeline Library | Code reusability | ⭐⭐⭐ Advanced | 6-10 hours |
| 13 - Terraform Pipeline | Infrastructure automation | ⭐⭐⭐⭐ Expert | 10-15 hours |
| 14 - ArgoCD GitOps | Progressive delivery | ⭐⭐⭐⭐ Expert | 12-16 hours |
| 15 - Crossplane | Cloud abstraction | ⭐⭐⭐⭐ Expert | 10-14 hours |

---

## 🎯 Project 11: GitHub Actions Monorepo CI/CD

### Overview
Build intelligent CI/CD pipelines that detect changes in monorepo structures and only build/deploy affected services, optimizing build times by 60-70%.

### What You'll Learn
- Path-based change detection with `dorny/paths-filter`
- Conditional job execution strategies
- Build matrices for multiple languages
- Service dependency management
- Selective deployment workflows

### Key Technologies
- GitHub Actions
- Change detection tools
- Multi-language support (Node.js, Go, Python, Java)
- Kubernetes selective deployment
- Dependency graph management

### Real-World Use Cases
- Large organizations with multiple teams
- Microservices architectures in single repo
- Projects with shared libraries/components
- Teams using monorepo patterns (Nx, Turborepo, Lerna)

### Implementation Highlights
```yaml
# Detect changes in specific services
- uses: dorny/paths-filter@v2
  with:
    filters: |
      api-gateway:
        - 'services/api-gateway/**'
        - 'shared/**'
      user-service:
        - 'services/user-service/**'
```

### Performance Metrics
- **Before**: 20 minutes (build all services)
- **After**: 5-8 minutes (selective builds)
- **Savings**: 60-70% reduction in CI time

### Best For
- Platform teams managing multiple services
- Organizations migrating to monorepo
- Teams optimizing CI/CD costs

📖 **[Full Documentation →](./11-github-actions-monorepo/README.md)**

---

## 🎯 Project 12: Jenkins Shared Pipeline Library

### Overview
Create a centralized library of reusable Jenkins pipeline components in Groovy, eliminating code duplication and standardizing CI/CD practices across 50+ projects.

### What You'll Learn
- Jenkins Shared Library architecture
- Custom pipeline steps in Groovy
- Global variable functions
- Resource management (templates, scripts)
- Library versioning and distribution

### Key Technologies
- Jenkins
- Groovy scripting
- Git for library distribution
- Docker integration
- Kubernetes deployment automation

### Real-World Use Cases
- Organizations with many Jenkins pipelines
- Teams standardizing deployment practices
- DevOps teams creating self-service platforms
- Companies with compliance requirements

### Implementation Highlights
```groovy
// Simple pipeline using shared library
@Library('company-pipeline-library@main') _

standardPipeline(
    appName: 'my-microservice',
    buildType: 'docker',
    runSecurityScan: true,
    namespace: 'production'
)
```

### Benefits
- **Code Reuse**: 80% reduction in Jenkinsfile code
- **Consistency**: Standardized across all projects
- **Maintenance**: Single source of truth
- **Onboarding**: New teams deploy in minutes

### Library Structure
```
jenkins-pipeline-library/
├── vars/                  # Global variables (pipeline steps)
├── src/                   # Groovy classes
├── resources/             # Templates and scripts
└── test/                  # Unit tests
```

### Best For
- Platform engineering teams
- Organizations with multiple development teams
- Companies needing standardized deployments

📖 **[Full Documentation →](./12-jenkins-pipeline-library/README.md)**

---

## 🎯 Project 13: Terraform Infrastructure Pipeline

### Overview
Implement complete GitOps for infrastructure using Terraform with automated testing, security scanning, cost estimation, and drift detection.

### What You'll Learn
- Infrastructure as Code best practices
- Terraform state management
- Automated security scanning (Checkov)
- Cost estimation with Infracost
- Drift detection and remediation
- Multi-environment workflows

### Key Technologies
- Terraform 1.6+
- GitHub Actions
- Checkov (security)
- Infracost (cost estimation)
- TFLint (linting)
- AWS/GCP/Azure

### Real-World Use Cases
- Cloud infrastructure automation
- Multi-environment management
- Compliance-driven organizations
- Cost-conscious teams

### Implementation Highlights
```yaml
# Automated Terraform workflow
- name: Terraform Plan
  run: terraform plan -out=tfplan

- name: Cost Estimation
  uses: infracost/actions/setup@v2
  
- name: Security Scan
  uses: bridgecrewio/checkov-action@v12
```

### Pipeline Features
- **Security**: Checkov scans, TFLint validation
- **Cost Control**: Infracost estimates on every PR
- **Safety**: Plan preview before apply
- **Compliance**: Automated drift detection every 6 hours

### Drift Detection
- Scheduled checks every 6 hours
- Automatic GitHub issue creation
- Slack notifications
- Remediation tracking

### Best For
- Platform teams managing cloud resources
- FinOps teams controlling costs
- Security-conscious organizations

📖 **[Full Documentation →](./13-terraform-infrastructure-pipeline/README.md)**

---

## 🎯 Project 14: ArgoCD GitOps Deployment

### Overview
Build a complete GitOps deployment pipeline with ArgoCD implementing progressive delivery strategies including canary and blue-green deployments with automated rollbacks.

### What You'll Learn
- GitOps principles and practices
- ArgoCD application management
- Progressive delivery patterns
- Canary deployments with Argo Rollouts
- Blue-green deployment strategies
- Metrics-based analysis and rollbacks

### Key Technologies
- ArgoCD
- Argo Rollouts
- Kustomize
- Prometheus (metrics)
- Istio (traffic splitting)
- Kubernetes

### Real-World Use Cases
- Production Kubernetes deployments
- Teams adopting GitOps
- Organizations requiring zero-downtime deployments
- Companies with strict compliance needs

### Implementation Highlights
```yaml
# Canary deployment with analysis
strategy:
  canary:
    steps:
      - setWeight: 10
      - pause: {duration: 5m}
      - analysis:
          templates:
            - templateName: success-rate
            - templateName: latency
      - setWeight: 50
      - setWeight: 100
```

### Progressive Delivery Features
- **Canary**: Gradual traffic increase (10% → 25% → 50% → 100%)
- **Blue-Green**: Instant switch with preview environment
- **Analysis**: Prometheus metrics validation
- **Rollback**: Automatic on metrics failure

### Benefits
- **Declarative**: Git as single source of truth
- **Automated**: Self-healing and auto-sync
- **Auditable**: All changes tracked in Git
- **Secure**: RBAC and policy enforcement
- **Reliable**: Progressive rollouts with metrics

### Best For
- Production Kubernetes environments
- Teams practicing GitOps
- Organizations requiring audit trails

📖 **[Full Documentation →](./14-argocd-gitops-deployment/README.md)**

---

## 🎯 Project 15: Crossplane Infrastructure Management

### Overview
Manage cloud infrastructure using Kubernetes CRDs with Crossplane, creating a self-service platform for provisioning databases, storage, and compute resources across AWS, GCP, and Azure.

### What You'll Learn
- Kubernetes-native infrastructure management
- Composite Resource Definitions (XRDs)
- Cloud provider abstractions
- Self-service infrastructure platforms
- GitOps for infrastructure provisioning

### Key Technologies
- Crossplane
- Kubernetes CRDs
- AWS/GCP/Azure providers
- Kustomize
- GitHub Actions

### Real-World Use Cases
- Self-service infrastructure platforms
- Multi-cloud abstractions
- Developer productivity initiatives
- Standardized resource provisioning

### Implementation Highlights
```yaml
# Simple database claim
apiVersion: database.example.com/v1alpha1
kind: PostgreSQLInstance
metadata:
  name: app-database-prod
spec:
  parameters:
    size: large
    version: "15"
    highAvailability: true
```

### Composite Resources
- **Databases**: RDS, Cloud SQL, Cosmos DB
- **Storage**: S3, GCS, Azure Blob
- **Compute**: EKS, GKE, AKS clusters
- **Networking**: VPCs, Subnets, Security Groups

### Benefits
- **Unified API**: Manage all infrastructure via Kubernetes
- **Self-Service**: Developers provision independently
- **Standardization**: Consistent configurations
- **Portability**: Cloud-agnostic abstractions
- **Automation**: GitOps-driven provisioning

### Best For
- Platform engineering teams
- Multi-cloud organizations
- Teams building internal developer platforms

📖 **[Full Documentation →](./15-crossplane-infrastructure/README.md)**

---

## 🚀 Getting Started

### Recommended Learning Order

1. **Start with Project 11** if you work with monorepos
2. **Move to Project 12** if you use Jenkins extensively
3. **Complete Project 13** to learn infrastructure automation
4. **Master Project 14** for Kubernetes deployments
5. **Finish with Project 15** for advanced infrastructure patterns

### Prerequisites

- **Knowledge**: Kubernetes basics, Docker, Git
- **Tools**: kubectl, docker, terraform (varies by project)
- **Access**: Cloud account (AWS/GCP/Azure), Kubernetes cluster
- **Time**: 8-16 hours per project

### Time Commitment

- **Total Time**: 46-67 hours for all 5 projects
- **Per Project**: 8-16 hours including setup and practice
- **Recommended**: 2-3 weeks at 10 hours/week

---

## 🎯 Project Comparison Matrix

| Feature | Project 11 | Project 12 | Project 13 | Project 14 | Project 15 |
|---------|------------|------------|------------|------------|------------|
| **Platform** | GitHub Actions | Jenkins | GitHub + Terraform | ArgoCD | Kubernetes + Crossplane |
| **Focus** | Build optimization | Code reuse | Infrastructure | Deployments | Cloud resources |
| **Complexity** | Advanced | Advanced | Expert | Expert | Expert |
| **Cloud Required** | No | No | Yes | Yes | Yes |
| **K8s Required** | Yes | No | No | Yes | Yes |
| **Best For** | Monorepos | Standardization | IaC automation | GitOps | Platform engineering |

---

## 🔗 Integration Possibilities

These projects can be combined:

- **Project 11 + 14**: Monorepo with GitOps deployment
- **Project 12 + 15**: Jenkins library with Crossplane provisioning
- **Project 13 + 15**: Terraform + Crossplane hybrid infrastructure
- **Project 14 + 15**: ArgoCD deploying Crossplane resources
- **All Projects**: Complete platform engineering solution

---

## 📚 Additional Resources

### Documentation
- [Crossplane Docs](https://docs.crossplane.io/)
- [ArgoCD Docs](https://argo-cd.readthedocs.io/)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [Jenkins Shared Libraries](https://www.jenkins.io/doc/book/pipeline/shared-libraries/)

### Community
- [CNCF Slack](https://slack.cncf.io/)
- [DevOps Subreddit](https://reddit.com/r/devops)
- [Kubernetes Slack](https://kubernetes.slack.com/)

### Courses
- [Platform Engineering Path](https://www.platformengineering.org/)
- [GitOps Fundamentals](https://www.gitops.tech/)
- [Terraform Associate Certification](https://www.hashicorp.com/certification/terraform-associate)

---

## 🤝 Contributing

Found an issue or have improvements? Please contribute:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

**Ready to level up your DevOps skills? Start with any project that matches your needs!** 🚀
