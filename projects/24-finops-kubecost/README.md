# Project 24: FinOps & Cloud Cost Optimization with Kubecost

## 🎯 Project Overview

Implement comprehensive FinOps practices for Kubernetes using Kubecost to track, analyze, and optimize cloud infrastructure costs. Build automated cost allocation, budget alerts, rightsizing recommendations, and multi-cluster cost visibility.

### What You'll Learn
- FinOps principles and cost optimization strategies
- Kubecost installation and configuration
- Real-time cost allocation by namespace, pod, label
- Automated rightsizing recommendations
- Cost anomaly detection and alerting
- Multi-cluster cost aggregation
- Showback and chargeback reporting
- Cloud provider cost integration (AWS, GCP, Azure)

### Technologies
- **Cost Management:** Kubecost 1.108
- **Monitoring:** Prometheus, Grafana
- **Kubernetes:** 1.28+
- **Cloud Providers:** AWS, GCP, Azure
- **Automation:** Python, GitHub Actions
- **Notifications:** Slack, Email

---

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                    FinOps Cost Management Platform                 │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                     Kubecost Architecture                      │ │
│  │                                                                 │ │
│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    │ │
│  │  │  Kubecost    │◄──►│  Prometheus  │◄──►│   Grafana    │    │ │
│  │  │   Frontend   │    │   (Metrics)  │    │  (Dashboards)│    │ │
│  │  └──────────────┘    └──────────────┘    └──────────────┘    │ │
│  │         │                     │                                │ │
│  │         │            ┌────────▼────────┐                      │ │
│  │         │            │  Cost Allocator │                      │ │
│  │         │            │   - CPU costs    │                      │ │
│  │         │            │   - Memory costs │                      │ │
│  │         │            │   - Storage costs│                      │ │
│  │         │            │   - Network costs│                      │ │
│  │         │            └────────┬────────┘                      │ │
│  │         │                     │                                │ │
│  │  ┌──────▼─────────────────────▼───────────────────┐           │ │
│  │  │         Kubernetes Clusters (Multi-Cluster)     │           │ │
│  │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐│           │ │
│  │  │  │ Cluster 1  │  │ Cluster 2  │  │ Cluster 3  ││           │ │
│  │  │  │ Production │  │  Staging   │  │    Dev     ││           │ │
│  │  │  └────────────┘  └────────────┘  └────────────┘│           │ │
│  │  └──────────────────────────────────────────────────┘           │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │              Cloud Provider Billing Integration                │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │ │
│  │  │  AWS Cost   │  │  GCP BigQuery│  │  Azure Cost │           │ │
│  │  │  & Usage    │  │   Billing   │  │  Management │           │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘           │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │            Automation & Optimization Engine                    │ │
│  │  • Rightsizing Recommendations                                 │ │
│  │  • Idle Resource Detection                                     │ │
│  │  • Cost Anomaly Alerts                                         │ │
│  │  • Budget Enforcement                                          │ │
│  │  • Automated Scaling Based on Cost                             │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                Reporting & Notifications                       │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │ │
│  │  │  Slack   │  │  Email   │  │  JIRA    │  │   CSV    │     │ │
│  │  │  Alerts  │  │  Reports │  │  Tickets │  │  Export  │     │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │ │
│  └───────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- Kubernetes cluster (1.28+)
- Helm 3.x
- Cloud provider billing access (AWS/GCP/Azure)
- Prometheus (can be installed with Kubecost)
- 8GB+ RAM available

---

## 🚀 Implementation

### Step 1: Install Kubecost

```bash
# Add Kubecost Helm repository
helm repo add kubecost https://kubecost.github.io/cost-analyzer/
helm repo update

# Create namespace
kubectl create namespace kubecost

# Install Kubecost with Prometheus
helm install kubecost kubecost/cost-analyzer \
  --namespace kubecost \
  --set kubecostToken="<your-kubecost-token>" \
  --set prometheus.server.global.external_labels.cluster_id="production-us-east-1" \
  --set prometheus.server.retention="15d" \
  --set persistentVolume.size="32Gi" \
  --set ingress.enabled=true \
  --set ingress.hosts[0]="kubecost.example.com"

# Wait for pods to be ready
kubectl wait --for=condition=Ready pods --all -n kubecost --timeout=300s

# Port-forward to access UI
kubectl port-forward -n kubecost svc/kubecost-cost-analyzer 9090:9090

# Access Kubecost at http://localhost:9090
```

### Step 2: Configure Cloud Provider Billing Integration

**AWS Cost and Usage Report Integration**

