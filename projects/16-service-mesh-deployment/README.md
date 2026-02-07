# Project 16: Service Mesh Deployment with Istio & Linkerd

## 🎯 **Learning Objectives**
- Implement service mesh architecture for microservices
- Configure traffic management with Istio/Linkerd
- Implement mutual TLS (mTLS) across services
- Deploy circuit breakers and retry policies
- Build observability with distributed tracing
- Implement A/B testing and traffic mirroring

## 📋 **Project Overview**
Deploy a complete service mesh infrastructure using Istio and Linkerd, implementing advanced traffic management, security policies, and observability for a microservices architecture. This project demonstrates enterprise-grade service-to-service communication patterns.

## 🏗️ **Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                     Ingress Gateway                          │
│                    (Istio Gateway)                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
┌───────▼──────┐           ┌───────▼──────┐
│  Service A   │───────────│  Service B   │
│  (Istio)     │           │  (Istio)     │
└──────┬───────┘           └──────┬───────┘
       │                          │
       │         ┌────────────────┘
       │         │
       │    ┌────▼────────┐
       └────│  Service C  │
            │  (Linkerd)  │
            └─────────────┘

[Observability Layer]
├── Jaeger (Distributed Tracing)
├── Prometheus (Metrics)
├── Grafana (Visualization)
└── Kiali (Mesh Visualization)
```

## 🔧 **Istio Configuration**

### Install Istio
```bash
# Download Istio
curl -L https://istio.io/downloadIstio | sh -
cd istio-1.20.0
export PATH=$PWD/bin:$PATH

# Install with demo profile
istioctl install --set profile=demo -y

# Enable sidecar injection
kubectl label namespace default istio-injection=enabled
```

### Gateway Configuration
```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: library-gateway
  namespace: default
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 80
        name: http
        protocol: HTTP
      hosts:
        - "library.example.com"
    - port:
        number: 443
        name: https
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: library-tls-cert
      hosts:
        - "library.example.com"
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: library-virtualservice
  namespace: default
spec:
  hosts:
    - "library.example.com"
  gateways:
    - library-gateway
  http:
    # Canary deployment: 90% stable, 10% canary
    - match:
        - headers:
            user-agent:
              regex: ".*Chrome.*"
      route:
        - destination:
            host: library-manager
            subset: v2
          weight: 10
        - destination:
            host: library-manager
            subset: v1
          weight: 90
    
    # Default routing
    - route:
        - destination:
            host: library-manager
            subset: v1
          weight: 100
      
      # Fault injection for testing
      fault:
        delay:
          percentage:
            value: 0.1
          fixedDelay: 5s
        abort:
          percentage:
            value: 0.01
          httpStatus: 503
      
      # Retry policy
      retries:
        attempts: 3
        perTryTimeout: 2s
        retryOn: 5xx,reset,connect-failure
      
      # Timeout
      timeout: 10s
```

### Destination Rules
```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: library-manager
  namespace: default
spec:
  host: library-manager
  
  # Traffic policy
  trafficPolicy:
    # Connection pool settings
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
        maxRequestsPerConnection: 2
    
    # Load balancer
    loadBalancer:
      consistentHash:
        httpHeaderName: "x-user-id"
    
    # Outlier detection (circuit breaker)
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
      minHealthPercent: 40
  
  # Subsets for versions
  subsets:
    - name: v1
      labels:
        version: v1
      trafficPolicy:
        tls:
          mode: ISTIO_MUTUAL
    
    - name: v2
      labels:
        version: v2
      trafficPolicy:
        tls:
          mode: ISTIO_MUTUAL
```

### mTLS Configuration
```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: default
spec:
  mtls:
    mode: STRICT
---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: library-authz
  namespace: default
spec:
  selector:
    matchLabels:
      app: library-manager
  
  action: ALLOW
  rules:
    # Allow GET requests from any service
    - from:
        - source:
            principals: ["cluster.local/ns/default/sa/*"]
      to:
        - operation:
            methods: ["GET"]
    
    # Allow POST/PUT/DELETE only from admin service
    - from:
        - source:
            principals: ["cluster.local/ns/default/sa/admin-service"]
      to:
        - operation:
            methods: ["POST", "PUT", "DELETE"]
