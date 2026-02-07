# Project 21: Chaos Engineering with Litmus & Chaos Mesh

## 🎯 Project Overview

Build a comprehensive chaos engineering platform to proactively test system resilience, fault tolerance, and recovery capabilities. This project implements automated chaos experiments using Litmus Chaos and Chaos Mesh to validate SLOs, uncover weaknesses, and build confidence in production systems.

### What You'll Learn
- Chaos engineering principles and methodologies
- Litmus Chaos workflows and experiments
- Chaos Mesh fault injection techniques
- Automated chaos testing in CI/CD
- Observability during chaos experiments
- GameDays and failure injection strategies
- Building resilient microservices architectures

### Technologies
- **Chaos Engineering:** Litmus 3.0, Chaos Mesh 2.6
- **Orchestration:** Kubernetes 1.28+
- **Monitoring:** Prometheus, Grafana, Jaeger
- **CI/CD:** GitHub Actions, ArgoCD
- **Service Mesh:** Istio (optional)
- **Notification:** Slack, PagerDuty

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Chaos Engineering Platform                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐  │
│  │   Litmus     │      │ Chaos Mesh   │      │   Steady     │  │
│  │   Chaos      │◄────►│   Control    │◄────►│   State      │  │
│  │   Center     │      │   Manager    │      │   Validator  │  │
│  └──────────────┘      └──────────────┘      └──────────────┘  │
│         │                      │                      │         │
│         │                      │                      │         │
│  ┌──────▼──────────────────────▼──────────────────────▼──────┐ │
│  │              Kubernetes Cluster (Target)                   │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐      │ │
│  │  │ Service │  │ Service │  │ Service │  │ Service │      │ │
│  │  │    A    │  │    B    │  │    C    │  │    D    │      │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘      │ │
│  │                                                             │ │
│  │  Fault Injection:                                          │ │
│  │  • Pod Failures    • Network Latency   • CPU Stress       │ │
│  │  • Disk Failures   • Memory Pressure   • DNS Errors       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              │                                  │
│  ┌───────────────────────────▼──────────────────────────────┐  │
│  │         Observability & Alerting Stack                    │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │  │
│  │  │ Prometheus │  │  Grafana   │  │   Jaeger   │         │  │
│  │  └────────────┘  └────────────┘  └────────────┘         │  │
│  │                                                            │  │
│  │  ┌────────────┐  ┌────────────┐                          │  │
│  │  │   Slack    │  │ PagerDuty  │                          │  │
│  │  └────────────┘  └────────────┘                          │  │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- Kubernetes cluster (1.28+)
- kubectl configured
- Helm 3.x
- Docker
- 8GB+ RAM available

---

## 🚀 Implementation

### Step 1: Install Litmus Chaos

```bash
# Add Litmus Helm repository
helm repo add litmuschaos https://litmuschaos.github.io/litmus-helm/
helm repo update

# Install Litmus Chaos Center
kubectl create namespace litmus
helm install chaos litmuschaos/litmus \
  --namespace litmus \
  --set portal.server.replicas=1 \
  --set mongodb.persistence.size=10Gi

# Wait for pods to be ready
kubectl wait --for=condition=Ready pods --all -n litmus --timeout=300s

# Get Litmus portal URL
kubectl get svc -n litmus chaos-litmus-frontend-service
```

### Step 2: Install Chaos Mesh

```bash
# Install Chaos Mesh using Helm
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update

kubectl create namespace chaos-mesh
helm install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-mesh \
  --set chaosDaemon.runtime=containerd \
  --set chaosDaemon.socketPath=/run/containerd/containerd.sock \
  --set dashboard.create=true

# Access Chaos Mesh Dashboard
kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333
```

### Step 3: Deploy Sample Microservices Application

**deployment.yaml**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: chaos-demo
  labels:
    chaos: "enabled"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: chaos-demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: chaos-demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: hashicorp/http-echo:latest
        args:
          - "-text=Backend Service v1.0"
        ports:
        - containerPort: 5678
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: chaos-demo
spec:
  selector:
    app: frontend
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: chaos-demo
spec:
  selector:
    app: backend
  ports:
  - port: 5678
    targetPort: 5678
```

### Step 4: Create Litmus Chaos Experiments

**litmus-pod-delete.yaml**
```yaml
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: frontend-chaos
  namespace: chaos-demo