```bash
# Create S3 bucket for AWS Cost and Usage Reports
aws s3 mb s3://my-org-cur-reports --region us-east-1

# Enable Cost and Usage Report in AWS Console
# Report name: kubecost-cur
# Time granularity: Hourly
# Include resource IDs: Yes
# Enable report data integration for: Amazon Athena

# Create IAM policy for Kubecost
cat <<EOF > kubecost-aws-policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-org-cur-reports",
        "arn:aws:s3:::my-org-cur-reports/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "athena:GetQueryExecution",
        "athena:GetQueryResults",
        "athena:StartQueryExecution"
      ],
      "Resource": "*"
    }
  ]
}
EOF

aws iam create-policy \
  --policy-name KubecostCURAccess \
  --policy-document file://kubecost-aws-policy.json

# Create secret with AWS credentials
kubectl create secret generic aws-billing-creds \
  --from-literal=aws-access-key-id=$AWS_ACCESS_KEY_ID \
  --from-literal=aws-secret-access-key=$AWS_SECRET_ACCESS_KEY \
  -n kubecost
```

**kubecost-values.yaml** (AWS Configuration)
```yaml
kubecostProductConfigs:
  awsServiceKeyName: "aws-billing-creds"
  awsServiceKeyPassword: "aws-secret-access-key"
  awsServiceKeySecret: "aws-access-key-id"
  athenaProjectID: "my-org-cur-reports"
  athenaBucketName: "s3://my-org-cur-reports"
  athenaRegion: "us-east-1"
  athenaDatabase: "athenacurcfn_kubecost_cur"
  athenaTable: "kubecost_cur"

# Upgrade with new values
helm upgrade kubecost kubecost/cost-analyzer \
  --namespace kubecost \
  -f kubecost-values.yaml
```

**GCP BigQuery Integration**

```bash
# Create service account for BigQuery access
gcloud iam service-accounts create kubecost-bigquery \
  --display-name="Kubecost BigQuery Reader"

# Grant BigQuery permissions
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:kubecost-bigquery@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/bigquery.user"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:kubecost-bigquery@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/bigquery.dataViewer"

# Create and download key
gcloud iam service-accounts keys create gcp-key.json \
  --iam-account=kubecost-bigquery@PROJECT_ID.iam.gserviceaccount.com

# Create secret
kubectl create secret generic gcp-billing-creds \
  --from-file=key.json=gcp-key.json \
  -n kubecost
```

### Step 3: Configure Cost Allocation Labels

**cost-allocation-labels.yaml**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cost-allocation-labels
  namespace: kubecost
data:
  labels: |
    team
    environment
    application
    cost-center
    project
    owner
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-app
  namespace: default
  labels:
    app: sample-app
    team: engineering
    environment: production
    cost-center: "CC-1234"
    project: "customer-portal"
    owner: "john.doe@example.com"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sample-app
  template:
    metadata:
      labels:
        app: sample-app
        team: engineering
        environment: production
        cost-center: "CC-1234"
    spec:
      containers:
      - name: app
        image: nginx:alpine
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
```

### Step 4: Create Budget Alerts

**budgets/team-budget.yaml**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: team-budgets
  namespace: kubecost
data:
  budgets.json: |
    {
      "budgets": [
        {
          "name": "Engineering Team Monthly Budget",
          "filter": {
            "label": {
              "team": "engineering"
            }
          },
          "amount": 10000,
          "window": "month",
          "alerts": [
            {
              "type": "slack",
              "threshold": 0.8,
              "webhookUrl": "$SLACK_WEBHOOK_URL",
              "message": "⚠️ Engineering team is at 80% of monthly budget ($10,000)"
            },
            {
              "type": "slack",
              "threshold": 1.0,
              "webhookUrl": "$SLACK_WEBHOOK_URL",
              "message": "🚨 Engineering team has exceeded monthly budget ($10,000)"
            }
          ]
        },
        {
          "name": "Production Environment Budget",
          "filter": {
            "label": {
              "environment": "production"
            }
          },
          "amount": 25000,
          "window": "month",
          "alerts": [
            {
              "type": "email",
              "threshold": 0.9,
              "recipients": ["finance@example.com", "ops@example.com"],
              "message": "Production environment costs approaching budget limit"
            }
          ]
        }
      ]
    }
```

### Step 5: Automated Rightsizing Script

