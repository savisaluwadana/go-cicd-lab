# Project 25: Event-Driven Architecture with Knative & KEDA

## 🎯 Project Overview

Build a serverless, event-driven architecture on Kubernetes using Knative for serverless containers and KEDA (Kubernetes Event-Driven Autoscaling) for advanced autoscaling based on event sources like message queues, metrics, and custom triggers.

### What You'll Learn
- Knative Serving for serverless workloads
- Knative Eventing for event-driven architectures
- KEDA scalers and triggers
- Event-driven autoscaling (0 to N)
- Integration with message brokers (Kafka, RabbitMQ, SQS)
- Cold start optimization
- Serverless observability
- Cost optimization through scale-to-zero

### Technologies
- **Serverless:** Knative 1.12 (Serving, Eventing)
- **Autoscaling:** KEDA 2.12
- **Messaging:** Apache Kafka, RabbitMQ, AWS SQS
- **Service Mesh:** Istio (optional)
- **Monitoring:** Prometheus, Grafana, Jaeger
- **Kubernetes:** 1.28+

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│              Event-Driven Serverless Platform                        │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │                    Event Sources                                │  │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐  │  │
│  │  │   HTTP    │  │   Kafka   │  │ RabbitMQ  │  │  AWS SQS  │  │  │
│  │  │  Requests │  │  Messages │  │   Queue   │  │   Queue   │  │  │
│  │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  │  │
│  └────────┼──────────────┼──────────────┼──────────────┼─────────┘  │
│           │              │              │              │             │
│  ┌────────▼──────────────▼──────────────▼──────────────▼─────────┐  │
│  │                  Knative Eventing                              │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │  │
│  │  │   Broker     │  │   Trigger    │  │  SinkBinding │        │  │
│  │  │  (InMemory,  │  │   (Filter &  │  │   (Event     │        │  │
│  │  │   Kafka)     │  │    Route)    │  │   Binding)   │        │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘        │  │
│  └────────────────────────────┬────────────────────────────────────┘  │
│                                │                                       │
│  ┌────────────────────────────▼────────────────────────────────────┐  │
│  │                   Knative Serving                               │  │
│  │  ┌──────────────────────────────────────────────────────────┐  │  │
│  │  │  Knative Service (KService)                              │  │  │
│  │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │  │  │
│  │  │  │ Revision 1 │  │ Revision 2 │  │ Revision 3 │         │  │  │
│  │  │  │  (20%)     │  │  (80%)     │  │  (Idle)    │         │  │  │
│  │  │  └────────────┘  └────────────┘  └────────────┘         │  │  │
│  │  │                                                           │  │  │
│  │  │  Features:                                                │  │  │
│  │  │  • Traffic Splitting       • Scale to Zero               │  │  │
│  │  │  • Auto-scaling (KPA)      • Blue/Green Deploy           │  │  │
│  │  │  • Revision Management     • Request-based scaling       │  │  │
│  │  └──────────────────────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                    KEDA Autoscaling                             │  │
│  │  ┌──────────────────────────────────────────────────────────┐  │  │
│  │  │  ScaledObject / ScaledJob                                │  │  │
│  │  │  ┌─────────────────────────────────────────────────────┐ │  │  │
│  │  │  │  Scalers:                                            │ │  │  │
│  │  │  │  • Kafka (lag, partition count)                     │ │  │  │
│  │  │  │  • RabbitMQ (queue depth)                           │ │  │  │
│  │  │  │  • Prometheus (custom metrics)                      │ │  │  │
│  │  │  │  • AWS CloudWatch, Azure Monitor                    │ │  │  │
│  │  │  │  • Cron (time-based scaling)                        │ │  │  │
│  │  │  └─────────────────────────────────────────────────────┘ │  │  │
│  │  │                                                            │  │  │
│  │  │  Auto-scales: Deployments, StatefulSets, Jobs            │  │  │
│  │  │  Range: 0 → 100+ replicas based on event metrics        │  │  │
│  │  └──────────────────────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │               Observability & Monitoring                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │  │
│  │  │Prometheus│  │ Grafana  │  │  Jaeger  │  │  Kiali   │       │  │
│  │  │(Metrics) │  │Dashboard │  │ Tracing  │  │ (Mesh)   │       │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │  │
│  └─────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- Kubernetes cluster (1.28+)
- kubectl configured
- Helm 3.x
- 8GB+ RAM available
- DNS configured for Knative
- (Optional) Istio for traffic management

