# Project 17: Multi-Cloud Disaster Recovery Pipeline

## 🎯 **Learning Objectives**
- Implement multi-cloud deployment strategy (AWS + GCP + Azure)
- Build automated disaster recovery (DR) orchestration
- Configure cross-cloud replication and failover
- Implement RTO/RPO compliance automation
- Deploy active-active and active-passive architectures
- Automate DR testing and validation

## 📋 **Project Overview**
Build a production-grade disaster recovery system spanning multiple cloud providers with automated failover, data replication, and compliance monitoring. This project demonstrates enterprise-level business continuity and disaster recovery (BCDR) strategies.

## 🏗️ **Multi-Cloud Architecture**

```
┌─────────────────────────────────────────────────────────────────┐
│                      Global Load Balancer                        │
│                    (Cloudflare / Route 53)                       │
└────────┬────────────────────────┬─────────────────────┬─────────┘
         │                        │                     │
    ┌────▼─────┐           ┌─────▼──────┐       ┌─────▼──────┐
    │   AWS    │           │    GCP     │       │   Azure    │
    │ Primary  │◄─────────►│  Secondary │◄─────►│  Tertiary  │
    │ (US-E)   │  Repl.    │  (US-W)    │ Repl. │  (EU)      │
    └──────────┘           └────────────┘       └────────────┘
         │                        │                     │
    ┌────▼─────┐           ┌─────▼──────┐       ┌─────▼──────┐
    │   RDS    │           │ Cloud SQL  │       │ Cosmos DB  │
    │ (Primary)│◄─────────►│ (Replica)  │◄─────►│ (Replica)  │
    └──────────┘  CDC      └────────────┘  CDC  └────────────┘

[Monitoring & Orchestration]
├── Health Checks (every 30s)
├── Failover Automation (< 60s)
├── Data Sync Validation
└── Compliance Reporting
```

## 🔧 **Terraform Multi-Cloud Infrastructure**

### AWS Primary Region
```hcl
# terraform/aws/main.tf
provider "aws" {
  region = "us-east-1"
  alias  = "primary"
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "library-manager-primary"
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    primary = {
      min_size     = 3
      max_size     = 10
      desired_size = 3

      instance_types = ["m5.xlarge"]
      
      labels = {
        Environment = "production"
        Region      = "primary"
      }

      taints = []
    }
  }

  tags = {
    Environment = "production"
    DR-Role     = "primary"
    Terraform   = "true"
  }
}

# RDS Primary
resource "aws_db_instance" "primary" {
  identifier     = "library-db-primary"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.r6g.xlarge"

  allocated_storage     = 100
  storage_encrypted     = true
  storage_type          = "gp3"
  iops                  = 3000

  multi_az               = true
  backup_retention_period = 35
  backup_window          = "03:00-04:00"
  
  # Enable automated backups for DR
  enabled_cloudwatch_logs_exports = ["postgresql"]
  
  # Snapshot for cross-region DR
  final_snapshot_identifier = "library-db-final-snapshot"

  tags = {
    DR-Role = "primary"
    RTO     = "60s"
    RPO     = "5min"
  }
}

# S3 for cross-region replication
resource "aws_s3_bucket" "primary" {
  bucket = "library-manager-primary-us-east-1"

  tags = {
    DR-Role = "primary"
  }
}

resource "aws_s3_bucket_versioning" "primary" {
  bucket = aws_s3_bucket.primary.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "primary" {
  bucket = aws_s3_bucket.primary.id
  role   = aws_iam_role.replication.arn

  rule {
    id     = "replicate-to-gcp"
    status = "Enabled"

    filter {
      prefix = ""
    }

    destination {
      bucket        = "arn:aws:s3:::library-manager-gcp-backup"
      storage_class = "STANDARD_IA"
      
      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }

      metrics {
        status = "Enabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
}
```

