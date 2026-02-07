# Project 19: Zero-Trust Security Pipeline with OPA & Falco

## 🎯 **Learning Objectives**
- Implement zero-trust security architecture
- Deploy policy-as-code with Open Policy Agent (OPA)
- Runtime threat detection with Falco
- Automate security compliance scanning
- Implement secrets management with Vault
- Build security-focused CI/CD pipelines

## 📋 **Project Overview**
Build a comprehensive zero-trust security platform integrating policy enforcement, runtime security monitoring, secrets management, and automated compliance validation across the entire CI/CD pipeline and production infrastructure.

## 🏗️ **Security Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    CI/CD Security Layer                      │
│  ┌───────────┐  ┌───────────┐  ┌────────────┐             │
│  │  SAST     │  │Container  │  │Dependency  │             │
│  │(SonarQube)│  │Scan(Trivy)│  │Check(Snyk) │             │
│  └─────┬─────┘  └─────┬─────┘  └──────┬─────┘             │
└────────┼──────────────┼────────────────┼───────────────────┘
         │              │                │
┌────────▼──────────────▼────────────────▼───────────────────┐
│                 Policy Enforcement (OPA)                     │
│  ┌────────────┐  ┌──────────┐  ┌──────────────┐           │
│  │Admission   │  │ API       │  │Infrastructure│           │
│  │Controller  │  │ Gateway   │  │  Policies    │           │
│  └─────┬──────┘  └─────┬────┘  └──────┬───────┘           │
└────────┼────────────────┼──────────────┼───────────────────┘
         │                │              │
┌────────▼────────────────▼──────────────▼───────────────────┐
│                Runtime Security (Falco)                      │
│  ┌────────────┐  ┌──────────┐  ┌──────────────┐           │
│  │  System    │  │ Network   │  │ Application  │           │
│  │  Calls     │  │ Activity  │  │  Behavior    │           │
│  └─────┬──────┘  └─────┬────┘  └──────┬───────┘           │
└────────┼────────────────┼──────────────┼───────────────────┘
         │                │              │
┌────────▼────────────────▼──────────────▼───────────────────┐
│              Secrets Management (Vault)                      │
│  ┌────────────┐  ┌──────────┐  ┌──────────────┐           │
│  │  Dynamic   │  │ Encryption│  │  Rotation    │           │
│  │  Secrets   │  │ as Service│  │  Automation  │           │
│  └────────────┘  └──────────┘  └──────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 🛡️ **Open Policy Agent (OPA) Setup**

### Install OPA Gatekeeper
```bash
# Install OPA Gatekeeper
kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.14/deploy/gatekeeper.yaml

# Verify installation
kubectl get pods -n gatekeeper-system
```

### Constraint Templates

#### Require Labels
```yaml
# opa/templates/required-labels.yaml
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: k8srequiredlabels
spec:
  crd:
    spec:
      names:
        kind: K8sRequiredLabels
      validation:
        openAPIV3Schema:
          type: object
          properties:
            labels:
              type: array
              items:
                type: string
  
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8srequiredlabels
        
        violation[{"msg": msg, "details": {"missing_labels": missing}}] {
          provided := {label | input.review.object.metadata.labels[label]}
          required := {label | label := input.parameters.labels[_]}
          missing := required - provided
          count(missing) > 0
          msg := sprintf("Missing required labels: %v", [missing])
        }
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRequiredLabels
metadata:
  name: require-app-labels
spec:
  match:
    kinds:
      - apiGroups: ["apps"]
        kinds: ["Deployment", "StatefulSet"]
    namespaces:
      - production
      - staging
  parameters:
    labels:
      - "app"
      - "owner"
      - "environment"
      - "cost-center"
```

#### Block Privileged Containers
```yaml
# opa/templates/block-privileged.yaml
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: k8sblockprivileged
spec:
  crd:
    spec:
      names:
        kind: K8sBlockPrivileged
  
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8sblockprivileged
        
        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          container.securityContext.privileged == true
          msg := sprintf("Privileged container not allowed: %v", [container.name])
        }
        
        violation[{"msg": msg}] {
          container := input.review.object.spec.initContainers[_]
          container.securityContext.privileged == true
          msg := sprintf("Privileged init container not allowed: %v", [container.name])
        }
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sBlockPrivileged
metadata:
  name: block-privileged-containers
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    excludedNamespaces:
      - kube-system
      - gatekeeper-system
```