---

## 🚀 Implementation

### Step 1: Install Knative Serving

```bash
# Install Knative CRDs
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.0/serving-crds.yaml

# Install Knative Serving core
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.0/serving-core.yaml

# Install networking layer (Kourier - lightweight alternative to Istio)
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.12.0/kourier.yaml

# Configure Knative to use Kourier
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

# Get Kourier external IP
kubectl get svc kourier -n kourier-system

# Configure DNS (Magic DNS for testing)
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.0/serving-default-domain.yaml

# Verify installation
kubectl get pods -n knative-serving
```

### Step 2: Install Knative Eventing

```bash
# Install Knative Eventing CRDs
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.0/eventing-crds.yaml

# Install Knative Eventing core
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.0/eventing-core.yaml

# Install In-Memory Channel (for development)
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.0/in-memory-channel.yaml

# Install Kafka Channel (for production)
kubectl apply -f https://github.com/knative-sandbox/eventing-kafka-broker/releases/download/knative-v1.12.0/eventing-kafka-controller.yaml
kubectl apply -f https://github.com/knative-sandbox/eventing-kafka-broker/releases/download/knative-v1.12.0/eventing-kafka-broker.yaml

# Verify installation
kubectl get pods -n knative-eventing
```

### Step 3: Install KEDA

```bash
# Add KEDA Helm repository
helm repo add kedacore https://kedacore.github.io/charts
helm repo update

# Install KEDA
helm install keda kedacore/keda --namespace keda --create-namespace

# Verify installation
kubectl get pods -n keda
```

### Step 4: Deploy Knative Service with Autoscaling

**knative-service.yaml**
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: event-processor
  namespace: default
spec:
  template:
    metadata:
      annotations:
        # Autoscaling configuration
        autoscaling.knative.dev/class: "kpa.autoscaling.knative.dev"  # Knative Pod Autoscaler
        autoscaling.knative.dev/metric: "concurrency"
        autoscaling.knative.dev/target: "10"  # Target 10 concurrent requests per pod
        autoscaling.knative.dev/minScale: "0"  # Scale to zero when idle
        autoscaling.knative.dev/maxScale: "50"
        autoscaling.knative.dev/scaleDownDelay: "30s"
        autoscaling.knative.dev/window: "60s"
    spec:
      containerConcurrency: 10
      containers:
      - name: event-processor
        image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Event Processor v1"
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 5
---
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: image-processor
  namespace: default
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/class: "kpa.autoscaling.knative.dev"
        autoscaling.knative.dev/metric: "rps"  # Requests per second
        autoscaling.knative.dev/target: "50"
        autoscaling.knative.dev/minScale: "1"  # Always keep 1 pod warm
        autoscaling.knative.dev/maxScale: "100"
    spec:
      containers:
      - name: processor
        image: your-registry/image-processor:v1.0
        env:
        - name: PROCESSING_TIMEOUT
          value: "30s"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
```

### Step 5: Configure Knative Eventing Broker & Triggers

**broker-trigger.yaml**
```yaml
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: default
  annotations:
    eventing.knative.dev/broker.class: MTChannelBasedBroker
spec:
  config:
    apiVersion: v1
    kind: ConfigMap
    name: config-br-default-channel
    namespace: knative-eventing
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: order-created-trigger
  namespace: default
spec:
  broker: default
  filter:
    attributes:
      type: com.example.order.created
      source: /orders/api
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: order-processor
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: payment-completed-trigger
  namespace: default
