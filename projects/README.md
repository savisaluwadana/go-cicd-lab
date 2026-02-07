# 20 Advanced CI/CD Projects: Enterprise DevOps & Platform Engineering

This directory contains **20 comprehensive CI/CD projects** demonstrating real-world pipelines, infrastructure automation, service mesh, ML operations, zero-trust security, and platform engineering. Each project includes complete code, configuration files, and detailed explanations.

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

### Enterprise Architecture Projects (16-20)

16. **[Service Mesh Deployment](./16-service-mesh-deployment/)** - Istio/Linkerd, mTLS, traffic management, observability
17. **[Multi-Cloud Disaster Recovery](./17-multi-cloud-disaster-recovery/)** - AWS+GCP+Azure, automated failover, RTO/RPO compliance
18. **[MLOps with Kubeflow](./18-mlops-kubeflow-mlflow/)** - ML pipelines, model serving, A/B testing, drift detection
19. **[Zero-Trust Security](./19-zero-trust-security-opa-falco/)** - OPA policies, Falco runtime security, Vault secrets
20. **[Platform Engineering IDP](./20-platform-engineering-idp/)** - Backstage portal, self-service, golden paths, SLOs

### Advanced Multi-Platform Project (10)

10. **[Hybrid CI/CD Platform](./10-advanced-hybrid/)** - Jenkins + GitLab + GitHub Actions with Kubernetes

---

## 🎯 Learning Path

**Beginners (Weeks 1-2):** Start with Projects 1-3 (GitHub Actions basics)  
**Intermediate (Weeks 3-4):** Move to Projects 4-6 (Jenkins) and 7-9 (GitLab)  
**Advanced (Weeks 5-6):** Complete Projects 11-12 (Monorepo, Shared Libraries)  
**Expert (Weeks 7-9):** Master Projects 13-15 (Infrastructure, GitOps)  
**Architect (Weeks 10-12):** Complete Projects 16-20 (Service Mesh, DR, MLOps, Security, Platform Engineering)  
**Mastery:** Complete Project 10 (Hybrid Platform) - Combines multiple concepts

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
- **Service Mesh:** Istio, Linkerd, Envoy
- **Cloud:** AWS, GCP, Azure (Multi-cloud)
- **Infrastructure as Code:** Terraform, Crossplane, Pulumi
- **GitOps:** ArgoCD, Flux
- **Security:** OPA, Falco, Vault, Trivy, Checkov, Snyk, SAST, DAST
- **MLOps:** Kubeflow, MLflow, Seldon Core
- **Platform Engineering:** Backstage, vCluster
- **Languages:** Go, Node.js, Python, Java, Groovy
- **Tools:** Kustomize, Infracost, TFLint, Prometheus, Jaeger, Kiali
- **Monitoring:** Prometheus, Grafana, Loki, Jaeger, Kiali
- **Compliance:** DORA metrics, SLOs, Policy-as-Code

---

## 📖 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Jenkins Documentation](https://www.jenkins.io/doc/)
- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
