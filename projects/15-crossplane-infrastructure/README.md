# Project 15: Crossplane - Kubernetes-Native Infrastructure Management

## 🎯 **Learning Objectives**
- Manage cloud infrastructure using Kubernetes CRDs
- Implement infrastructure as code with Crossplane
- Create composite resources and claims
- Build self-service infrastructure platforms
- Implement GitOps for infrastructure provisioning

## 📋 **Project Overview**
Build a complete infrastructure platform using Crossplane to provision and manage cloud resources (AWS, GCP, Azure) directly from Kubernetes. Create reusable compositions for databases, storage, and compute resources with automated CI/CD pipelines.

## 🏗️ **Repository Structure**
```
crossplane-infrastructure/
├── providers/
│   ├── provider-aws.yaml
│   ├── provider-gcp.yaml
│   └── provider-azure.yaml
├── configurations/
│   ├── aws/
│   │   ├── composition-rds.yaml
│   │   ├── composition-s3.yaml
│   │   └── composition-eks.yaml
│   ├── gcp/
│   │   ├── composition-cloudsql.yaml
│   │   └── composition-gke.yaml
│   └── azure/
│       ├── composition-cosmosdb.yaml
│       └── composition-aks.yaml
├── apis/
│   ├── database-definition.yaml
│   ├── storage-definition.yaml
│   └── cluster-definition.yaml
├── claims/
│   ├── dev/
│   ├── staging/
│   └── production/
└── .github/
    └── workflows/
        ├── validate-crossplane.yml
        └── deploy-infrastructure.yml
```

## 🔧 **Provider Configuration**

### `providers/provider-aws.yaml`
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws
spec:
  package: xpkg.upbound.io/upbound/provider-aws:v0.41.0
  packagePullPolicy: IfNotPresent

---
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-credentials
      key: credentials
```

## 🗄️ **Composite Resource Definitions**

### `apis/database-definition.yaml`
```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xpostgresqlinstances.database.example.com
spec:
  group: database.example.com
  names:
    kind: XPostgreSQLInstance
    plural: xpostgresqlinstances
  claimNames:
    kind: PostgreSQLInstance
    plural: postgresqlinstances
  
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
                    size:
                      type: string
                      description: Size of the database (small, medium, large)
                      enum:
                        - small
                        - medium
                        - large
                    version:
                      type: string
                      description: PostgreSQL version
                      enum:
                        - "13"
                        - "14"
                        - "15"
                    storageGB:
                      type: integer
                      description: Storage size in GB
                      minimum: 20
                      maximum: 1000
                    highAvailability:
                      type: boolean
                      description: Enable multi-AZ deployment
                      default: false
                    backupRetentionDays:
                      type: integer
                      description: Backup retention period
                      minimum: 1
                      maximum: 35
                      default: 7
                  required:
                    - size
                    - version
                compositionSelector:
                  matchLabels:
                    provider: aws
                    type: postgresql
              required:
                - parameters
            status:
              type: object
              properties:
                endpoint:
                  type: string
                  description: Database endpoint
                port:
                  type: integer
                  description: Database port
                connectionSecret:
                  type: string
                  description: Name of the connection secret
```

## 🎨 **Composition - AWS RDS PostgreSQL**

### `configurations/aws/composition-rds.yaml`
```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: xpostgresqlinstances.aws.database.example.com
  labels:
    provider: aws
    type: postgresql