spec:
  broker: default
  filter:
    attributes:
      type: com.example.payment.completed
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: fulfillment-service
```

**event-producer.yaml** (Send events to broker)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: event-producer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: event-producer
  template:
    metadata:
      labels:
        app: event-producer
    spec:
      containers:
      - name: producer
        image: curlimages/curl
        command: ["sh", "-c"]
        args:
          - |
            while true; do
              curl -v "http://broker-ingress.knative-eventing.svc.cluster.local/default/default" \
                -X POST \
                -H "Ce-Id: $(uuidgen)" \
                -H "Ce-Specversion: 1.0" \
                -H "Ce-Type: com.example.order.created" \
                -H "Ce-Source: /orders/api" \
                -H "Content-Type: application/json" \
                -d '{"orderId": "12345", "amount": 99.99, "customerId": "cust-789"}'
              
              sleep 5
            done
```

### Step 6: KEDA with Kafka Scaler

**Install Kafka (for testing)**
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install kafka bitnami/kafka \
  --set replicaCount=3 \
  --set zookeeper.replicaCount=3 \
  --namespace kafka \
  --create-namespace
```

**keda-kafka-scaler.yaml**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kafka-secrets
  namespace: default
stringData:
  sasl: "plaintext"
  username: "user"
  password: "password"
---
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: kafka-consumer-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: kafka-consumer  # Target deployment
  pollingInterval: 30  # Check every 30 seconds
  cooldownPeriod: 120  # Wait 2 minutes before scaling down
  minReplicaCount: 0   # Scale to zero when no messages
  maxReplicaCount: 50
  triggers:
    - type: kafka
      metadata:
        bootstrapServers: kafka.kafka.svc.cluster.local:9092
        consumerGroup: my-consumer-group
        topic: orders
        lagThreshold: "10"  # Scale up when lag > 10 messages per replica
        activationLagThreshold: "5"  # Activate (scale from 0) when lag > 5
        offsetResetPolicy: latest
        allowIdleConsumers: "false"
      authenticationRef:
        name: kafka-trigger-auth
---
apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: kafka-trigger-auth
  namespace: default
spec:
  secretTargetRef:
    - parameter: sasl
      name: kafka-secrets
      key: sasl
    - parameter: username
      name: kafka-secrets
      key: username
    - parameter: password
      name: kafka-secrets
      key: password
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kafka-consumer
  namespace: default
spec:
  replicas: 1  # KEDA will manage this
  selector:
    matchLabels:
      app: kafka-consumer
  template:
    metadata:
      labels:
        app: kafka-consumer
    spec:
      containers:
      - name: consumer
        image: your-registry/kafka-consumer:v1.0
        env:
        - name: KAFKA_BROKERS
          value: "kafka.kafka.svc.cluster.local:9092"
        - name: KAFKA_TOPIC
          value: "orders"
        - name: KAFKA_GROUP_ID
          value: "my-consumer-group"
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
```

### Step 7: KEDA with RabbitMQ Scaler

**keda-rabbitmq-scaler.yaml**
```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: rabbitmq-consumer-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: rabbitmq-consumer
  minReplicaCount: 0
  maxReplicaCount: 30
  triggers:
    - type: rabbitmq
      metadata:
        host: amqp://guest:guest@rabbitmq.default.svc.cluster.local:5672
        queueName: tasks
        queueLength: "20"  # Target 20 messages per replica
        activationQueueLength: "5"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq-consumer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq-consumer
  template:
    metadata:
      labels:
        app: rabbitmq-consumer
    spec:
      containers:
      - name: consumer
        image: your-registry/task-processor:v1.0
        env:
        - name: RABBITMQ_URL
          value: "amqp://guest:guest@rabbitmq.default.svc.cluster.local:5672"
        - name: QUEUE_NAME
          value: "tasks"
```

### Step 8: KEDA with Prometheus Scaler

**keda-prometheus-scaler.yaml**
```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: custom-metric-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: api-server
  minReplicaCount: 2
  maxReplicaCount: 20
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring.svc:9090
        metricName: http_requests_per_second
        query: |
          sum(rate(http_requests_total{job="api-server"}[1m]))
        threshold: "100"  # Scale up when RPS > 100
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring.svc:9090
        metricName: api_latency_p99
        query: |
          histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="api-server"}[5m]))
        threshold: "0.5"  # Scale up when p99 latency > 500ms
```