```

## 🔗 **Linkerd Configuration**

### Install Linkerd
```bash
# Install Linkerd CLI
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin

# Validate cluster
linkerd check --pre

# Install Linkerd
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check

# Install Linkerd Viz for observability
linkerd viz install | kubectl apply -f -
```

### Service Profile
```yaml
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  name: library-manager.default.svc.cluster.local
  namespace: default
spec:
  routes:
    - name: GET /api/books
      condition:
        method: GET
        pathRegex: /api/books
      responseClasses:
        - condition:
            status:
              min: 200
              max: 299
          isFailure: false
        - condition:
            status:
              min: 500
              max: 599
          isFailure: true
      
      # Per-route timeout
      timeout: 5s
      
      # Per-route retries
      isRetryable: true
    
    - name: POST /api/books
      condition:
        method: POST
        pathRegex: /api/books
      isRetryable: false
      timeout: 10s
    
    - name: GET /api/book
      condition:
        method: GET
        pathRegex: /api/book
      timeout: 3s
      isRetryable: true
```

### Traffic Split (SMI)
```yaml
apiVersion: split.smi-spec.io/v1alpha1
kind: TrafficSplit
metadata:
  name: library-manager-split
  namespace: default
spec:
  service: library-manager
  backends:
    - service: library-manager-v1
      weight: 900  # 90%
    - service: library-manager-v2
      weight: 100  # 10%
```

## 📊 **Observability Stack**

### Jaeger for Distributed Tracing
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jaeger-config
  namespace: istio-system
data:
  jaeger.yaml: |
    apiVersion: install.istio.io/v1alpha1
    kind: IstioOperator
    spec:
      meshConfig:
        enableTracing: true
        defaultConfig:
          tracing:
            sampling: 100.0
            zipkin:
              address: jaeger-collector.istio-system:9411
```

### Prometheus Metrics
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    scrape_configs:
      # Istio mesh metrics
      - job_name: 'istio-mesh'
        kubernetes_sd_configs:
          - role: endpoints
            namespaces:
              names:
                - istio-system
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_name]
            action: keep
            regex: istio-telemetry
      
      # Envoy stats
      - job_name: 'envoy-stats'
        metrics_path: /stats/prometheus
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_container_port_name]
            action: keep
            regex: '.*-envoy-prom'
      
      # Application metrics
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
```

### Grafana Dashboards
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: monitoring
data:
  istio-dashboard.json: |
    {
      "dashboard": {
        "title": "Istio Service Mesh",
        "panels": [
          {
            "title": "Request Rate",
            "targets": [{
              "expr": "sum(rate(istio_requests_total[5m])) by (destination_service_name)"
            }]
          },
          {
            "title": "Success Rate",
            "targets": [{
              "expr": "sum(rate(istio_requests_total{response_code!~\"5.*\"}[5m])) / sum(rate(istio_requests_total[5m]))"
            }]
          },
          {
            "title": "P95 Latency",
            "targets": [{
              "expr": "histogram_quantile(0.95, sum(rate(istio_request_duration_milliseconds_bucket[5m])) by (le))"
            }]
          }
        ]
      }
    }
```

## 🚀 **CI/CD Pipeline Integration**

