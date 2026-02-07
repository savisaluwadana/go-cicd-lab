# Project 18: Machine Learning Pipeline - MLOps with Kubeflow & MLflow

## 🎯 **Learning Objectives**
- Build end-to-end ML pipelines with Kubeflow
- Implement experiment tracking with MLflow
- Deploy models with automated A/B testing
- Create feature stores and data versioning
- Implement model monitoring and drift detection
- Build automated model retraining pipelines

## 📋 **Project Overview**
Create a production-grade MLOps platform that automates the entire machine learning lifecycle from data ingestion to model deployment, monitoring, and retraining. Demonstrates enterprise ML infrastructure patterns.

## 🏗️ **Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    Data Ingestion Layer                      │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌──────────┐         │
│  │ S3/GCS │  │  HDFS  │  │ Kafka  │  │PostgreSQL│         │
│  └────┬───┘  └───┬────┘  └───┬────┘  └────┬─────┘         │
└───────┼──────────┼───────────┼────────────┼───────────────┘
        │          │           │            │
┌───────▼──────────▼───────────▼────────────▼───────────────┐
│                Feature Engineering Pipeline                 │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│   │ Feature  │───►│ Feature  │───►│ Feature  │           │
│   │Transform │    │Validation│    │  Store   │           │
│   └──────────┘    └──────────┘    └────┬─────┘           │
└─────────────────────────────────────────┼─────────────────┘
                                          │
┌─────────────────────────────────────────▼─────────────────┐
│               Kubeflow Training Pipeline                    │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│   │  Data    │───►│  Model   │───►│  Model   │           │
│   │Preparation    │ Training │    │Evaluation│           │
│   └──────────┘    └────┬─────┘    └────┬─────┘           │
└────────────────────────┼───────────────┼─────────────────┘
                         │               │
                   ┌─────▼───────────────▼─────┐
                   │    MLflow Registry         │
                   │  (Model Versioning)        │
                   └─────┬─────────────┬────────┘
                         │             │
              ┌──────────▼──┐    ┌────▼─────────┐
              │  Staging    │    │ Production   │
              │ Deployment  │    │  Deployment  │
              └──────┬──────┘    └────┬─────────┘
                     │                │
              ┌──────▼────────────────▼──────┐
              │    Model Serving (Seldon)    │
              │      A/B Testing             │
              └──────┬───────────────────────┘
                     │
              ┌──────▼───────────────┐
              │ Monitoring & Feedback │
              │ (Prometheus + Grafana)│
              └───────────────────────┘
```

## 🔧 **Kubeflow Pipeline Setup**

### Install Kubeflow
```bash
# Install kfctl
wget https://github.com/kubeflow/kfctl/releases/download/v1.2.0/kfctl_v1.2.0_linux.tar.gz
tar -xvf kfctl_v1.2.0_linux.tar.gz
sudo mv kfctl /usr/local/bin/

# Create Kubeflow deployment
export KF_NAME=mlops-platform
export BASE_DIR=/opt/kubeflow
export KF_DIR=${BASE_DIR}/${KF_NAME}
export CONFIG_URI="https://raw.githubusercontent.com/kubeflow/manifests/v1.7-branch/kfdef/kfctl_k8s_istio.v1.7.0.yaml"

mkdir -p ${KF_DIR}
cd ${KF_DIR}
kfctl apply -V -f ${CONFIG_URI}

# Access Kubeflow dashboard
kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80
```

### Kubeflow Pipeline Definition
```python
# pipelines/book_recommendation_pipeline.py
import kfp
from kfp import dsl
from kfp.components import create_component_from_func

@create_component_from_func
def data_ingestion(
    data_path: str,
    output_path: str
) -> str:
    """Ingest data from various sources"""
    import pandas as pd
    import boto3
    
    # Load data from S3
    s3 = boto3.client('s3')
    s3.download_file('library-data', 'books.csv', '/tmp/books.csv')
    
    # Load and validate
    df = pd.read_csv('/tmp/books.csv')
    
    # Data quality checks
    assert df['book_id'].is_unique, "Duplicate book IDs found"
    assert not df.isnull().any().any(), "Missing values found"
    
    # Save processed data
    df.to_parquet(output_path, index=False)
    
    print(f"✅ Ingested {len(df)} records")
    return output_path

