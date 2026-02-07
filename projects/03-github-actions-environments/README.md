# Project 3: Multi-Environment Deployment with GitHub Actions

## Overview
Deploy to multiple environments (Dev, Staging, Production) with approval gates, environment-specific configurations, and rollback capabilities.

## What You'll Learn
- GitHub Environments and protection rules
- Manual approval workflows
- Environment secrets
- Deployment strategies
- Rollback procedures

## Project Structure
```
03-github-actions-environments/
├── .github/
│   └── workflows/
│       ├── deploy.yaml
│       └── rollback.yaml
├── environments/
│   ├── dev/
│   │   └── config.yaml
│   ├── staging/
│   │   └── config.yaml
│   └── production/
│       └── config.yaml
├── scripts/
│   ├── deploy.sh
│   └── health-check.sh
└── README.md
```

## Implementation

### 1. Deployment Workflow

**.github/workflows/deploy.yaml:**
```yaml
name: Multi-Environment Deployment

on:
  push:
    branches: [ main, develop ]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        type: choice
        options:
          - dev
          - staging
          - production

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: |
          docker build -t app:${{ github.sha }} .
          docker save app:${{ github.sha }} > app.tar
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: docker-image
          path: app.tar

  deploy-dev:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: dev
      url: https://dev.example.com
    if: github.ref == 'refs/heads/develop'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Download artifact
        uses: actions/download-artifact@v3
        with:
          name: docker-image
      
      - name: Load Docker image
        run: docker load < app.tar
      
      - name: Deploy to Dev
        env:
          DEPLOY_KEY: ${{ secrets.DEV_DEPLOY_KEY }}
          DB_URL: ${{ secrets.DEV_DB_URL }}
        run: |
          ./scripts/deploy.sh dev ${{ github.sha }}
      
      - name: Health check
        run: |
          ./scripts/health-check.sh https://dev.example.com

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.example.com
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Download artifact
        uses: actions/download-artifact@v3
        with:
          name: docker-image
      
      - name: Load Docker image
        run: docker load < app.tar
      
      - name: Deploy to Staging
        env:
          DEPLOY_KEY: ${{ secrets.STAGING_DEPLOY_KEY }}
          DB_URL: ${{ secrets.STAGING_DB_URL }}
        run: |
          ./scripts/deploy.sh staging ${{ github.sha }}
      
      - name: Run smoke tests
        run: |
          npm install -g newman
          newman run tests/api-tests.json --env-var base_url=https://staging.example.com
      
      - name: Health check
        run: |
          ./scripts/health-check.sh https://staging.example.com

  deploy-production:
    needs: [build, deploy-staging]
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://example.com
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Download artifact
        uses: actions/download-artifact@v3
        with:
          name: docker-image
      
      - name: Load Docker image
        run: docker load < app.tar
      
      - name: Create deployment
        uses: chrnorm/deployment-action@v2
        id: deployment
        with:
          token: ${{ github.token }}
          environment: production
      
      - name: Deploy to Production
        env:
          DEPLOY_KEY: ${{ secrets.PROD_DEPLOY_KEY }}
          DB_URL: ${{ secrets.PROD_DB_URL }}
        run: |
          ./scripts/deploy.sh production ${{ github.sha }}
      
      - name: Health check
        id: health
        run: |
          ./scripts/health-check.sh https://example.com
      
      - name: Update deployment status (success)
        if: success()
        uses: chrnorm/deployment-status@v2
        with:
          token: ${{ github.token }}
          deployment-id: ${{ steps.deployment.outputs.deployment_id }}
          state: success
          environment-url: https://example.com
      
      - name: Update deployment status (failure)
        if: failure()
        uses: chrnorm/deployment-status@v2
        with:
          token: ${{ github.token }}
          deployment-id: ${{ steps.deployment.outputs.deployment_id }}
          state: failure
      
      - name: Notify Slack
        if: always()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK }}
          payload: |
            {
              "text": "Production deployment ${{ job.status }}",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Deployment Status:* ${{ job.status }}\n*Environment:* Production\n*Commit:* ${{ github.sha }}\n*URL:* https://example.com"
                  }
                }
              ]
            }
```

### 2. Rollback Workflow

**.github/workflows/rollback.yaml:**
```yaml
name: Rollback Deployment

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to rollback'
        required: true
        type: choice
        options:
          - staging
          - production
      version:
        description: 'Version to rollback to (commit SHA or tag)'
        required: true
        type: string

jobs:
  rollback:
    runs-on: ubuntu-latest
    environment:
      name: ${{ github.event.inputs.environment }}
    
    steps:
      - uses: actions/checkout@v3
        with:
          ref: ${{ github.event.inputs.version }}
      
      - name: Validate version
        run: |
          git rev-parse --verify ${{ github.event.inputs.version }}
      
      - name: Build image for rollback
        run: |
          docker build -t app:${{ github.event.inputs.version }} .
      
      - name: Deploy rollback
        env:
          DEPLOY_KEY: ${{ secrets[format('{0}_DEPLOY_KEY', github.event.inputs.environment)] }}
        run: |
          ./scripts/deploy.sh ${{ github.event.inputs.environment }} ${{ github.event.inputs.version }}
      
      - name: Verify rollback
        run: |
          ./scripts/health-check.sh https://${{ github.event.inputs.environment }}.example.com
      
      - name: Create rollback issue
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: `Rollback performed: ${context.payload.inputs.environment}`,
              body: `## Rollback Details\n\n- **Environment:** ${context.payload.inputs.environment}\n- **Rolled back to:** ${context.payload.inputs.version}\n- **Performed by:** @${context.actor}\n- **Reason:** Manual rollback triggered\n\n**Action Required:** Investigate why rollback was necessary.`,
              labels: ['rollback', 'incident']
            })