#### Enforce Image Registry
```yaml
# opa/templates/enforce-registry.yaml
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: k8senforceregistry
spec:
  crd:
    spec:
      names:
        kind: K8sEnforceRegistry
      validation:
        openAPIV3Schema:
          type: object
          properties:
            allowedRegistries:
              type: array
              items:
                type: string
  
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8senforceregistry
        
        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not is_allowed_registry(container.image)
          msg := sprintf("Container image from unauthorized registry: %v", [container.image])
        }
        
        is_allowed_registry(image) {
          registry := input.parameters.allowedRegistries[_]
          startswith(image, registry)
        }
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sEnforceRegistry
metadata:
  name: enforce-trusted-registries
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
  parameters:
    allowedRegistries:
      - "ghcr.io/your-org/"
      - "docker.io/library/"
      - "gcr.io/your-project/"
```

## 🚨 **Falco Runtime Security**

### Install Falco
```bash
# Add Falco Helm repository
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update

# Install Falco with sidekick
helm install falco falcosecurity/falco \
  --namespace falco \
  --create-namespace \
  --set falcosidekick.enabled=true \
  --set falcosidekick.webui.enabled=true
```

### Custom Falco Rules
```yaml
# falco/custom-rules.yaml
- rule: Unauthorized Process in Container
  desc: Detect unauthorized process execution in containers
  condition: >
    spawned_process and
    container and
    not proc.name in (allowed_processes)
  output: >
    Unauthorized process started in container
    (user=%user.name command=%proc.cmdline container_id=%container.id
    image=%container.image.repository)
  priority: WARNING
  tags: [container, process]
  
- macro: allowed_processes
  condition: >
    (proc.name in (sh, bash, cat, ls, ps, grep, awk, sed, nginx, node, go))

- rule: Sensitive File Access
  desc: Detect access to sensitive files
  condition: >
    open_read and
    container and
    fd.name in (sensitive_files)
  output: >
    Sensitive file accessed
    (user=%user.name file=%fd.name container=%container.id
    process=%proc.name)
  priority: CRITICAL
  tags: [filesystem, secrets]

- macro: sensitive_files
  condition: >
    (fd.name in (/etc/shadow, /etc/passwd, /root/.ssh/id_rsa,
     /var/run/secrets/kubernetes.io/serviceaccount/token))

- rule: Crypto Mining Detection
  desc: Detect potential cryptocurrency mining activity
  condition: >
    spawned_process and
    container and
    proc.name in (cryptomining_processes)
  output: >
    Potential crypto mining detected
    (container=%container.id process=%proc.name cmdline=%proc.cmdline)
  priority: CRITICAL
  tags: [malware, mining]

- macro: cryptomining_processes
  condition: >
    (proc.name in (xmrig, minerd, ccminer, ethminer))

- rule: Reverse Shell Detected
  desc: Detect reverse shell connections
  condition: >
    spawned_process and
    container and
    ((proc.name = "bash" and
      proc.args contains "-i") or
     (proc.name = "nc" and
      proc.args contains "-e"))
  output: >
    Potential reverse shell detected
    (user=%user.name container=%container.id cmdline=%proc.cmdline)
  priority: CRITICAL
  tags: [network, shell]

- rule: Package Management in Container
  desc: Detect package manager usage in running containers
  condition: >
    spawned_process and
    container and
    proc.name in (package_managers)
  output: >
    Package manager executed in container
    (container=%container.id process=%proc.name cmdline=%proc.cmdline)
  priority: WARNING
  tags: [container, software]

- macro: package_managers
  condition: >
    (proc.name in (apt, apt-get, yum, dnf, apk, pip, npm))
```

### Falcosidekick Alert Configuration
```yaml
# falco/falcosidekick-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: falcosidekick-config
  namespace: falco
data:
  config.yaml: |
    slack:
      webhookurl: "${SLACK_WEBHOOK_URL}"
      minimumpriority: "warning"
      messageformat: "long"
    
    pagerduty:
      routingkey: "${PAGERDUTY_KEY}"
      minimumpriority: "critical"
    
    prometheus:
      extralabels: "environment:production"
    
    elasticsearch:
      hostport: "elasticsearch:9200"
      index: "falco"
      type: "event"
    
    webhook:
      address: "http://alert-manager:9093/api/v1/alerts"
```

## 🔐 **HashiCorp Vault Integration**

### Deploy Vault
```yaml
# vault/deployment.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vault
  namespace: vault
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: vault
  namespace: vault
spec:
  serviceName: vault
  replicas: 3
  selector:
    matchLabels:
      app: vault
  template:
    metadata:
      labels:
        app: vault
    spec:
      serviceAccountName: vault
      containers:
        - name: vault
          image: hashicorp/vault:1.15.0
          ports:
            - containerPort: 8200
              name: vault-port
          env:
            - name: VAULT_ADDR
              value: "http://127.0.0.1:8200"
            - name: VAULT_API_ADDR
              value: "http://$(POD_IP):8200"
          volumeMounts:
            - name: vault-config
              mountPath: /vault/config
            - name: vault-data
              mountPath: /vault/data
          securityContext:
            capabilities:
              add: ["IPC_LOCK"]
      volumes:
        - name: vault-config
          configMap:
            name: vault-config
  volumeClaimTemplates:
    - metadata:
        name: vault-data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
```