spec:
  appinfo:
    appns: chaos-demo
    applabel: 'app=frontend'
    appkind: deployment
  engineState: active
  chaosServiceAccount: litmus-admin
  experiments:
    - name: pod-delete
      spec:
        components:
          env:
            - name: TOTAL_CHAOS_DURATION
              value: '60'
            - name: CHAOS_INTERVAL
              value: '10'
            - name: FORCE
              value: 'false'
            - name: PODS_AFFECTED_PERC
              value: '50'
        probe:
          - name: check-frontend-availability
            type: httpProbe
            mode: Continuous
            runProperties:
              probeTimeout: 5
              interval: 2
              retry: 3
            httpProbe/inputs:
              url: http://frontend.chaos-demo.svc.cluster.local
              method:
                get:
                  criteria: ==
                  responseCode: "200"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: litmus-admin
  namespace: chaos-demo
  labels:
    app.kubernetes.io/name: litmus
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: litmus-admin
  namespace: chaos-demo
rules:
  - apiGroups: [""]
    resources: ["pods", "events", "configmaps", "secrets", "pods/log", "pods/exec", "serviceaccounts"]
    verbs: ["create", "list", "get", "patch", "update", "delete", "deletecollection"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets", "daemonsets"]
    verbs: ["list", "get", "patch", "update"]
  - apiGroups: ["litmuschaos.io"]
    resources: ["chaosengines", "chaosexperiments", "chaosresults"]
    verbs: ["create", "list", "get", "patch", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: litmus-admin
  namespace: chaos-demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: litmus-admin
subjects:
  - kind: ServiceAccount
    name: litmus-admin
    namespace: chaos-demo
```

**litmus-network-latency.yaml**
```yaml
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: backend-network-chaos
  namespace: chaos-demo
spec:
  appinfo:
    appns: chaos-demo
    applabel: 'app=backend'
    appkind: deployment
  engineState: active
  chaosServiceAccount: litmus-admin
  experiments:
    - name: pod-network-latency
      spec:
        components:
          env:
            - name: NETWORK_INTERFACE
              value: 'eth0'
            - name: NETWORK_LATENCY
              value: '2000'  # 2 seconds
            - name: TOTAL_CHAOS_DURATION
              value: '120'
            - name: PODS_AFFECTED_PERC
              value: '33'
            - name: JITTER
              value: '100'
        probe:
          - name: check-backend-latency
            type: cmdProbe
            mode: Edge
            runProperties:
              probeTimeout: 10
              interval: 5
              retry: 3
            cmdProbe/inputs:
              command: |
                curl -o /dev/null -s -w '%{time_total}\n' http://backend.chaos-demo.svc.cluster.local:5678
              comparator:
                type: float
                criteria: <=
                value: "5.0"
```

### Step 5: Chaos Mesh Experiments

**chaos-mesh-pod-kill.yaml**
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pod-kill-frontend
  namespace: chaos-demo
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - chaos-demo
    labelSelectors:
      app: frontend
  duration: '60s'
  scheduler:
    cron: '@every 5m'
```

**chaos-mesh-network-partition.yaml**
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-partition
  namespace: chaos-demo
spec:
  action: partition
  mode: all
  selector:
    namespaces:
      - chaos-demo
    labelSelectors:
      app: backend
  direction: to
  target:
    mode: all
    selector:
      namespaces:
        - chaos-demo
      labelSelectors:
        app: frontend
  duration: '30s'
  scheduler:
    cron: '@every 10m'
```

**chaos-mesh-stress-cpu.yaml**
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: stress-cpu-backend
  namespace: chaos-demo
spec:
  mode: one
  selector:
    namespaces:
      - chaos-demo
    labelSelectors:
      app: backend
  stressors:
    cpu:
      workers: 2
      load: 80
  duration: '2m'
```

**chaos-mesh-io-latency.yaml**
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: io-delay
  namespace: chaos-demo
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - chaos-demo
    labelSelectors:
      app: backend
  volumePath: /tmp
  path: '*'
  delay: '100ms'
  percent: 50
  duration: '1m'
```

### Step 6: Chaos Workflow with Litmus

**chaos-workflow.yaml**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: comprehensive-chaos-workflow
  namespace: litmus
spec:
  entrypoint: chaos-pipeline
  serviceAccountName: argo-chaos
  arguments:
    parameters:
      - name: adminModeNamespace
        value: litmus
  templates:
    - name: chaos-pipeline
      steps:
        - - name: install-experiments
            template: install-chaos-experiments
        - - name: pod-delete
            template: pod-delete-experiment
          - name: network-latency
            template: network-latency-experiment
          - name: cpu-stress
            template: cpu-stress-experiment
        - - name: revert-chaos
            template: revert
    
    - name: install-chaos-experiments
      inputs:
        artifacts:
          - name: install-chaos-experiments
            path: /tmp/chaosengine.yaml
            raw:
              data: |
                apiVersion: litmuschaos.io/v1alpha1
                kind: ChaosExperiment
                metadata:
                  name: pod-delete
                  namespace: chaos-demo
                spec:
                  definition:
                    scope: Namespaced
                    permissions:
                      - apiGroups: [""]
                        resources: ["pods"]
                        verbs: ["create","delete","get","list","patch","update", "deletecollection"]
      container:
        image: litmuschaos/k8s:latest
        command: [sh, -c]
        args: ['kubectl apply -f /tmp/chaosengine.yaml']
    
    - name: pod-delete-experiment
      container:
        image: litmuschaos/k8s:latest
        command: [sh, -c]
        args: ['kubectl apply -f /tmp/pod-delete-engine.yaml && sleep 90']
    
    - name: network-latency-experiment
      container:
        image: litmuschaos/k8s:latest
        command: [sh, -c]
        args: ['kubectl apply -f /tmp/network-latency-engine.yaml && sleep 150']
    
    - name: cpu-stress-experiment
      container:
        image: litmuschaos/k8s:latest
        command: [sh, -c]
        args: ['kubectl apply -f /tmp/cpu-stress-engine.yaml && sleep 120']
    
    - name: revert
      container:
        image: litmuschaos/k8s:latest
        command: [sh, -c]
        args: ['kubectl delete chaosengine --all -n chaos-demo']
```

### Step 7: Automated Chaos in CI/CD Pipeline

**.github/workflows/chaos-testing.yaml**
```yaml
name: Chaos Engineering Tests

on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Mondays at 2 AM
  workflow_dispatch:
    inputs:
      experiment:
        description: 'Chaos experiment to run'
        required: true
        type: choice
        options:
          - all
          - pod-delete
          - network-latency
          - cpu-stress
          - io-chaos

jobs:
  chaos-test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG }}
      
      - name: Install Chaos Mesh CLI
        run: |
          curl -sSL https://mirrors.chaos-mesh.org/latest/install.sh | bash
      
      - name: Run Pre-Chaos Health Checks
        run: |
          kubectl get pods -n chaos-demo
          kubectl top nodes
          kubectl top pods -n chaos-demo
      
      - name: Execute Chaos Experiment - Pod Delete
        if: github.event.inputs.experiment == 'pod-delete' || github.event.inputs.experiment == 'all'
        run: |
          kubectl apply -f chaos-experiments/litmus-pod-delete.yaml
          sleep 90
          kubectl get chaosresults -n chaos-demo
      
      - name: Execute Chaos Experiment - Network Latency
        if: github.event.inputs.experiment == 'network-latency' || github.event.inputs.experiment == 'all'
        run: |
          kubectl apply -f chaos-experiments/chaos-mesh-network-partition.yaml
          sleep 60
          kubectl get networkchaos -n chaos-demo
      
      - name: Execute Chaos Experiment - CPU Stress
        if: github.event.inputs.experiment == 'cpu-stress' || github.event.inputs.experiment == 'all'
        run: |
          kubectl apply -f chaos-experiments/chaos-mesh-stress-cpu.yaml
          sleep 120
          kubectl get stresschaos -n chaos-demo
      
      - name: Collect Chaos Results
        if: always()
        run: |
          kubectl get chaosresults -n chaos-demo -o yaml > chaos-results.yaml
          kubectl logs -n chaos-demo -l app=frontend --tail=100 > frontend-logs.txt
          kubectl logs -n chaos-demo -l app=backend --tail=100 > backend-logs.txt
      
      - name: Validate SLOs
        run: |
          # Check if services recovered within SLO (99.9% uptime = 43s downtime/month)
          # Check Prometheus metrics for availability
          curl -s "http://prometheus:9090/api/v1/query?query=up{job='chaos-demo'}" | jq '.data.result[0].value[1]'
      
      - name: Cleanup Chaos Resources
        if: always()
        run: |
          kubectl delete chaosengine --all -n chaos-demo
          kubectl delete podchaos --all -n chaos-demo
          kubectl delete networkchaos --all -n chaos-demo
          kubectl delete stresschaos --all -n chaos-demo
      
      - name: Post-Chaos Health Checks
        if: always()
        run: |
          kubectl get pods -n chaos-demo
          kubectl get events -n chaos-demo --sort-by='.lastTimestamp'
      
      - name: Upload Chaos Reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: chaos-reports
          path: |
            chaos-results.yaml
            frontend-logs.txt
            backend-logs.txt
      
      - name: Notify Slack
        if: failure()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK }}
          payload: |
            {
              "text": "🔥 Chaos Engineering Test Failed",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Chaos Test Failed* ⚠️\n*Experiment:* ${{ github.event.inputs.experiment }}\n*Workflow:* ${{ github.workflow }}\n<${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|View Results>"
                  }
                }
              ]
            }
```

### Step 8: Monitoring and Observability

**prometheus-rules.yaml**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: chaos-alerts
  namespace: monitoring
spec:
  groups:
    - name: chaos-engineering
      interval: 30s
      rules:
        - alert: ChaosExperimentFailed
          expr: litmuschaos_experiment_verdict{verdict="Fail"} > 0
          for: 1m
          labels:
            severity: critical
            team: platform
          annotations:
            summary: "Chaos experiment {{ $labels.experiment }} failed"
            description: "The chaos experiment {{ $labels.experiment }} in namespace {{ $labels.namespace }} has failed. System may not be resilient to this failure scenario."
        
        - alert: HighPodRestartRate
          expr: rate(kube_pod_container_status_restarts_total{namespace="chaos-demo"}[5m]) > 0.5
          for: 2m
          labels:
            severity: warning
          annotations:
            summary: "High pod restart rate in chaos-demo namespace"
            description: "Pod {{ $labels.pod }} is restarting frequently ({{ $value }} restarts/min)"
        
        - alert: ServiceUnavailableDuringChaos
          expr: up{job="chaos-demo"} == 0
          for: 1m
          labels:
            severity: warning
          annotations:
            summary: "Service unavailable during chaos experiment"
            description: "Service {{ $labels.job }} is down. This may be expected during chaos testing."
```

**grafana-dashboard.json** (excerpt)
```json
{
  "dashboard": {
    "title": "Chaos Engineering Metrics",
    "panels": [
      {
        "title": "Chaos Experiments Status",
        "targets": [
          {
            "expr": "litmuschaos_experiment_verdict",
            "legendFormat": "{{ experiment }} - {{ verdict }}"
          }
        ]
      },
      {
        "title": "Pod Availability During Chaos",
        "targets": [
          {
            "expr": "kube_deployment_status_replicas_available{namespace='chaos-demo'} / kube_deployment_spec_replicas{namespace='chaos-demo'} * 100",
            "legendFormat": "{{ deployment }}"
          }
        ]
      },
      {
        "title": "Network Latency (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{ service }}"
          }
        ]
      }
    ]
  }
}
```

---

## 🧪 Testing Scenarios

### Scenario 1: Resilience to Pod Failures
```bash
# Apply pod delete chaos
kubectl apply -f litmus-pod-delete.yaml