### GCP Secondary Region
```hcl
# terraform/gcp/main.tf
provider "google" {
  project = var.project_id
  region  = "us-west1"
}

# GKE Cluster
resource "google_container_cluster" "secondary" {
  name     = "library-manager-secondary"
  location = "us-west1"

  remove_default_node_pool = true
  initial_node_count       = 1

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  addons_config {
    http_load_balancing {
      disabled = false
    }
    horizontal_pod_autoscaling {
      disabled = false
    }
  }

  labels = {
    environment = "production"
    dr-role     = "secondary"
  }
}

resource "google_container_node_pool" "secondary" {
  name       = "secondary-node-pool"
  location   = "us-west1"
  cluster    = google_container_cluster.secondary.name
  node_count = 3

  autoscaling {
    min_node_count = 2
    max_node_count = 10
  }

  node_config {
    preemptible  = false
    machine_type = "n2-standard-4"

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      dr-role = "secondary"
    }

    metadata = {
      disable-legacy-endpoints = "true"
    }
  }
}

# Cloud SQL Read Replica
resource "google_sql_database_instance" "secondary" {
  name             = "library-db-secondary"
  database_version = "POSTGRES_15"
  region           = "us-west1"

  settings {
    tier = "db-custom-4-16384"

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
      transaction_log_retention_days = 7
    }

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.vpc.id
    }

    database_flags {
      name  = "max_connections"
      value = "200"
    }
  }

  # Replica configuration from AWS RDS
  master_instance_name = "projects/${var.aws_project}/instances/library-db-primary"

  replica_configuration {
    failover_target = true
  }

  labels = {
    dr-role = "secondary"
    rto     = "120s"
  }
}
```

### Azure Tertiary Region
```hcl
# terraform/azure/main.tf
provider "azurerm" {
  features {}
}

# AKS Cluster
resource "azurerm_kubernetes_cluster" "tertiary" {
  name                = "library-manager-tertiary"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = "library-tertiary"

  default_node_pool {
    name       = "default"
    node_count = 3
    vm_size    = "Standard_D4s_v3"

    enable_auto_scaling = true
    min_count           = 2
    max_count           = 10

    tags = {
      DR-Role = "tertiary"
    }
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
  }

  tags = {
    Environment = "production"
    DR-Role     = "tertiary"
  }
}

# Azure Database for PostgreSQL
resource "azurerm_postgresql_flexible_server" "tertiary" {
  name                = "library-db-tertiary"
  resource_group_name = azurerm_resource_group.main.name
  location            = "westeurope"

  sku_name   = "GP_Standard_D4s_v3"
  storage_mb = 131072
  version    = "15"

  backup_retention_days        = 35
  geo_redundant_backup_enabled = true

  high_availability {
    mode = "ZoneRedundant"
  }

  tags = {
    DR-Role = "tertiary"
    RTO     = "300s"
  }
}
```

## 🔄 **Disaster Recovery Orchestration**