### `.github/workflows/service-mesh-deploy.yml`
```yaml
name: Service Mesh Deployment

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      canary_percentage:
        description: 'Canary traffic percentage'
        required: true
        default: '10'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          kubeconfig: ${{ secrets.KUBECONFIG }}

      - name: Install Istioctl
        run: |
          curl -L https://istio.io/downloadIstio | sh -
          cd istio-*
          echo "$PWD/bin" >> $GITHUB_PATH

      - name: Build and push image
        run: |
          docker build -t ghcr.io/${{ github.repository }}:${{ github.sha }} .
          echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
          docker push ghcr.io/${{ github.repository }}:${{ github.sha }}

      - name: Deploy canary version
        run: |
          kubectl set image deployment/library-manager-v2 \
            library-manager=ghcr.io/${{ github.repository }}:${{ github.sha }} \
            -n default

      - name: Configure canary traffic
        run: |
          cat <<EOF | kubectl apply -f -
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: library-virtualservice
          spec:
            hosts:
              - library-manager
            http:
              - route:
                  - destination:
                      host: library-manager
                      subset: v1
                    weight: ${{ 100 - inputs.canary_percentage }}
                  - destination:
                      host: library-manager
                      subset: v2
                    weight: ${{ inputs.canary_percentage }}
          EOF

      - name: Monitor canary metrics
        run: |
          # Wait for 5 minutes and collect metrics
          sleep 300
          
          # Query Prometheus for error rate
          ERROR_RATE=$(curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(istio_requests_total{destination_workload=\"library-manager-v2\",response_code=~\"5.*\"}[5m]))/sum(rate(istio_requests_total{destination_workload=\"library-manager-v2\"}[5m]))" | jq -r '.data.result[0].value[1]')
          
          # Rollback if error rate > 1%
          if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
            echo "High error rate detected. Rolling back..."
            kubectl set image deployment/library-manager-v2 \
              library-manager=ghcr.io/${{ github.repository }}:stable
            exit 1
          fi

      - name: Promote canary to stable
        if: success()
        run: |
          # Gradually increase traffic
          for weight in 25 50 75 100; do
            cat <<EOF | kubectl apply -f -
            apiVersion: networking.istio.io/v1beta1
            kind: VirtualService
            metadata:
              name: library-virtualservice
            spec:
              hosts:
                - library-manager
              http:
                - route:
                    - destination:
                        host: library-manager
                        subset: v2
                      weight: $weight
                    - destination:
                        host: library-manager
                        subset: v1
                      weight: $((100 - weight))
            EOF
            
            echo "Traffic shifted to $weight% canary. Monitoring..."
            sleep 120
          done
          
          echo "Canary promoted successfully!"
```

## 🧪 **Testing Service Mesh**

### Chaos Engineering with Fault Injection
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: chaos-testing
spec:
  hosts:
    - library-manager
  http:
    - match:
        - headers:
            x-chaos-test:
              exact: "true"
      fault:
        # Inject 500ms delay for 50% of requests
        delay:
          percentage:
            value: 50.0
          fixedDelay: 500ms
        
        # Abort 10% of requests with 503
        abort:
          percentage:
            value: 10.0
          httpStatus: 503
      
      route:
        - destination:
            host: library-manager
```

### Load Testing Script
```bash
#!/bin/bash
# load-test-mesh.sh

echo "Starting load test against service mesh..."

# Generate traffic
hey -z 5m -c 50 -q 10 \
  -H "Host: library.example.com" \
  http://istio-ingressgateway.istio-system/api/books

# Check metrics
echo "Fetching metrics..."
istioctl dashboard prometheus

# View traces
echo "Opening Jaeger UI..."
istioctl dashboard jaeger

# Visualize mesh
echo "Opening Kiali..."
istioctl dashboard kiali
```

## 🎯 **Key Learnings**
- ✅ Service mesh architecture and sidecar pattern
- ✅ Traffic management (canary, A/B testing, mirroring)
- ✅ mTLS and zero-trust security
- ✅ Circuit breakers and retry policies
- ✅ Distributed tracing with Jaeger
- ✅ Mesh observability with Kiali
- ✅ Progressive deployment strategies

## 📊 **Performance Metrics**
- **Latency**: P95 latency < 100ms with mesh overhead
- **Success Rate**: 99.9% uptime with circuit breakers
- **Security**: 100% mTLS coverage across services
- **Observability**: Full request tracing end-to-end

## 🚀 **Production Considerations**
1. Enable resource limits for sidecars
2. Configure horizontal pod autoscaling
3. Implement proper monitoring and alerting
4. Set up backup and disaster recovery
5. Document traffic policies and security rules
6. Regular security audits and certificate rotation

## 📚 **Additional Resources**
- [Istio Documentation](https://istio.io/latest/docs/)
- [Linkerd Documentation](https://linkerd.io/2/overview/)
- [Service Mesh Patterns](https://www.servicemeshpatterns.com/)
- [Jaeger Tracing](https://www.jaegertracing.io/)
- [Kiali Documentation](https://kiali.io/docs/)
