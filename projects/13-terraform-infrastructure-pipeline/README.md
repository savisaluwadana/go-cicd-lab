# Project 13: Terraform Infrastructure Pipeline - GitOps for Infrastructure

## 🎯 **Learning Objectives**
- Implement Infrastructure as Code (IaC) with Terraform
- Build automated infrastructure deployment pipelines
- Implement Terraform state management
- Use GitOps principles for infrastructure
- Implement drift detection and remediation

## 📋 **Project Overview**
Create a complete CI/CD pipeline for managing infrastructure using Terraform, including automated testing, security scanning, and multi-environment deployments with approval gates.

## 🏗️ **Repository Structure**
```
terraform-infrastructure/
├── .github/
│   └── workflows/
│       ├── terraform-plan.yml
│       ├── terraform-apply.yml
│       └── drift-detection.yml
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   ├── staging/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   └── production/
│       ├── main.tf
│       ├── variables.tf
│       └── terraform.tfvars
├── modules/
│   ├── eks-cluster/
│   ├── vpc/
│   ├── rds/
│   └── s3-bucket/
├── policies/
│   ├── sentinel/
│   └── opa/
└── scripts/
    ├── validate-terraform.sh
    └── cost-estimate.sh
```

## 🔧 **Terraform Plan Workflow**

### `.github/workflows/terraform-plan.yml`
```yaml
name: Terraform Plan

on:
  pull_request:
    branches: [main]
    paths:
      - 'environments/**'
      - 'modules/**'
      - '.github/workflows/terraform-plan.yml'

env:
  TF_VERSION: '1.6.0'
  AWS_REGION: 'us-east-1'

jobs:
  terraform-plan:
    name: Plan - ${{ matrix.environment }}
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
      id-token: write
    
    strategy:
      matrix:
        environment: [dev, staging, production]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        id: fmt
        working-directory: environments/${{ matrix.environment }}
        run: terraform fmt -check -recursive
        continue-on-error: true

      - name: Terraform Init
        id: init
        working-directory: environments/${{ matrix.environment }}
        run: |
          terraform init \
            -backend-config="bucket=${{ secrets.TF_STATE_BUCKET }}" \
            -backend-config="key=${{ matrix.environment }}/terraform.tfstate" \
            -backend-config="region=${{ env.AWS_REGION }}"

      - name: Terraform Validate
        id: validate
        working-directory: environments/${{ matrix.environment }}
        run: terraform validate -no-color

      - name: Run TFLint
        uses: terraform-linters/setup-tflint@v4
        with:
          tflint_version: latest
      
      - name: Initialize TFLint
        working-directory: environments/${{ matrix.environment }}
        run: tflint --init

      - name: Run TFLint
        id: tflint
        working-directory: environments/${{ matrix.environment }}
        run: tflint -f compact

      - name: Run Checkov Security Scan
        id: checkov
        uses: bridgecrewio/checkov-action@v12
        with:
          directory: environments/${{ matrix.environment }}
          framework: terraform
          output_format: cli
          soft_fail: true

      - name: Terraform Plan
        id: plan
        working-directory: environments/${{ matrix.environment }}
        run: |
          terraform plan -no-color -out=tfplan
          terraform show -no-color tfplan > plan.txt
        continue-on-error: true

      - name: Cost Estimation with Infracost
        uses: infracost/actions/setup@v2
        with:
          api-key: ${{ secrets.INFRACOST_API_KEY }}

      - name: Generate Infracost JSON
        working-directory: environments/${{ matrix.environment }}
        run: |
          infracost breakdown --path tfplan \
            --format json \
            --out-file infracost.json

      - name: Post Infracost Comment
        uses: infracost/actions/comment@v1
        with:
          path: environments/${{ matrix.environment }}/infracost.json
          behavior: update

      - name: Upload Plan Artifact
        uses: actions/upload-artifact@v4
        with:
          name: tfplan-${{ matrix.environment }}
          path: |
            environments/${{ matrix.environment }}/tfplan
            environments/${{ matrix.environment }}/plan.txt

      - name: Comment PR with Plan
        uses: actions/github-script@v7
        if: github.event_name == 'pull_request'
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fs = require('fs');
            const plan = fs.readFileSync('environments/${{ matrix.environment }}/plan.txt', 'utf8');
            const truncatedPlan = plan.length > 65000 ? plan.substring(0, 65000) + '\n...(truncated)' : plan;
            
            const output = `### Terraform Plan - ${{ matrix.environment }}
            
            #### Format and Style 🖌\`${{ steps.fmt.outcome }}\`
            #### Initialization ⚙️\`${{ steps.init.outcome }}\`
            #### Validation 🤖\`${{ steps.validate.outcome }}\`
            #### TFLint 🔍\`${{ steps.tflint.outcome }}\`
            #### Checkov Security 🔒\`${{ steps.checkov.outcome }}\`
            #### Plan 📖\`${{ steps.plan.outcome }}\`
            
            <details><summary>Show Plan</summary>
            
            \`\`\`terraform
            ${truncatedPlan}
            \`\`\`
            
            </details>
            
            *Pusher: @${{ github.actor }}, Action: \`${{ github.event_name }}\`*`;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            });

      - name: Terraform Plan Status
        if: steps.plan.outcome == 'failure'
        run: exit 1