```

### 3. Deployment Script

**scripts/deploy.sh:**
```bash
#!/bin/bash
set -euo pipefail

ENVIRONMENT=$1
VERSION=$2

echo "🚀 Deploying version $VERSION to $ENVIRONMENT"

# Load environment-specific configuration
CONFIG_FILE="environments/${ENVIRONMENT}/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file not found: $CONFIG_FILE"
    exit 1
fi

# Parse config
APP_NAME=$(yq e '.app.name' $CONFIG_FILE)
REPLICAS=$(yq e '.app.replicas' $CONFIG_FILE)
NAMESPACE=$(yq e '.kubernetes.namespace' $CONFIG_FILE)

echo "📦 Application: $APP_NAME"
echo "🔢 Replicas: $REPLICAS"
echo "🏷️  Namespace: $NAMESPACE"

# Tag and push image
docker tag app:$VERSION ${REGISTRY}/${APP_NAME}:${VERSION}
docker push ${REGISTRY}/${APP_NAME}:${VERSION}

# Deploy to Kubernetes
kubectl set image deployment/${APP_NAME} \
    ${APP_NAME}=${REGISTRY}/${APP_NAME}:${VERSION} \
    -n ${NAMESPACE}

# Wait for rollout
kubectl rollout status deployment/${APP_NAME} -n ${NAMESPACE} --timeout=5m

echo "✅ Deployment successful!"
```

### 4. Health Check Script

**scripts/health-check.sh:**
```bash
#!/bin/bash
set -euo pipefail

URL=$1
MAX_RETRIES=30
RETRY_DELAY=10

echo "🏥 Performing health check on $URL"

for i in $(seq 1 $MAX_RETRIES); do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" ${URL}/health || echo "000")
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✅ Health check passed (HTTP $HTTP_CODE)"
        
        # Additional checks
        RESPONSE=$(curl -s ${URL}/health)
        if echo "$RESPONSE" | jq -e '.status == "healthy"' > /dev/null; then
            echo "✅ Application reports healthy status"
            exit 0
        fi
    fi
    
    echo "⏳ Attempt $i/$MAX_RETRIES: HTTP $HTTP_CODE - Retrying in ${RETRY_DELAY}s..."
    sleep $RETRY_DELAY
done

echo "❌ Health check failed after $MAX_RETRIES attempts"
exit 1
```

### 5. Environment Configurations

**environments/production/config.yaml:**
```yaml
app:
  name: library-manager
  replicas: 3
  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"

kubernetes:
  namespace: production
  ingress:
    host: example.com
    tls: true

database:
  connection_pool_size: 20
  max_connections: 100

monitoring:
  enabled: true
  alert_threshold: 90

autoscaling:
  enabled: true
  min_replicas: 3
  max_replicas: 10
  target_cpu_percent: 70
```

## Setting Up GitHub Environments

### 1. Create Environments
Go to: Settings → Environments → New environment

Create three environments:
- `dev`
- `staging`
- `production`

### 2. Configure Production Protection Rules

For the `production` environment:
- ✅ Required reviewers: Add team members
- ✅ Wait timer: 5 minutes
- ✅ Deployment branches: `main` only
- ✅ Environment secrets: Add production credentials

### 3. Add Environment Secrets

For each environment, add:
```
DEV_DEPLOY_KEY
DEV_DB_URL

STAGING_DEPLOY_KEY
STAGING_DB_URL

PROD_DEPLOY_KEY
PROD_DB_URL
SLACK_WEBHOOK
```

## Deployment Strategies

### Blue-Green Deployment
```yaml
- name: Deploy Blue environment
  run: kubectl apply -f k8s/blue-deployment.yaml

- name: Test Blue environment
  run: ./scripts/integration-test.sh blue

- name: Switch traffic to Blue
  run: kubectl patch service app -p '{"spec":{"selector":{"version":"blue"}}}'
```

### Canary Deployment
```yaml
- name: Deploy canary (10% traffic)
  run: |
    kubectl apply -f k8s/canary-deployment.yaml
    kubectl patch service app -p '{"spec":{"selector":{"version":"canary"}}}'
    
- name: Monitor canary metrics
  run: ./scripts/monitor-canary.sh

- name: Promote or rollback
  run: |
    if [ "$CANARY_HEALTHY" = "true" ]; then
      kubectl scale deployment/main --replicas=0
      kubectl scale deployment/canary --replicas=3
    else
      kubectl delete deployment/canary
    fi
```

## Best Practices

✅ Use environment protection rules for production  
✅ Require manual approvals for critical deployments  
✅ Implement comprehensive health checks  
✅ Always test in staging before production  
✅ Maintain deployment history  
✅ Have a rollback plan  
✅ Monitor deployments in real-time  
✅ Notify teams of deployment status  

## Troubleshooting

**Issue:** Approval not working  
**Solution:** Check environment protection rules and reviewer permissions

**Issue:** Secrets not available  
**Solution:** Verify secrets are set at environment level, not repository level

**Issue:** Deployment hangs  
**Solution:** Check kubectl timeout and cluster connectivity

**Issue:** Health check fails  
**Solution:** Verify application startup time and health endpoint

## Next Steps

- Implement progressive delivery with Flagger
- Add deployment analytics and DORA metrics
- Set up automated rollback on errors
- Integrate with feature flags (LaunchDarkly)
- Add deployment smoke tests

## Resources

- [GitHub Environments](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [Deployment Strategies](https://blog.container-solutions.com/deployment-strategies)
- [Progressive Delivery](https://www.weave.works/blog/progressive-delivery)