**scripts/rightsizing-automation.py**
```python
#!/usr/bin/env python3
import requests
import json
import os
from kubernetes import client, config

KUBECOST_URL = os.getenv("KUBECOST_URL", "http://kubecost-cost-analyzer.kubecost:9090")

def get_rightsizing_recommendations():
    """Fetch rightsizing recommendations from Kubecost API"""
    url = f"{KUBECOST_URL}/model/savings/requestSizing"
    params = {
        "window": "7d",
        "targetUtilization": 0.7,
        "filterClusters": "production"
    }
    
    response = requests.get(url, params=params)
    response.raise_for_status()
    return response.json()

def apply_rightsizing(recommendations):
    """Apply rightsizing recommendations to deployments"""
    config.load_incluster_config()  # or config.load_kube_config() for local
    apps_v1 = client.AppsV1Api()
    
    for rec in recommendations:
        if rec['savings'] > 50:  # Only apply if savings > $50/month
            namespace = rec['namespace']
            deployment_name = rec['controllerName']
            
            # Get current deployment
            deployment = apps_v1.read_namespaced_deployment(deployment_name, namespace)
            
            # Update resource requests
            for container in deployment.spec.template.spec.containers:
                if container.name == rec['containerName']:
                    container.resources.requests = {
                        'cpu': rec['recommendedRequest']['cpu'],
                        'memory': rec['recommendedRequest']['memory']
                    }
            
            # Apply update
            apps_v1.patch_namespaced_deployment(deployment_name, namespace, deployment)
            print(f"✅ Updated {namespace}/{deployment_name}: Estimated savings ${rec['savings']}/month")

def detect_idle_resources():
    """Detect idle resources (CPU < 5% for 7 days)"""
    url = f"{KUBECOST_URL}/model/savings/idleResources"
    params = {"window": "7d"}
    
    response = requests.get(url, params=params)
    idle_resources = response.json()
    
    for resource in idle_resources:
        if resource['type'] == 'deployment' and resource['savings'] > 10:
            print(f"⚠️ Idle resource detected: {resource['namespace']}/{resource['name']}")
            print(f"   Estimated savings: ${resource['savings']}/month")
            print(f"   Recommendation: Consider scaling down or removing")

if __name__ == "__main__":
    print("🔍 Fetching rightsizing recommendations...")
    recommendations = get_rightsizing_recommendations()
    
    print(f"\n📊 Found {len(recommendations)} optimization opportunities")
    
    # Uncomment to auto-apply recommendations
    # apply_rightsizing(recommendations)
    
    print("\n🔍 Detecting idle resources...")
    detect_idle_resources()
```

### Step 6: GitHub Actions Cost Optimization Workflow

**.github/workflows/cost-optimization.yaml**
```yaml
name: FinOps Cost Optimization

on:
  schedule:
    - cron: '0 9 * * MON'  # Every Monday at 9 AM
  workflow_dispatch:

jobs:
  analyze-costs:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG }}
      
      - name: Get Cost Summary
        id: cost_summary
        run: |
          KUBECOST_URL="http://kubecost-cost-analyzer.kubecost:9090"
          
          # Port-forward Kubecost
          kubectl port-forward -n kubecost svc/kubecost-cost-analyzer 9090:9090 &
          sleep 10
          
          # Fetch 7-day cost summary
          COST_DATA=$(curl -s "$KUBECOST_URL/model/allocation?window=7d&aggregate=namespace")
          
          # Extract top 5 expensive namespaces
          echo "## 💰 Top 5 Cost Namespaces (Last 7 Days)" > cost-report.md
          echo "$COST_DATA" | jq -r '.data[] | "\(.name): $\(.totalCost | tonumber | round)"' | head -5 >> cost-report.md
      
      - name: Check Budget Compliance
        run: |
          # Check if any team exceeded budget
          BUDGET_STATUS=$(curl -s "$KUBECOST_URL/model/budgets")
          
          EXCEEDED=$(echo "$BUDGET_STATUS" | jq '[.budgets[] | select(.percentUsed > 1)] | length')
          
          if [ "$EXCEEDED" -gt 0 ]; then
            echo "🚨 Budget Alert: $EXCEEDED team(s) exceeded budget"
            echo "$BUDGET_STATUS" | jq '.budgets[] | select(.percentUsed > 1)'
            exit 1
          fi
      
      - name: Get Rightsizing Recommendations
        run: |
          RECOMMENDATIONS=$(curl -s "$KUBECOST_URL/model/savings/requestSizing?window=7d")
          
          TOTAL_SAVINGS=$(echo "$RECOMMENDATIONS" | jq '[.[] | .savings] | add')
          
          echo "## 💡 Rightsizing Recommendations" >> cost-report.md
          echo "Potential monthly savings: \$$TOTAL_SAVINGS" >> cost-report.md
          echo "" >> cost-report.md
          echo "$RECOMMENDATIONS" | jq -r '.[] | "- \(.namespace)/\(.controllerName): $\(.savings)/mo"' | head -10 >> cost-report.md
      
      - name: Detect Idle Resources
        run: |
          IDLE=$(curl -s "$KUBECOST_URL/model/savings/idleResources?window=7d")
          
          echo "" >> cost-report.md
          echo "## ⚠️ Idle Resources" >> cost-report.md
          echo "$IDLE" | jq -r '.[] | "- \(.namespace)/\(.name) (\(.type)): $\(.savings)/mo"' >> cost-report.md
      
      - name: Post to Slack
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK }}
          payload: |
            {
              "text": "📊 Weekly FinOps Cost Report",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Weekly Cloud Cost Analysis*\n```\n$(cat cost-report.md)\n```"
                  }
                }
              ]
            }
      
      - name: Upload Cost Report
        uses: actions/upload-artifact@v4
        with:
          name: weekly-cost-report
          path: cost-report.md
```