# Monitor recovery
watch kubectl get pods -n chaos-demo

# Expected: Pods should recover within 30 seconds
```

### Scenario 2: Network Partition Tolerance
```bash
# Create network partition
kubectl apply -f chaos-mesh-network-partition.yaml

# Test service communication
kubectl run test --rm -it --image=curlimages/curl -- \
  curl http://frontend.chaos-demo.svc.cluster.local

# Expected: Graceful degradation or retry mechanisms
```

### Scenario 3: Resource Starvation
```bash
# Induce CPU stress
kubectl apply -f chaos-mesh-stress-cpu.yaml

# Monitor resource usage
kubectl top pods -n chaos-demo

# Expected: HPA should scale up pods, QoS should prevent eviction
```

---

## 📊 Success Metrics

- **Experiment Success Rate:** >95% of chaos experiments pass
- **Recovery Time Objective (RTO):** <2 minutes for pod failures
- **Service Availability:** >99.9% during chaos tests
- **Alert Accuracy:** Zero false negatives for critical failures
- **Incident Reduction:** 40% reduction in production incidents

---

## 🎓 Best Practices

1. **Start Small:** Begin with low-impact experiments (single pod failures)
2. **Steady State Validation:** Always define and verify steady state before/after chaos
3. **Blast Radius:** Limit experiment scope using namespaces and labels
4. **Observability First:** Ensure comprehensive monitoring before chaos testing
5. **GameDays:** Schedule regular chaos GameDays with team participation
6. **Automate Gradually:** Start manual, then automate proven experiments
7. **Document Learnings:** Maintain runbooks from each experiment
8. **Rollback Plan:** Always have immediate rollback procedures

---

## 🐛 Troubleshooting

**Chaos experiments not starting:**
```bash
# Check chaos operator logs
kubectl logs -n litmus -l app.kubernetes.io/component=operator

# Verify RBAC permissions
kubectl auth can-i create chaosengines --as=system:serviceaccount:chaos-demo:litmus-admin
```

**Pods not recovering:**
```bash
# Check deployment status
kubectl describe deployment frontend -n chaos-demo

# Review events
kubectl get events -n chaos-demo --sort-by='.lastTimestamp'
```

---

## 🚀 Next Steps

- Implement chaos experiments in staging before production
- Create custom chaos experiments for your specific failure modes
- Integrate with PagerDuty for incident management
- Build chaos engineering into your deployment pipeline
- Conduct quarterly GameDays with cross-functional teams

---

## 📚 Additional Resources

- [Litmus Documentation](https://docs.litmuschaos.io/)
- [Chaos Mesh Documentation](https://chaos-mesh.org/docs/)
- [Principles of Chaos Engineering](https://principlesofchaos.org/)
- [Google SRE Chaos Engineering](https://sre.google/workbook/chaos-engineering/)