```

## 🚀 **Terraform Apply Workflow**

### `.github/workflows/terraform-apply.yml`
```yaml
name: Terraform Apply

on:
  push:
    branches: [main]
    paths:
      - 'environments/**'
      - 'modules/**'
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
  terraform-apply-dev:
    name: Apply - Dev
    if: github.event_name == 'push' || github.event.inputs.environment == 'dev'
    runs-on: ubuntu-latest
    environment: development
    permissions:
      id-token: write
      contents: read
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN_DEV }}
          aws-region: us-east-1

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.6.0

      - name: Terraform Init
        working-directory: environments/dev
        run: |
          terraform init \
            -backend-config="bucket=${{ secrets.TF_STATE_BUCKET }}" \
            -backend-config="key=dev/terraform.tfstate"

      - name: Terraform Apply
        working-directory: environments/dev
        run: terraform apply -auto-approve

      - name: Extract Outputs
        id: outputs
        working-directory: environments/dev
        run: |
          echo "cluster_endpoint=$(terraform output -raw cluster_endpoint)" >> $GITHUB_OUTPUT
          echo "vpc_id=$(terraform output -raw vpc_id)" >> $GITHUB_OUTPUT

      - name: Update Parameter Store
        run: |
          aws ssm put-parameter \
            --name "/dev/cluster_endpoint" \
            --value "${{ steps.outputs.outputs.cluster_endpoint }}" \
            --type String \
            --overwrite

  terraform-apply-staging:
    name: Apply - Staging
    if: github.event.inputs.environment == 'staging'
    runs-on: ubuntu-latest
    environment: staging
    needs: [terraform-apply-dev]
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN_STAGING }}
          aws-region: us-east-1

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Init & Apply
        working-directory: environments/staging
        run: |
          terraform init -backend-config="bucket=${{ secrets.TF_STATE_BUCKET }}" \
                        -backend-config="key=staging/terraform.tfstate"
          terraform apply -auto-approve

  terraform-apply-production:
    name: Apply - Production
    if: github.event.inputs.environment == 'production'
    runs-on: ubuntu-latest
    environment: production
    needs: [terraform-apply-staging]
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Manual Approval Check
        uses: trstringer/manual-approval@v1
        with:
          secret: ${{ github.TOKEN }}
          approvers: platform-team,devops-leads
          minimum-approvals: 2

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN_PROD }}
          aws-region: us-east-1

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Init & Apply
        working-directory: environments/production
        run: |
          terraform init -backend-config="bucket=${{ secrets.TF_STATE_BUCKET }}" \
                        -backend-config="key=production/terraform.tfstate"
          terraform apply -auto-approve

      - name: Notify Deployment
        uses: slackapi/slack-github-action@v1
        with:
          channel-id: 'infrastructure'
          slack-message: "✅ Production infrastructure updated successfully!"
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