### Step 7: Cost Anomaly Detection

**alerts/cost-anomaly-alert.yaml**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kubecost-anomaly-alerts
  namespace: kubecost
spec:
  groups:
    - name: cost-anomalies
      interval: 1h
      rules:
        - alert: UnexpectedCostSpike
          expr: |
            (
              sum(rate(node_cpu_seconds_total{mode!="idle"}[1h])) by (cluster)
              /
              sum(rate(node_cpu_seconds_total{mode!="idle"}[1h] offset 24h)) by (cluster)
            ) > 1.5
          for: 2h
          labels:
            severity: warning
            team: finops
          annotations:
            summary: "Unexpected 50% cost increase detected"
            description: "Cluster {{ $labels.cluster }} CPU usage increased by 50% compared to same time yesterday"
        
        - alert: HighDailySpend
          expr: |
            sum(kubecost_cluster_management_cost + kubecost_cluster_total_cost) > 1000
          for: 1h
          labels:
            severity: critical
          annotations:
            summary: "Daily spend exceeds $1000"
            description: "Current daily spend: ${{ $value }}"
        
        - alert: UnusedPersistentVolumes
          expr: |
            (kube_persistentvolume_status_phase{phase="Released"} == 1) 
            or 
            (kube_persistentvolume_status_phase{phase="Available"} == 1)
          for: 24h
          labels:
            severity: warning
          annotations:
            summary: "Unused PersistentVolume detected"
            description: "PV {{ $labels.persistentvolume }} has been unused for 24h"
```

### Step 8: Multi-Cluster Cost Aggregation

**multi-cluster/federated-kubecost.yaml**
```yaml
# Install Kubecost on secondary clusters with federation
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubecost-federation
  namespace: kubecost
data:
  federation.yaml: |
    federatedStorageConfigSecret: federated-store
    clusters:
      - name: production-us-east-1
        address: http://kubecost-cost-analyzer.kubecost.svc.cluster.local:9090
      - name: production-eu-west-1
        address: http://kubecost-eu.example.com:9090
      - name: staging-us-east-1
        address: http://kubecost-staging.example.com:9090
---
apiVersion: v1
kind: Secret
metadata:
  name: federated-store
  namespace: kubecost
type: Opaque
stringData:
  federated-store.yaml: |
    type: S3
    config:
      bucket: kubecost-federated-data
      region: us-east-1
      access_key_id: ${AWS_ACCESS_KEY_ID}
      secret_access_key: ${AWS_SECRET_ACCESS_KEY}
```

---

## 📊 Cost Optimization Best Practices

### 1. Rightsizing Guidelines
- Set CPU requests based on p95 usage
- Set memory requests based on p99 usage
- Review recommendations weekly
- Implement HPA for variable workloads

### 2. Reserved Instances / Savings Plans
```bash
# Analyze RI/SP coverage
curl "$KUBECOST_URL/model/savings/requestSizingV2?window=30d" | \
  jq '.recommendations[] | select(.savingsType == "reservedInstances")'
```

### 3. Idle Resource Elimination
- Delete unused PVCs
- Remove idle LoadBalancers ($20/mo each)
- Consolidate underutilized nodes
- Use spot/preemptible instances for non-critical workloads

---

## 🧪 Testing

```bash
# Verify Kubecost is collecting data
kubectl logs -n kubecost -l app=cost-analyzer --tail=100

# Test API endpoints
curl http://localhost:9090/model/allocation?window=7d

# Simulate cost spike (for testing alerts)
kubectl scale deployment sample-app --replicas=50
```

---

## 📊 Success Metrics

- **Monthly Cloud Spend Reduction:** 20-30%
- **Idle Resource Waste:** <5% of total spend
- **Budget Compliance:** 95%+ teams within budget
- **Cost Allocation Coverage:** 100% resources labeled
- **Rightsizing Adoption:** 80%+ recommendations applied

---

## 📚 Additional Resources

- [Kubecost Documentation](https://docs.kubecost.com/)
- [FinOps Foundation](https://www.finops.org/)
- [AWS Cost Optimization](https://aws.amazon.com/aws-cost-management/)
- [GCP Cost Management](https://cloud.google.com/cost-management)
