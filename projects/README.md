# 15 Advanced CI/CD Projects: Complete DevOps Learning Platform

This directory contains **15 comprehensive CI/CD projects** demonstrating real-world pipelines, infrastructure automation, and modern DevOps practices. Each project includes complete code, configuration files, and detailed explanations.

---

## 📋 Project Index

### GitHub Actions Projects (1-3, 11)

1. **[Basic CI Pipeline](./01-github-actions-basic/)** - Test, lint, build Go application
2. **[Docker Build & Push](./02-github-actions-docker/)** - Multi-arch builds, caching, Docker Hub/GHCR
3. **[Multi-Environment Deployment](./03-github-actions-environments/)** - Dev/Staging/Prod with approvals
11. **[Monorepo CI/CD](./11-github-actions-monorepo/)** - Selective builds, change detection, multi-service deployment

### Jenkins Projects (4-6, 12)

4. **[Declarative Pipeline](./04-jenkins-declarative/)** - Complete Jenkins pipeline with stages
5. **[Multibranch Pipeline](./05-jenkins-multibranch/)** - Branch-based builds and PR validation
6. **[Blue-Green Deployment](./06-jenkins-blue-green/)** - Zero-downtime deployments with Jenkins
12. **[Shared Pipeline Library](./12-jenkins-pipeline-library/)** - Reusable Groovy components, standardized pipelines

### GitLab CI Projects (7-9)

7. **[GitLab Auto DevOps](./07-gitlab-auto-devops/)** - Automated testing, building, and deployment
8. **[Kubernetes Deployment](./08-gitlab-kubernetes/)** - Deploy to GKE/EKS/AKS with GitLab
9. **[Security Scanning Pipeline](./09-gitlab-security/)** - SAST, DAST, dependency scanning

### Infrastructure & GitOps Projects (13-15)

13. **[Terraform Infrastructure Pipeline](./13-terraform-infrastructure-pipeline/)** - IaC automation, drift detection, cost estimation
14. **[ArgoCD GitOps Deployment](./14-argocd-gitops-deployment/)** - Progressive delivery, canary/blue-green rollouts
15. **[Crossplane Infrastructure](./15-crossplane-infrastructure/)** - Kubernetes-native cloud resource management

### Advanced Multi-Platform Project (10)

10. **[Hybrid CI/CD Platform](./10-advanced-hybrid/)** - Jenkins + GitLab + GitHub Actions with Kubernetes

---

## 🎯 Learning Path

**Beginners (Weeks 1-2):** Start with Projects 1-3 (GitHub Actions basics)  
**Intermediate (Weeks 3-4):** Move to Projects 4-6 (Jenkins) and 7-9 (GitLab)  
**Advanced (Weeks 5-6):** Complete Projects 11-12 (Monorepo, Shared Libraries)  
**Expert (Weeks 7-8):** Master Projects 13-15 (Infrastructure, GitOps)  
**Mastery:** Complete Project 10 (Hybrid Platform)

---

## 🚀 Quick Start

Each project contains:
- `README.md` - Detailed instructions
- Pipeline configuration files
- Sample application code
- Infrastructure as Code (IaC)
- Testing scripts
- Troubleshooting guide

Navigate to any project directory and follow its README.

---

## 📚 Prerequisites

- Docker installed
- Git configured
- Access to GitHub/GitLab/Jenkins (depending on project)
- kubectl (for Kubernetes projects)
- Basic understanding of YAML and shell scripting

---

## 🛠 Technologies Covered

- **CI/CD:** GitHub Actions, Jenkins, GitLab CI
- **Containers:** Docker, Docker Compose
- **Orchestration:** Kubernetes, Helm, ArgoCD, Argo Rollouts
- **Cloud:** AWS, GCP, Azure
- **Infrastructure as Code:** Terraform, Crossplane
- **GitOps:** ArgoCD, Flux
- **Security:** SAST, DAST, secret scanning, Trivy, Checkov, Snyk
- **Languages:** Go, Node.js, Python, Java
- **Tools:** Kustomize, Infracost, TFLint, Prometheus
- **Monitoring:** Prometheus, Grafana, Loki

---

## 📖 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Jenkins Documentation](https://www.jenkins.io/doc/)
- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