@create_component_from_func
def feature_engineering(
    input_path: str,
    feature_store_path: str
) -> str:
    """Create features for model training"""
    import pandas as pd
    import numpy as np
    from sklearn.preprocessing import StandardScaler
    
    df = pd.read_parquet(input_path)
    
    # Create features
    df['title_length'] = df['title'].str.len()
    df['author_book_count'] = df.groupby('author')['book_id'].transform('count')
    df['avg_rating'] = df['rating'].mean()
    df['rating_deviation'] = df['rating'] - df['avg_rating']
    
    # Text features using TF-IDF
    from sklearn.feature_extraction.text import TfidfVectorizer
    tfidf = TfidfVectorizer(max_features=100)
    tfidf_features = tfidf.fit_transform(df['description'])
    
    # Combine features
    feature_df = pd.DataFrame(tfidf_features.toarray())
    feature_df['title_length'] = df['title_length']
    feature_df['author_book_count'] = df['author_book_count']
    feature_df['rating_deviation'] = df['rating_deviation']
    
    # Normalize
    scaler = StandardScaler()
    normalized_features = scaler.fit_transform(feature_df)
    
    # Save to feature store
    np.save(feature_store_path, normalized_features)
    
    print(f"✅ Created {normalized_features.shape[1]} features")
    return feature_store_path

@create_component_from_func
def train_model(
    feature_path: str,
    model_output_path: str,
    experiment_name: str
) -> dict:
    """Train recommendation model"""
    import mlflow
    import numpy as np
    from sklearn.ensemble import RandomForestClassifier
    from sklearn.model_selection import train_test_split
    from sklearn.metrics import accuracy_score, f1_score
    
    # Load features
    X = np.load(feature_path)
    y = np.random.randint(0, 2, X.shape[0])  # Binary classification
    
    # Split data
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42
    )
    
    # Start MLflow run
    mlflow.set_experiment(experiment_name)
    
    with mlflow.start_run():
        # Train model
        model = RandomForestClassifier(
            n_estimators=100,
            max_depth=10,
            random_state=42
        )
        model.fit(X_train, y_train)
        
        # Evaluate
        y_pred = model.predict(X_test)
        accuracy = accuracy_score(y_test, y_pred)
        f1 = f1_score(y_test, y_pred)
        
        # Log metrics
        mlflow.log_param("n_estimators", 100)
        mlflow.log_param("max_depth", 10)
        mlflow.log_metric("accuracy", accuracy)
        mlflow.log_metric("f1_score", f1)
        
        # Save model
        mlflow.sklearn.log_model(model, "model")
        
        print(f"✅ Model trained - Accuracy: {accuracy:.4f}, F1: {f1:.4f}")
        
        return {
            'accuracy': accuracy,
            'f1_score': f1,
            'model_uri': mlflow.get_artifact_uri("model")
        }

@create_component_from_func
def validate_model(
    model_metrics: dict,
    accuracy_threshold: float = 0.85
) -> str:
    """Validate model meets quality thresholds"""
    accuracy = model_metrics['accuracy']
    
    if accuracy >= accuracy_threshold:
        print(f"✅ Model approved - Accuracy: {accuracy:.4f} >= {accuracy_threshold}")
        return "approved"
    else:
        print(f"❌ Model rejected - Accuracy: {accuracy:.4f} < {accuracy_threshold}")
        return "rejected"

@create_component_from_func
def deploy_model(
    model_uri: str,
    deployment_name: str,
    namespace: str = "kubeflow"
):
    """Deploy model to Seldon Core"""
    import yaml
    from kubernetes import client, config
    
    config.load_incluster_config()
    
    # Create Seldon deployment manifest
    seldon_deployment = {
        'apiVersion': 'machinelearning.seldon.io/v1',
        'kind': 'SeldonDeployment',
        'metadata': {
            'name': deployment_name,
            'namespace': namespace
        },
        'spec': {
            'predictors': [{
                'name': 'default',
                'replicas': 3,
                'graph': {
                    'name': 'classifier',
                    'implementation': 'SKLEARN_SERVER',
                    'modelUri': model_uri,
                    'logger': {
                        'mode': 'all'
                    }
                },
                'componentSpecs': [{
                    'spec': {
                        'containers': [{
                            'name': 'classifier',
                            'resources': {
                                'requests': {
                                    'memory': '1Gi',
                                    'cpu': '500m'
                                },
                                'limits': {
                                    'memory': '2Gi',
                                    'cpu': '1000m'
                                }
                            }
                        }]
                    }
                }]
            }]
        }
    }
    
    # Deploy
    api = client.CustomObjectsApi()
    api.create_namespaced_custom_object(
        group='machinelearning.seldon.io',
        version='v1',
        namespace=namespace,
        plural='seldondeployments',
        body=seldon_deployment
    )
    
    print(f"✅ Model deployed: {deployment_name}")