## 🔍 **Drift Detection Workflow**

### `.github/workflows/drift-detection.yml`
```yaml
name: Terraform Drift Detection

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:

jobs:
  detect-drift:
    name: Detect Drift - ${{ matrix.environment }}
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [dev, staging, production]
    
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Init
        working-directory: environments/${{ matrix.environment }}
        run: |
          terraform init -backend-config="bucket=${{ secrets.TF_STATE_BUCKET }}" \
                        -backend-config="key=${{ matrix.environment }}/terraform.tfstate"

      - name: Terraform Plan (Drift Check)
        id: plan
        working-directory: environments/${{ matrix.environment }}
        run: |
          terraform plan -detailed-exitcode -no-color > plan_output.txt
          exit_code=$?
          echo "exit_code=$exit_code" >> $GITHUB_OUTPUT
          cat plan_output.txt
        continue-on-error: true

      - name: Create Issue on Drift
        if: steps.plan.outputs.exit_code == '2'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const plan = fs.readFileSync('environments/${{ matrix.environment }}/plan_output.txt', 'utf8');
            
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: '🚨 Infrastructure Drift Detected - ${{ matrix.environment }}',
              body: `### Drift Detection Alert
              
              Infrastructure drift has been detected in the **${{ matrix.environment }}** environment.
              
              <details><summary>Terraform Plan Output</summary>
              
              \`\`\`terraform
              ${plan}
              \`\`\`
              
              </details>
              
              **Action Required**: Review the changes and either:
              1. Apply the Terraform configuration to remediate drift
              2. Update the Terraform code to match the current infrastructure state
              
              **Detection Time**: ${new Date().toISOString()}
              `,
              labels: ['infrastructure', 'drift', '${{ matrix.environment }}']
            });

      - name: Send Slack Alert
        if: steps.plan.outputs.exit_code == '2'
        uses: slackapi/slack-github-action@v1
        with:
          channel-id: 'infrastructure-alerts'
          slack-message: |
            🚨 *Infrastructure Drift Detected*
            Environment: ${{ matrix.environment }}
            Time: $(date)
            Review: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

## 📝 **Example Terraform Module**

### `modules/eks-cluster/main.tf`
```hcl
terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster.arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = var.enable_public_access
    security_group_ids      = [aws_security_group.cluster.id]
  }

  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }

  enabled_cluster_log_types = [
    "api",
    "audit",
    "authenticator",
    "controllerManager",
    "scheduler"
  ]

  tags = merge(
    var.tags,
    {
      Name        = var.cluster_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  )
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.cluster_name}-workers"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids

  scaling_config {
    desired_size = var.desired_size
    max_size     = var.max_size
    min_size     = var.min_size
  }

  instance_types = var.instance_types
  capacity_type  = var.capacity_type

  update_config {
    max_unavailable_percentage = 25
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [scaling_config[0].desired_size]
  }
}
```

## 🎯 **Key Learnings**
- ✅ Infrastructure as Code with Terraform
- ✅ Automated security scanning (Checkov)
- ✅ Cost estimation with Infracost
- ✅ Drift detection and remediation
- ✅ GitOps workflows for infrastructure
- ✅ Multi-environment deployments
- ✅ State management best practices

## 📊 **Pipeline Features**
- **Security**: Checkov scans, TFLint validation
- **Cost Control**: Infracost estimates on every PR
- **Safety**: Plan preview before apply
- **Compliance**: Automated drift detection
- **Visibility**: PR comments with detailed plans

## 🚀 **Best Practices**
1. Always use remote state (S3 + DynamoDB)
2. Enable state locking
3. Use workspaces or separate state files per environment
4. Implement RBAC for Terraform operations
5. Version pin all providers
6. Use modules for reusable components
7. Tag all resources consistently

## 📚 **Additional Resources**
- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [Checkov Documentation](https://www.checkov.io/)
- [Infracost Guide](https://www.infracost.io/docs/)