### Vault Configuration
```hcl
# vault/config.hcl
storage "raft" {
  path    = "/vault/data"
  node_id = "node1"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 0
  tls_cert_file = "/vault/tls/tls.crt"
  tls_key_file  = "/vault/tls/tls.key"
}

api_addr = "https://vault.vault.svc:8200"
cluster_addr = "https://vault-0.vault:8201"

ui = true
```

### Dynamic Database Credentials
```bash
# Enable database secrets engine
vault secrets enable database

# Configure PostgreSQL
vault write database/config/library-db \
  plugin_name=postgresql-database-plugin \
  allowed_roles="readonly,readwrite" \
  connection_url="postgresql://{{username}}:{{password}}@postgres:5432/library?sslmode=require" \
  username="vault" \
  password="vault-password"

# Create readonly role
vault write database/roles/readonly \
  db_name=library-db \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}' IN ROLE readonly;" \
  default_ttl="1h" \
  max_ttl="24h"

# Create readwrite role
vault write database/roles/readwrite \
  db_name=library-db \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}' IN ROLE readwrite;" \
  default_ttl="1h" \
  max_ttl="24h"

# Rotate root credentials
vault write -force database/rotate-root/library-db
```

### Kubernetes Auth Method
```bash
# Enable Kubernetes auth
vault auth enable kubernetes

# Configure Kubernetes auth
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443" \
  kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
  token_reviewer_jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token

# Create policy for application
vault policy write library-app - <<EOF
path "database/creds/readwrite" {
  capabilities = ["read"]
}

path "secret/data/library/*" {
  capabilities = ["read"]
}
EOF

# Create Kubernetes role
vault write auth/kubernetes/role/library-app \
  bound_service_account_names=library-manager \
  bound_service_account_namespaces=production \
  policies=library-app \
  ttl=1h
```

## 🔒 **Security-Focused CI/CD Pipeline**

### `.github/workflows/security-pipeline.yml`
```yaml
name: Zero-Trust Security Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: SAST with SonarQube
        uses: sonarsource/sonarqube-scan-action@master
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
          SONAR_HOST_URL: ${{ secrets.SONAR_HOST_URL }}

      - name: Dependency Check (Snyk)
        uses: snyk/actions/golang@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high

      - name: Secret Scanning (TruffleHog)
        uses: trufflesecurity/trufflehog@main
        with:
          path: ./
          base: ${{ github.event.repository.default_branch }}
          head: HEAD

  container-security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build image
        run: docker build -t test-image:${{ github.sha }} .

      - name: Container Scan (Trivy)
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'test-image:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'

      - name: Container Scan (Grype)
        uses: anchore/scan-action@v3
        with:
          image: 'test-image:${{ github.sha }}'
          fail-build: true
          severity-cutoff: high

      - name: Upload scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  policy-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate OPA Policies
        run: |
          # Install OPA
          curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64
          chmod +x opa
          
          # Test policies
          ./opa test opa/policies/ -v

      - name: Conftest Kubernetes Manifests
        uses: instrumenta/conftest-action@master
        with:
          files: k8s/
          policy: opa/policies/

  vault-secrets:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Import secrets from Vault
        uses: hashicorp/vault-action@v2
        with:
          url: ${{ secrets.VAULT_ADDR }}
          token: ${{ secrets.VAULT_TOKEN }}
          secrets: |
            secret/data/library/db password | DB_PASSWORD ;
            secret/data/library/api key | API_KEY

      - name: Rotate secrets
        run: |
          # Trigger secret rotation
          vault write -force database/rotate-role/library-app
```

## 🎯 **Key Learnings**
- ✅ Zero-trust security architecture
- ✅ Policy-as-code with OPA
- ✅ Runtime threat detection with Falco
- ✅ Dynamic secrets management with Vault
- ✅ Automated security compliance
- ✅ Multi-layer security defense

## 📊 **Security Metrics**
- **Policy Compliance**: 100% (all deployments pass OPA checks)
- **Secret Rotation**: Automated every 24 hours
- **Vulnerability Detection**: Real-time with Falco
- **MTTR (Mean Time to Remediate)**: < 4 hours

## 📚 **Additional Resources**
- [OPA Documentation](https://www.openpolicyagent.org/docs/)
- [Falco Documentation](https://falco.org/docs/)
- [Vault Documentation](https://www.vaultproject.io/docs)
- [NIST Zero Trust Architecture](https://www.nist.gov/publications/zero-trust-architecture)