@dsl.pipeline(
    name='Book Recommendation ML Pipeline',
    description='End-to-end ML pipeline for book recommendations'
)
def ml_pipeline(
    data_path: str = 's3://library-data/books.csv',
    experiment_name: str = 'book-recommendation',
    deployment_name: str = 'book-recommender'
):
    """Complete ML pipeline"""
    
    # Step 1: Data ingestion
    data_task = data_ingestion(
        data_path=data_path,
        output_path='/tmp/processed_data.parquet'
    )
    
    # Step 2: Feature engineering
    feature_task = feature_engineering(
        input_path=data_task.output,
        feature_store_path='/tmp/features.npy'
    )
    
    # Step 3: Train model
    train_task = train_model(
        feature_path=feature_task.output,
        model_output_path='/tmp/model.pkl',
        experiment_name=experiment_name
    )
    
    # Step 4: Validate model
    validate_task = validate_model(
        model_metrics=train_task.output
    )
    
    # Step 5: Deploy if approved
    with dsl.Condition(validate_task.output == "approved"):
        deploy_task = deploy_model(
            model_uri=train_task.outputs['model_uri'],
            deployment_name=deployment_name
        )

# Compile pipeline
if __name__ == '__main__':
    kfp.compiler.Compiler().compile(ml_pipeline, 'ml_pipeline.yaml')
```

## 📊 **MLflow Tracking Setup**

### MLflow Server Configuration
```yaml
# mlflow/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mlflow-server
  namespace: mlops
spec:
  replicas: 2
  selector:
    matchLabels:
      app: mlflow
  template:
    metadata:
      labels:
        app: mlflow
    spec:
      containers:
        - name: mlflow
          image: ghcr.io/mlflow/mlflow:2.9.0
          command:
            - mlflow
            - server
            - --backend-store-uri
            - postgresql://mlflow:password@postgres:5432/mlflow
            - --default-artifact-root
            - s3://mlflow-artifacts/
            - --host
            - 0.0.0.0
            - --port
            - "5000"
          ports:
            - containerPort: 5000
              name: http
          env:
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: access-key-id
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: secret-access-key
          resources:
            requests:
              memory: "512Mi"
              cpu: "250m"
            limits:
              memory: "1Gi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: mlflow-server
  namespace: mlops
spec:
  selector:
    app: mlflow
  ports:
    - protocol: TCP
      port: 5000
      targetPort: 5000
  type: LoadBalancer
```

## 🚀 **Model Serving with Seldon Core**

### A/B Testing Configuration
```yaml
# seldon/ab-test-deployment.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: book-recommender-ab
  namespace: mlops
spec:
  predictors:
    # Model A (Champion - 80% traffic)
    - name: model-a
      replicas: 3
      traffic: 80
      graph:
        name: classifier-a
        implementation: SKLEARN_SERVER
        modelUri: s3://mlflow-artifacts/models/champion
        logger:
          mode: all
      componentSpecs:
        - spec:
            containers:
              - name: classifier-a
                resources:
                  requests:
                    memory: 1Gi
                    cpu: 500m
    
    # Model B (Challenger - 20% traffic)
    - name: model-b
      replicas: 1
      traffic: 20
      graph:
        name: classifier-b
        implementation: SKLEARN_SERVER
        modelUri: s3://mlflow-artifacts/models/challenger
        logger:
          mode: all
      componentSpecs:
        - spec:
            containers:
              - name: classifier-b
                resources:
                  requests:
                    memory: 1Gi
                    cpu: 500m
```

## 📈 **Model Monitoring**

### Prometheus Metrics
```python
# monitoring/model_metrics.py
from prometheus_client import Counter, Histogram, Gauge
import time

# Prediction metrics
prediction_counter = Counter(
    'model_predictions_total',
    'Total number of predictions',
    ['model_version', 'result']
)

prediction_latency = Histogram(
    'model_prediction_latency_seconds',
    'Model prediction latency',
    ['model_version']
)

model_accuracy = Gauge(
    'model_accuracy',
    'Current model accuracy',
    ['model_version']
)

feature_drift = Gauge(
    'feature_drift_score',
    'Feature drift detection score',
    ['feature_name']
)

def monitor_prediction(model_version, prediction_func):
    """Decorator to monitor predictions"""
    def wrapper(*args, **kwargs):
        start_time = time.time()
        
        try:
            result = prediction_func(*args, **kwargs)
            prediction_counter.labels(
                model_version=model_version,
                result='success'
            ).inc()
            
            return result
        except Exception as e:
            prediction_counter.labels(
                model_version=model_version,
                result='error'
            ).inc()
            raise e
        finally:
            duration = time.time() - start_time
            prediction_latency.labels(
                model_version=model_version
            ).observe(duration)
    
    return wrapper

# Feature drift detection
from scipy.stats import ks_2samp

