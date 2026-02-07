# go-cicd-lab — Library Manager & CI/CD Learning Platform

![CI Status](https://github.com/savisaluwadana/go-cicd-lab/workflows/CI%20Pipeline/badge.svg)

A comprehensive learning platform combining a practical Library Manager application with **10 real-world CI/CD projects** covering GitHub Actions, Jenkins, and GitLab CI.

---

## 🎯 What's Inside

### 📚 Sample Application: Library Manager
A full-stack CRUD application demonstrating:
- **Backend:** Go REST API with SQLite persistence
- **Frontend:** HTML/CSS/JavaScript
- **Containerization:** Docker + Docker Compose
- **CI/CD:** GitHub Actions pipeline
- **Tests:** Unit and integration tests

### 🚀 10 CI/CD Projects

Explore comprehensive, production-ready CI/CD implementations:

#### GitHub Actions (Projects 1-3)
1. **[Basic CI Pipeline](./projects/01-github-actions-basic/)** - Testing, linting, building
2. **[Docker Build & Push](./projects/02-github-actions-docker/)** - Multi-arch images, caching
3. **[Multi-Environment Deployment](./projects/03-github-actions-environments/)** - Dev/Staging/Prod with approvals

#### Jenkins (Projects 4-6)
4. **[Declarative Pipeline](./projects/04-jenkins-declarative/)** - Complete pipeline with stages
5. **[Multibranch Pipeline](./projects/05-jenkins-multibranch/)** - Branch-based builds
6. **[Blue-Green Deployment](./projects/06-jenkins-blue-green/)** - Zero-downtime deployments

#### GitLab CI (Projects 7-9)
7. **[GitLab Auto DevOps](./projects/07-gitlab-auto-devops/)** - Automated testing & deployment
8. **[Kubernetes Deployment](./projects/08-gitlab-kubernetes/)** - Deploy to K8s clusters
9. **[Security Scanning](./projects/09-gitlab-security/)** - SAST, DAST, container scanning

#### Advanced (Project 10)
10. **[Hybrid CI/CD Platform](./projects/10-advanced-hybrid/)** - Multi-platform orchestration with Kubernetes

📖 **[View All Projects →](./projects/README.md)**

---

## 🏃 Quick Start

### Run the Library Manager App

**Option 1: Local Development**
```bash
go get modernc.org/sqlite@latest
go mod tidy
go run .
# Open http://localhost:8080
```

**Option 2: Docker**
```bash
docker build -t library-manager:local .
docker run -p 8080:8080 library-manager:local
```

**Option 3: Docker Compose** (Recommended)
```bash
docker compose build
docker compose up -d

# Open http://localhost:8080
# Inspect DB:
docker compose run --rm sqlite sqlite3 /data/data.db "SELECT * FROM books;"
```

### Explore CI/CD Projects

```bash
cd projects/
# Read the main projects README
cat README.md

# Start with GitHub Actions basics
cd 01-github-actions-basic/
cat README.md
```

---

## 📖 API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/books` | List all books |
| GET | `/api/book?id={id}` | Get a specific book |
| POST | `/api/books` | Create a new book |
| PUT | `/api/books` | Update an existing book |
| DELETE | `/api/book?id={id}` | Delete a book |

**Example Request:**
```bash
# Create a book
curl -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{"id":"3","title":"DevOps Handbook","author":"Gene Kim","year":2016}'

# List all books
curl http://localhost:8080/api/books
```

---

## 🛠 Technologies & Tools

### Application Stack
- **Language:** Go 1.19
- **Database:** SQLite (modernc.org/sqlite)
- **Frontend:** Vanilla HTML/CSS/JavaScript
- **Containerization:** Docker, Docker Compose

### CI/CD Platforms Covered
- **GitHub Actions** - Cloud-native CI/CD
- **Jenkins** - Self-hosted automation server
- **GitLab CI** - Integrated DevOps platform

### Deployment Targets
- **Kubernetes** - Container orchestration
- **Docker** - Containerization
- **Cloud Platforms** - AWS, GCP, Azure

### Additional Tools
- ArgoCD (GitOps)
- Helm (Package management)
- Prometheus/Grafana (Monitoring)
- Trivy (Security scanning)
- Terraform (Infrastructure as Code)

---

## 📚 Learning Path

### Beginner (Week 1-2)
1. Run the Library Manager app locally
2. Complete Project 1: Basic CI Pipeline
3. Complete Project 2: Docker Build & Push
4. Understand GitHub Actions workflows

### Intermediate (Week 3-4)
5. Complete Project 3: Multi-Environment Deployment
6. Set up Jenkins locally
7. Complete Projects 4-5: Jenkins Pipelines
8. Learn Kubernetes basics

### Advanced (Week 5-8)
9. Complete Projects 6-9: Advanced deployments & security
10. Complete Project 10: Hybrid CI/CD Platform
11. Implement GitOps with ArgoCD
12. Add monitoring and observability

---

## 🔒 Security Best Practices

All projects demonstrate:
- ✅ Secret management with environment variables
- ✅ Container image scanning
- ✅ Dependency vulnerability checks
- ✅ SAST/DAST integration
- ✅ Least privilege access
- ✅ Non-root container users

---

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## 📝 Additional Resources

- [Kubernetes Commands Cheatsheet](./kubernetescommands.md)
- [Projects Overview](./projects/README.md)
- [Projects 5-10 Summary](./projects/PROJECTS-5-10-SUMMARY.md)

---

## 📄 License

MIT License - feel free to use this for learning and teaching!

---

## 🙋 Questions?

- Open an issue for bugs or feature requests
- Check existing projects for examples
- Review the detailed README in each project folder

**Happy Learning! 🚀**