spec:
  writeConnectionSecretsToNamespace: crossplane-system
  
  compositeTypeRef:
    apiVersion: database.example.com/v1alpha1
    kind: XPostgreSQLInstance
  
  resources:
    # Security Group
    - name: securitygroup
      base:
        apiVersion: ec2.aws.upbound.io/v1beta1
        kind: SecurityGroup
        spec:
          forProvider:
            region: us-east-1
            description: Security group for PostgreSQL RDS instance
            vpcIdSelector:
              matchControllerRef: true
      
      patches:
        - fromFieldPath: metadata.uid
          toFieldPath: spec.forProvider.name
          transforms:
            - type: string
              string:
                fmt: "sg-postgresql-%s"
    
    # Security Group Rule - Allow PostgreSQL
    - name: securitygrouprule
      base:
        apiVersion: ec2.aws.upbound.io/v1beta1
        kind: SecurityGroupRule
        spec:
          forProvider:
            region: us-east-1
            type: ingress
            fromPort: 5432
            toPort: 5432
            protocol: tcp
            cidrBlocks:
              - 10.0.0.0/8
            securityGroupIdSelector:
              matchControllerRef: true
    
    # DB Subnet Group
    - name: dbsubnetgroup
      base:
        apiVersion: rds.aws.upbound.io/v1beta1
        kind: SubnetGroup
        spec:
          forProvider:
            region: us-east-1
            description: Subnet group for PostgreSQL RDS
            subnetIdSelector:
              matchLabels:
                type: private
      
      patches:
        - fromFieldPath: metadata.uid
          toFieldPath: spec.forProvider.name
          transforms:
            - type: string
              string:
                fmt: "subnet-group-%s"
    
    # RDS Instance
    - name: rdsinstance
      base:
        apiVersion: rds.aws.upbound.io/v1beta1
        kind: Instance
        spec:
          forProvider:
            region: us-east-1
            engine: postgres
            instanceClass: db.t3.micro
            allocatedStorage: 20
            storageType: gp3
            storageEncrypted: true
            skipFinalSnapshot: true
            publiclyAccessible: false
            dbSubnetGroupNameSelector:
              matchControllerRef: true
            vpcSecurityGroupIdSelector:
              matchControllerRef: true
          
          writeConnectionSecretToRef:
            namespace: crossplane-system
      
      patches:
        # Size mapping
        - type: map
          fromFieldPath: spec.parameters.size
          toFieldPath: spec.forProvider.instanceClass
          transforms:
            - type: map
              map:
                small: db.t3.micro
                medium: db.t3.medium
                large: db.m5.large
        
        # Storage size
        - fromFieldPath: spec.parameters.storageGB
          toFieldPath: spec.forProvider.allocatedStorage
        
        # PostgreSQL version
        - fromFieldPath: spec.parameters.version
          toFieldPath: spec.forProvider.engineVersion
        
        # High availability
        - fromFieldPath: spec.parameters.highAvailability
          toFieldPath: spec.forProvider.multiAz
        
        # Backup retention
        - fromFieldPath: spec.parameters.backupRetentionDays
          toFieldPath: spec.forProvider.backupRetentionPeriod
        
        # Database name
        - fromFieldPath: metadata.uid
          toFieldPath: spec.forProvider.dbName
          transforms:
            - type: string
              string:
                type: Format
                fmt: "db%s"
        
        # Master username
        - type: string
          string:
            fmt: "admin"
          toFieldPath: spec.forProvider.username
        
        # Connection secret name
        - fromFieldPath: metadata.uid
          toFieldPath: spec.writeConnectionSecretToRef.name
          transforms:
            - type: string
              string:
                fmt: "postgresql-conn-%s"
      
      connectionDetails:
        - name: username
          fromConnectionSecretKey: username
        - name: password
          fromConnectionSecretKey: password
        - name: endpoint
          fromConnectionSecretKey: endpoint
        - name: port
          fromConnectionSecretKey: port
```

## 📝 **Resource Claims**

### `claims/production/database-claim.yaml`
```yaml
apiVersion: database.example.com/v1alpha1
kind: PostgreSQLInstance
metadata:
  name: app-database-prod
  namespace: production
  labels:
    app: library-manager
    environment: production
spec:
  parameters:
    size: large
    version: "15"
    storageGB: 100
    highAvailability: true
    backupRetentionDays: 30
  
  compositionSelector:
    matchLabels:
      provider: aws
      type: postgresql
  
  writeConnectionSecretToRef:
    name: app-database-credentials
```

### `claims/dev/database-claim.yaml`
```yaml
apiVersion: database.example.com/v1alpha1
kind: PostgreSQLInstance
metadata:
  name: app-database-dev
  namespace: dev
spec:
  parameters:
    size: small
    version: "15"
    storageGB: 20
    highAvailability: false
    backupRetentionDays: 7
  
  compositionSelector:
    matchLabels:
      provider: aws
      type: postgresql
  
  writeConnectionSecretToRef:
    name: app-database-credentials
```

## 🚀 **CI/CD Pipeline**

### `.github/workflows/deploy-infrastructure.yml`
```yaml
name: Deploy Infrastructure with Crossplane