### Step 9: KEDA ScaledJob for Batch Processing

**keda-scaledjob.yaml**
```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledJob
metadata:
  name: batch-processor
  namespace: default
spec:
  jobTargetRef:
    template:
      spec:
        containers:
        - name: processor
          image: your-registry/batch-processor:v1.0
          env:
          - name: SQS_QUEUE_URL
            value: "https://sqs.us-east-1.amazonaws.com/123456789/batch-jobs"
          resources:
            requests:
              memory: "1Gi"
              cpu: "500m"
        restartPolicy: Never
  pollingInterval: 30
  successfulJobsHistoryLimit: 5
  failedJobsHistoryLimit: 5
  maxReplicaCount: 100
  scalingStrategy:
    strategy: "accurate"  # Create 1 job per message
  triggers:
    - type: aws-sqs-queue
      metadata:
        queueURL: https://sqs.us-east-1.amazonaws.com/123456789/batch-jobs
        queueLength: "1"
        awsRegion: "us-east-1"
        identityOwner: pod
```

### Step 10: Cold Start Optimization

**optimized-knative-service.yaml**
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: low-latency-api
spec:
  template:
    metadata:
      annotations:
        # Keep 1 pod always warm
        autoscaling.knative.dev/minScale: "1"
        
        # Faster scale-up
        autoscaling.knative.dev/initialScale: "2"
        autoscaling.knative.dev/activationScale: "3"
        
        # Progressive rollout for zero downtime
        serving.knative.dev/rolloutDuration: "60s"
        
        # Enable request queueing during scale-up
        queue.sidecar.serving.knative.dev/resourcePercentage: "70"
    spec:
      containers:
      - name: api
        image: your-registry/api:v2.0
        # Fast startup application
        command: ["/app/server"]
        # Pre-warmed initialization
        lifecycle:
          postStart:
            exec:
              command: ["/bin/sh", "-c", "curl -X POST localhost:8080/warmup"]
        resources:
          requests:
            memory: "256Mi"
            cpu: "500m"
```

---

## 🧪 Testing

```bash
# Test Knative Service
kubectl get ksvc
curl http://event-processor.default.example.com

# Send CloudEvent to broker
kubectl run curl --image=curlimages/curl -it --rm -- \
  curl -v "http://broker-ingress.knative-eventing.svc.cluster.local/default/default" \
  -X POST \
  -H "Ce-Id: 12345" \
  -H "Ce-Specversion: 1.0" \
  -H "Ce-Type: com.example.order.created" \
  -H "Ce-Source: /orders/api" \
  -H "Content-Type: application/json" \
  -d '{"orderId": "789", "total": 150.00}'

# Monitor autoscaling
watch kubectl get pods
watch kubectl get scaledobject
watch kubectl get hpa

# Load test to trigger scaling
hey -z 2m -c 50 -q 10 http://event-processor.default.example.com
```

---

## 📊 Success Metrics

- **Cold Start Time:** <2 seconds
- **Scale-to-Zero Idle Time:** 30 seconds
- **Scale-Up Time (0→10 pods):** <15 seconds
- **Cost Savings:** 60% reduction vs always-on workloads
- **Event Processing Latency:** <100ms p99

---

## 🎓 Best Practices

1. **Minimize Cold Starts:** Use `minScale: 1` for latency-sensitive services
2. **Right-size Resources:** Set appropriate CPU/memory requests
3. **Use Health Checks:** Implement readiness/liveness probes
4. **Event Filtering:** Use Trigger filters to reduce unnecessary invocations
5. **Monitor Queue Lag:** Set appropriate KEDA thresholds
6. **Idempotency:** Ensure event handlers are idempotent
7. **Dead Letter Queues:** Configure DLQs for failed events

---

## 📚 Additional Resources

- [Knative Documentation](https://knative.dev/docs/)
- [KEDA Documentation](https://keda.sh/docs/)
- [CloudEvents Specification](https://cloudevents.io/)
- [Serverless on Kubernetes](https://www.manning.com/books/kubernetes-native-microservices-with-quarkus-and-microprofile)