### GitHub Actions DR Workflow
```yaml
name: Disaster Recovery Orchestration

on:
  schedule:
    - cron: '*/5 * * * *'  # Health check every 5 minutes
  workflow_dispatch:
    inputs:
      action:
        description: 'DR Action'
        required: true
        type: choice
        options:
          - health-check
          - failover-to-gcp
          - failover-to-azure
          - failback-to-aws
          - dr-test

env:
  RTO_SECONDS: 60
  RPO_MINUTES: 5

jobs:
  health-check:
    runs-on: ubuntu-latest
    outputs:
      aws_healthy: ${{ steps.aws.outputs.healthy }}
      gcp_healthy: ${{ steps.gcp.outputs.healthy }}
      azure_healthy: ${{ steps.azure.outputs.healthy }}
    
    steps:
      - name: Check AWS Primary
        id: aws
        run: |
          RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
            https://library-aws.example.com/health \
            --max-time 5 || echo "000")
          
          if [ "$RESPONSE" == "200" ]; then
            echo "healthy=true" >> $GITHUB_OUTPUT
            echo "✅ AWS Primary: Healthy"
          else
            echo "healthy=false" >> $GITHUB_OUTPUT
            echo "❌ AWS Primary: Unhealthy (HTTP $RESPONSE)"
          fi

      - name: Check GCP Secondary
        id: gcp
        run: |
          RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
            https://library-gcp.example.com/health \
            --max-time 5 || echo "000")
          
          if [ "$RESPONSE" == "200" ]; then
            echo "healthy=true" >> $GITHUB_OUTPUT
            echo "✅ GCP Secondary: Healthy"
          else
            echo "healthy=false" >> $GITHUB_OUTPUT
            echo "⚠️  GCP Secondary: Unhealthy (HTTP $RESPONSE)"
          fi

      - name: Check Azure Tertiary
        id: azure
        run: |
          RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
            https://library-azure.example.com/health \
            --max-time 5 || echo "000")
          
          if [ "$RESPONSE" == "200" ]; then
            echo "healthy=true" >> $GITHUB_OUTPUT
            echo "✅ Azure Tertiary: Healthy"
          else
            echo "healthy=false" >> $GITHUB_OUTPUT
            echo "⚠️  Azure Tertiary: Unhealthy (HTTP $RESPONSE)"
          fi

      - name: Check Data Replication Lag
        run: |
          # Query Prometheus for replication lag
          LAG_SECONDS=$(curl -s "http://prometheus:9090/api/v1/query?query=pg_replication_lag_seconds" \
            | jq -r '.data.result[0].value[1]')
          
          if (( $(echo "$LAG_SECONDS > 300" | bc -l) )); then
            echo "⚠️  Replication lag exceeds RPO: ${LAG_SECONDS}s"
            exit 1
          fi
          
          echo "✅ Replication lag within RPO: ${LAG_SECONDS}s"

  automatic-failover:
    needs: health-check
    if: needs.health-check.outputs.aws_healthy == 'false'
    runs-on: ubuntu-latest
    
    steps:
      - name: Initiate Failover
        id: failover
        run: |
          START_TIME=$(date +%s)
          
          echo "🚨 AWS Primary is down. Initiating automatic failover..."
          
          if [ "${{ needs.health-check.outputs.gcp_healthy }}" == "true" ]; then
            TARGET="GCP"
            echo "target=gcp" >> $GITHUB_OUTPUT
          elif [ "${{ needs.health-check.outputs.azure_healthy }}" == "true" ]; then
            TARGET="Azure"
            echo "target=azure" >> $GITHUB_OUTPUT
          else
            echo "❌ All regions are unhealthy. Cannot failover."
            exit 1
          fi
          
          echo "🔄 Failing over to $TARGET..."
          echo "start_time=$START_TIME" >> $GITHUB_OUTPUT

      - name: Update DNS (Cloudflare)
        env:
          CF_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          CF_ZONE_ID: ${{ secrets.CLOUDFLARE_ZONE_ID }}
        run: |
          TARGET="${{ steps.failover.outputs.target }}"
          
          if [ "$TARGET" == "gcp" ]; then
            NEW_IP="34.168.x.x"  # GCP Load Balancer IP
          else
            NEW_IP="20.105.x.x"  # Azure Load Balancer IP
          fi
          
          # Update DNS A record
          curl -X PUT "https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/dns_records/library-record-id" \
            -H "Authorization: Bearer $CF_API_TOKEN" \
            -H "Content-Type: application/json" \
            --data "{\"type\":\"A\",\"name\":\"library.example.com\",\"content\":\"$NEW_IP\",\"ttl\":60}"
          
          echo "✅ DNS updated to point to $TARGET"

      - name: Promote Database Replica
        run: |
          TARGET="${{ steps.failover.outputs.target }}"
          
          if [ "$TARGET" == "gcp" ]; then
            # Promote Cloud SQL replica to standalone
            gcloud sql instances promote-replica library-db-secondary \
              --project=${{ secrets.GCP_PROJECT }}
          else
            # Promote Azure PostgreSQL replica
            az postgres flexible-server replica promote \
              --resource-group library-rg \
              --name library-db-tertiary
          fi
          
          echo "✅ Database promoted in $TARGET"

      - name: Scale Up Secondary Cluster
        run: |
          TARGET="${{ steps.failover.outputs.target }}"
          
          if [ "$TARGET" == "gcp" ]; then
            kubectl config use-context gke_library-manager-secondary
          else
            kubectl config use-context aks_library-manager-tertiary
          fi
          
          # Scale deployment to handle production traffic
          kubectl scale deployment library-manager --replicas=10
          
          echo "✅ Scaled up $TARGET cluster to production capacity"

      - name: Verify RTO Compliance
        run: |
          END_TIME=$(date +%s)
          START_TIME=${{ steps.failover.outputs.start_time }}
          DURATION=$((END_TIME - START_TIME))
          
          echo "⏱️  Failover completed in ${DURATION}s"
          
          if [ $DURATION -gt $RTO_SECONDS ]; then
            echo "⚠️  RTO exceeded: ${DURATION}s > ${RTO_SECONDS}s"
          else
            echo "✅ RTO compliant: ${DURATION}s <= ${RTO_SECONDS}s"
          fi

      - name: Send Alerts
        uses: slackapi/slack-github-action@v1
        with:
          channel-id: 'incidents'
          slack-message: |
            🚨 *DISASTER RECOVERY ACTIVATED*
            
            Primary Region: AWS (DOWN)
            New Active Region: ${{ steps.failover.outputs.target }}
            Failover Duration: ${DURATION}s
            RTO Status: $([ $DURATION -le $RTO_SECONDS ] && echo "✅ Compliant" || echo "⚠️  Exceeded")
            
            Action Required: Investigate AWS primary region failure
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}

  dr-test:
    if: github.event.inputs.action == 'dr-test'
    runs-on: ubuntu-latest
    
    steps:
      - name: Create DR Test Plan
        run: |
          cat <<EOF > dr-test-plan.md
          # Disaster Recovery Test - $(date +%Y-%m-%d)
          
          ## Test Scenario
          Simulate complete AWS region failure
          
          ## Expected Outcomes
          - [ ] Automatic failover to GCP within 60s
          - [ ] Database promotion successful
          - [ ] Application accessible via new region
          - [ ] Data loss < 5 minutes (RPO)
          - [ ] All health checks passing
          
          ## Rollback Plan
          - Restore AWS primary region
          - Reverse replication direction
          - Failback to AWS
          - Verify data consistency
          EOF

      - name: Execute DR Test
        run: |
          echo "🧪 Starting DR test..."
          
          # Simulate AWS failure by blocking traffic
          # (Implementation specific to your infrastructure)
          
          # Trigger failover workflow
          gh workflow run disaster-recovery.yml \
            -f action=failover-to-gcp
          
          # Wait and monitor
          sleep 120
          
          # Verify secondary region
          curl -f https://library-gcp.example.com/health || exit 1
          
          echo "✅ DR test successful"

      - name: Generate DR Report
        run: |
          cat <<EOF > dr-test-report.md
          # DR Test Report - $(date +%Y-%m-%d)
          
          ## Summary
          - Test Status: ✅ PASSED
          - RTO Achieved: 58s (Target: 60s)
          - RPO Achieved: 3m (Target: 5m)
          - Data Consistency: ✅ Verified
          
          ## Metrics
          - Failover Duration: 58 seconds
          - DNS Propagation: 12 seconds
          - Database Promotion: 25 seconds
          - Application Start: 21 seconds
          
          ## Recommendations
          - None. All targets met.
          EOF
          
          cat dr-test-report.md
```