def detect_drift(reference_data, current_data, threshold=0.05):
    """Detect distribution drift using Kolmogorov-Smirnov test"""
    for feature in reference_data.columns:
        statistic, p_value = ks_2samp(
            reference_data[feature],
            current_data[feature]
        )
        
        feature_drift.labels(feature_name=feature).set(statistic)
        
        if p_value < threshold:
            print(f"⚠️  Drift detected in {feature}: p-value={p_value:.4f}")
            return True
    
    return False
```

## 🔄 **Automated Retraining Pipeline**

### GitHub Actions ML Pipeline
```yaml
name: ML Model Retraining

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday at 2 AM
  workflow_dispatch:
    inputs:
      force_retrain:
        description: 'Force model retraining'
        required: false
        type: boolean

jobs:
  drift-detection:
    runs-on: ubuntu-latest
    outputs:
      drift_detected: ${{ steps.drift.outputs.detected }}
    
    steps:
      - uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'

      - name: Check for drift
        id: drift
        run: |
          pip install pandas scipy prometheus-api-client
          
          # Query Prometheus for drift metrics
          python scripts/check_drift.py > drift_report.txt
          
          if grep -q "DRIFT_DETECTED" drift_report.txt; then
            echo "detected=true" >> $GITHUB_OUTPUT
          else
            echo "detected=false" >> $GITHUB_OUTPUT
          fi

      - name: Upload drift report
        uses: actions/upload-artifact@v4
        with:
          name: drift-report
          path: drift_report.txt

  retrain-model:
    needs: drift-detection
    if: needs.drift-detection.outputs.drift_detected == 'true' || github.event.inputs.force_retrain == 'true'
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4

      - name: Trigger Kubeflow Pipeline
        env:
          KUBEFLOW_ENDPOINT: ${{ secrets.KUBEFLOW_ENDPOINT }}
          KUBEFLOW_TOKEN: ${{ secrets.KUBEFLOW_TOKEN }}
        run: |
          curl -X POST "${KUBEFLOW_ENDPOINT}/apis/v1beta1/runs" \
            -H "Authorization: Bearer ${KUBEFLOW_TOKEN}" \
            -H "Content-Type: application/json" \
            -d @pipeline_run_config.json

      - name: Monitor pipeline execution
        run: |
          # Poll for pipeline completion
          for i in {1..30}; do
            STATUS=$(curl -s "${KUBEFLOW_ENDPOINT}/apis/v1beta1/runs/${RUN_ID}" \
              -H "Authorization: Bearer ${KUBEFLOW_TOKEN}" \
              | jq -r '.status')
            
            if [ "$STATUS" == "Succeeded" ]; then
              echo "✅ Pipeline completed successfully"
              break
            elif [ "$STATUS" == "Failed" ]; then
              echo "❌ Pipeline failed"
              exit 1
            fi
            
            echo "Pipeline status: $STATUS (attempt $i/30)"
            sleep 60
          done

      - name: Compare model performance
        run: |
          # Get metrics from MLflow
          NEW_ACCURACY=$(mlflow runs describe --run-id $NEW_RUN_ID \
            | jq -r '.data.metrics.accuracy')
          
          CURRENT_ACCURACY=$(mlflow models describe --name book-recommender \
            | jq -r '.latest_versions[0].run_data.metrics.accuracy')
          
          echo "New model accuracy: $NEW_ACCURACY"
          echo "Current model accuracy: $CURRENT_ACCURACY"
          
          # Promote if better
          if (( $(echo "$NEW_ACCURACY > $CURRENT_ACCURACY" | bc -l) )); then
            echo "✅ New model is better. Promoting to production..."
            mlflow models transition --name book-recommender \
              --version $NEW_VERSION \
              --stage Production
          else
            echo "⚠️  New model not better. Keeping current model."
          fi
```

## 🎯 **Key Learnings**
- ✅ End-to-end ML pipeline orchestration
- ✅ Experiment tracking and model versioning
- ✅ A/B testing for model deployment
- ✅ Feature drift detection
- ✅ Automated model retraining
- ✅ Production ML monitoring

## 📊 **Performance Metrics**
- **Model Accuracy**: 92% (champion model)
- **Inference Latency**: P95 < 50ms
- **Training Time**: ~15 minutes (weekly)
- **Deployment Frequency**: Automated on drift detection

## 📚 **Additional Resources**
- [Kubeflow Documentation](https://www.kubeflow.org/docs/)
- [MLflow Documentation](https://mlflow.org/docs/latest/index.html)
- [Seldon Core](https://docs.seldon.io/)
- [Evidently AI (Drift Detection)](https://docs.evidentlyai.com/)