on:
  push:
    branches: [main]
    paths:
      - 'claims/**'
      - 'configurations/**'
  pull_request:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'v1.28.0'

      - name: Validate YAML
        run: |
          for file in $(find . -name '*.yaml'); do
            echo "Validating $file"
            kubectl --dry-run=client apply -f "$file" || exit 1
          done

      - name: Install Crossplane CLI
        run: |
          curl -sL "https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh" | sh
          sudo mv kubectl-crossplane /usr/local/bin

      - name: Validate Compositions
        run: |
          for comp in configurations/**/*.yaml; do
            echo "Validating composition: $comp"
            kubectl crossplane beta validate "$comp"
          done

  deploy-dev:
    needs: validate
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: development
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        env:
          KUBECONFIG_DATA: ${{ secrets.KUBECONFIG_DEV }}
        run: |
          mkdir -p ~/.kube
          echo "$KUBECONFIG_DATA" | base64 -d > ~/.kube/config

      - name: Apply claims
        run: |
          kubectl apply -f claims/dev/

      - name: Wait for resources
        run: |
          kubectl wait --for=condition=Ready \
            postgresqlinstance/app-database-dev \
            -n dev \
            --timeout=10m

      - name: Get connection details
        run: |
          kubectl get secret app-database-credentials \
            -n dev \
            -o jsonpath='{.data.endpoint}' | base64 -d
          echo "Database endpoint retrieved successfully"

  deploy-production:
    needs: [validate, deploy-dev]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        env:
          KUBECONFIG_DATA: ${{ secrets.KUBECONFIG_PROD }}
        run: |
          mkdir -p ~/.kube
          echo "$KUBECONFIG_DATA" | base64 -d > ~/.kube/config

      - name: Apply claims
        run: |
          kubectl apply -f claims/production/

      - name: Wait for resources
        run: |
          kubectl wait --for=condition=Ready \
            postgresqlinstance/app-database-prod \
            -n production \
            --timeout=15m

      - name: Notify deployment
        uses: slackapi/slack-github-action@v1
        with:
          channel-id: 'infrastructure'
          slack-message: '✅ Production database provisioned via Crossplane'
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

## 🎯 **Using the Provisioned Database**

### Application Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: library-manager
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: library-manager
  template:
    metadata:
      labels:
        app: library-manager
    spec:
      containers:
        - name: app
          image: library-manager:latest
          env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: app-database-credentials
                  key: endpoint
            - name: DB_PORT
              valueFrom:
                secretKeyRef:
                  name: app-database-credentials
                  key: port
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: app-database-credentials
                  key: username
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-database-credentials
                  key: password
```

## 🔧 **Crossplane CLI Commands**
```bash
# Install Crossplane
kubectl create namespace crossplane-system
helm install crossplane \
  --namespace crossplane-system \
  crossplane-stable/crossplane

# Install provider
kubectl apply -f providers/provider-aws.yaml

# Create AWS credentials secret
kubectl create secret generic aws-credentials \
  -n crossplane-system \
  --from-file=credentials=./aws-credentials.txt

# Apply compositions
kubectl apply -f configurations/aws/

# Create claim
kubectl apply -f claims/dev/database-claim.yaml

# Check status
kubectl get postgresqlinstance -n dev
kubectl describe postgresqlinstance app-database-dev -n dev

# Get connection secret
kubectl get secret app-database-credentials -n dev -o yaml
```

## 🎯 **Key Learnings**
- ✅ Kubernetes-native infrastructure management
- ✅ Self-service infrastructure platform
- ✅ Reusable compositions
- ✅ GitOps for infrastructure
- ✅ Multi-cloud abstractions
- ✅ Policy-driven provisioning

## 📊 **Benefits**
- **Unified API**: Manage all infrastructure via Kubernetes
- **Self-Service**: Developers provision resources independently
- **Standardization**: Consistent configurations across teams
- **Portability**: Cloud-agnostic abstractions
- **Automation**: GitOps-driven provisioning

## 📚 **Additional Resources**
- [Crossplane Documentation](https://docs.crossplane.io/)
- [Provider AWS](https://marketplace.upbound.io/providers/upbound/provider-aws)
- [Composition Guide](https://docs.crossplane.io/latest/concepts/compositions/)