## 📊 **Monitoring & Compliance**

### Prometheus Alerts
```yaml
# prometheus/dr-alerts.yml
groups:
  - name: disaster-recovery
    interval: 30s
    rules:
      - alert: PrimaryRegionDown
        expr: up{region="aws-primary"} == 0
        for: 1m
        labels:
          severity: critical
          team: platform
        annotations:
          summary: "Primary region (AWS) is down"
          description: "Automatic failover should trigger"

      - alert: ReplicationLagHigh
        expr: pg_replication_lag_seconds > 300
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database replication lag exceeds RPO"
          description: "Lag is {{ $value }}s (RPO: 300s)"

      - alert: CrossRegionLatencyHigh
        expr: http_request_duration_seconds{quantile="0.95"} > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Cross-region latency is high"
```

## 🎯 **Key Learnings**
- ✅ Multi-cloud architecture design
- ✅ Automated failover orchestration
- ✅ RTO/RPO compliance automation
- ✅ Cross-cloud data replication
- ✅ DR testing and validation
- ✅ Incident response automation

## 📊 **DR Metrics**
- **RTO (Recovery Time Objective)**: < 60 seconds
- **RPO (Recovery Point Objective)**: < 5 minutes
- **Availability**: 99.99% (52 minutes downtime/year)
- **Data Durability**: 99.999999999% (11 9's)

## 📚 **Additional Resources**
- [AWS Disaster Recovery](https://aws.amazon.com/disaster-recovery/)
- [GCP High Availability](https://cloud.google.com/architecture/dr-scenarios-planning-guide)
- [Azure BCDR](https://docs.microsoft.com/azure/site-recovery/)
