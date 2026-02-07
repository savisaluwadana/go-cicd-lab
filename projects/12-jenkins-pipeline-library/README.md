# Project 12: Jenkins Shared Pipeline Library - Reusable CI/CD Components

## 🎯 **Learning Objectives**
- Create reusable pipeline components
- Implement custom Jenkins steps
- Share pipeline logic across teams
- Build a centralized pipeline library
- Version control pipeline code

## 📋 **Project Overview**
Build a Jenkins Shared Library that provides reusable pipeline steps, reducing code duplication across multiple projects and standardizing CI/CD practices organization-wide.

## 🏗️ **Library Structure**
```
jenkins-pipeline-library/
├── vars/
│   ├── buildDockerImage.groovy
│   ├── deployToKubernetes.groovy
│   ├── runSecurityScan.groovy
│   ├── sendSlackNotification.groovy
│   └── standardPipeline.groovy
├── src/
│   └── org/
│       └── company/
│           ├── Docker.groovy
│           ├── Kubernetes.groovy
│           └── Security.groovy
├── resources/
│   ├── templates/
│   │   ├── Dockerfile.template
│   │   └── deployment.yaml.template
│   └── scripts/
│       ├── security-scan.sh
│       └── performance-test.sh
└── README.md
```

## 🔧 **Core Library Components**

### `vars/standardPipeline.groovy`
```groovy
#!/usr/bin/env groovy

def call(Map config = [:]) {
    pipeline {
        agent any
        
        options {
            timestamps()
            timeout(time: config.timeout ?: 60, unit: 'MINUTES')
            buildDiscarder(logRotator(numToKeepStr: '10'))
        }
        
        environment {
            APP_NAME = config.appName ?: env.JOB_NAME
            BUILD_VERSION = "${env.BUILD_NUMBER}"
            DOCKER_REGISTRY = config.registry ?: 'docker.io'
        }
        
        stages {
            stage('Checkout') {
                steps {
                    checkout scm
                    script {
                        env.GIT_COMMIT_SHORT = sh(
                            script: "git rev-parse --short HEAD",
                            returnStdout: true
                        ).trim()
                    }
                }
            }
            
            stage('Build') {
                steps {
                    script {
                        if (config.buildType == 'docker') {
                            buildDockerImage(
                                imageName: env.APP_NAME,
                                tag: env.GIT_COMMIT_SHORT,
                                dockerfile: config.dockerfile ?: 'Dockerfile',
                                buildArgs: config.buildArgs ?: [:]
                            )
                        } else if (config.buildType == 'maven') {
                            sh 'mvn clean package'
                        } else if (config.buildType == 'npm') {
                            sh 'npm ci && npm run build'
                        } else if (config.buildType == 'go') {
                            sh 'go build -o app .'
                        }
                    }
                }
            }
            
            stage('Test') {
                parallel {
                    stage('Unit Tests') {
                        when {
                            expression { config.runTests != false }
                        }
                        steps {
                            script {
                                runTests(
                                    type: 'unit',
                                    command: config.testCommand ?: 'npm test'
                                )
                            }
                        }
                    }
                    
                    stage('Security Scan') {
                        when {
                            expression { config.runSecurityScan == true }
                        }
                        steps {
                            script {
                                runSecurityScan(
                                    scanType: config.securityScanType ?: 'trivy',
                                    imageName: "${env.APP_NAME}:${env.GIT_COMMIT_SHORT}"
                                )
                            }
                        }
                    }
                }
            }
            
            stage('Deploy') {
                when {
                    branch config.deployBranch ?: 'main'
                }
                steps {
                    script {
                        deployToKubernetes(
                            namespace: config.namespace ?: 'default',
                            deployment: env.APP_NAME,
                            image: "${env.DOCKER_REGISTRY}/${env.APP_NAME}:${env.GIT_COMMIT_SHORT}",
                            replicas: config.replicas ?: 3
                        )
                    }
                }
            }
        }
        
        post {
            success {
                script {
                    sendSlackNotification(
                        channel: config.slackChannel ?: '#deployments',
                        status: 'SUCCESS',
                        message: "✅ ${env.APP_NAME} deployed successfully"
                    )
                }
            }
            failure {
                script {
                    sendSlackNotification(
                        channel: config.slackChannel ?: '#deployments',
                        status: 'FAILURE',
                        message: "❌ ${env.APP_NAME} deployment failed"
                    )
                }
            }
        }
    }
}
```

### `vars/buildDockerImage.groovy`
```groovy
#!/usr/bin/env groovy

def call(Map config = [:]) {
    def imageName = config.imageName ?: error("imageName is required")
    def tag = config.tag ?: 'latest'
    def dockerfile = config.dockerfile ?: 'Dockerfile'
    def buildArgs = config.buildArgs ?: [:]
    def context = config.context ?: '.'
    def push = config.push != false
    
    // Build arguments string
    def buildArgsStr = buildArgs.collect { k, v -> "--build-arg ${k}=${v}" }.join(' ')
    
    echo "🐳 Building Docker image: ${imageName}:${tag}"
    
    sh """
        docker build ${buildArgsStr} \
            -t ${imageName}:${tag} \
            -f ${dockerfile} \
            ${context}
    """
    
    if (push) {
        echo "📤 Pushing Docker image to registry"
        withCredentials([usernamePassword(
            credentialsId: 'docker-registry-creds',
            usernameVariable: 'DOCKER_USER',
            passwordVariable: 'DOCKER_PASS'
        )]) {
            sh """
                echo "\${DOCKER_PASS}" | docker login -u "\${DOCKER_USER}" --password-stdin
                docker push ${imageName}:${tag}
            """
        }
    }
    
    return "${imageName}:${tag}"
}
```

### `vars/deployToKubernetes.groovy`
```groovy
#!/usr/bin/env groovy

def call(Map config = [:]) {
    def namespace = config.namespace ?: 'default'
    def deployment = config.deployment ?: error("deployment name is required")
    def image = config.image ?: error("image is required")
    def replicas = config.replicas ?: 3
    def healthCheckPath = config.healthCheckPath ?: '/health'
    
    echo "☸️  Deploying to Kubernetes: ${deployment} in ${namespace}"
    
    withCredentials([file(credentialsId: 'kubeconfig', variable: 'KUBECONFIG')]) {
        // Update deployment image
        sh """
            kubectl set image deployment/${deployment} \
                ${deployment}=${image} \
                -n ${namespace} \
                --record
        """
        
        // Scale deployment
        sh """
            kubectl scale deployment/${deployment} \
                --replicas=${replicas} \
                -n ${namespace}
        """
        
        // Wait for rollout
        sh """
            kubectl rollout status deployment/${deployment} \
                -n ${namespace} \
                --timeout=5m
        """
        
        // Verify deployment
        def podStatus = sh(
            script: """
                kubectl get pods -n ${namespace} \
                    -l app=${deployment} \
                    -o jsonpath='{.items[*].status.phase}'
            """,
            returnStdout: true
        ).trim()
        
        if (!podStatus.contains('Running')) {
            error("Deployment failed: Pods are not running")
        }
        
        echo "✅ Deployment successful: ${deployment}"
    }
}
```

### `vars/runSecurityScan.groovy`
```groovy
#!/usr/bin/env groovy

def call(Map config = [:]) {
    def scanType = config.scanType ?: 'trivy'
    def imageName = config.imageName
    def severity = config.severity ?: 'CRITICAL,HIGH'
    def exitCode = config.exitCode ?: 0
    
    echo "🔒 Running security scan with ${scanType}"
    
    switch(scanType) {
        case 'trivy':
            sh """
                trivy image \
                    --severity ${severity} \
                    --exit-code ${exitCode} \
                    --no-progress \
                    ${imageName}
            """
            break
            
        case 'snyk':
            withCredentials([string(credentialsId: 'snyk-token', variable: 'SNYK_TOKEN')]) {
                sh """
                    snyk container test ${imageName} \
                        --severity-threshold=high \
                        --json > snyk-results.json
                """
            }
            break
            
        case 'grype':
            sh """
                grype ${imageName} \
                    -o json \
                    --fail-on ${severity.toLowerCase()} \
                    > grype-results.json
            """
            break
            
        default:
            error("Unknown scan type: ${scanType}")
    }
    
    // Archive results
    archiveArtifacts artifacts: '*-results.json', allowEmptyArchive: true
}
```

### `vars/sendSlackNotification.groovy`
```groovy
#!/usr/bin/env groovy

def call(Map config = [:]) {
    def channel = config.channel ?: '#general'
    def status = config.status ?: 'INFO'
    def message = config.message ?: 'Pipeline notification'
    
    def color = status == 'SUCCESS' ? 'good' : 
                status == 'FAILURE' ? 'danger' : 'warning'
    
    def emoji = status == 'SUCCESS' ? ':white_check_mark:' :
                status == 'FAILURE' ? ':x:' : ':warning:'
    
    def payload = [
        channel: channel,
        username: 'Jenkins CI',
        icon_emoji: ':jenkins:',
        attachments: [[
            color: color,
            title: "${emoji} ${env.JOB_NAME} - Build #${env.BUILD_NUMBER}",
            text: message,
            fields: [
                [title: 'Branch', value: env.BRANCH_NAME ?: 'N/A', short: true],
                [title: 'Status', value: status, short: true],
                [title: 'Duration', value: currentBuild.durationString, short: true]
            ],
            footer: 'Jenkins CI',
            footer_icon: 'https://jenkins.io/favicon.ico',
            ts: System.currentTimeMillis() / 1000
        ]]
    ]
    
    withCredentials([string(credentialsId: 'slack-webhook', variable: 'SLACK_WEBHOOK')]) {
        sh """
            curl -X POST \${SLACK_WEBHOOK} \
                -H 'Content-Type: application/json' \
                -d '${groovy.json.JsonOutput.toJson(payload)}'
        """
    }
}
```

## 🔧 **Using the Shared Library**

### Jenkins Configuration
```groovy
// In Jenkins > Manage Jenkins > Configure System > Global Pipeline Libraries
Name: company-pipeline-library
Default version: main
Retrieval method: Modern SCM
Source Code Management: Git
Repository URL: https://github.com/your-org/jenkins-pipeline-library.git
```

### Simple Jenkinsfile (Using Library)
```groovy
@Library('company-pipeline-library@main') _

standardPipeline(
    appName: 'my-microservice',
    buildType: 'docker',
    testCommand: 'go test -v ./...',
    runSecurityScan: true,
    namespace: 'production',
    replicas: 5,
    deployBranch: 'main',
    slackChannel: '#deployments'
)
```

### Custom Jenkinsfile (Advanced)
```groovy
@Library('company-pipeline-library@main') _

pipeline {
    agent any
    
    stages {
        stage('Build') {
            steps {
                script {
                    def imageTag = buildDockerImage(
                        imageName: 'my-app',
                        tag: "${env.BUILD_NUMBER}",
                        buildArgs: [
                            GO_VERSION: '1.21',
                            APP_ENV: 'production'
                        ]
                    )
                    env.IMAGE_TAG = imageTag
                }
            }
        }
        
        stage('Security Scan') {
            steps {
                runSecurityScan(
                    scanType: 'trivy',
                    imageName: env.IMAGE_TAG,
                    severity: 'CRITICAL,HIGH,MEDIUM'
                )
            }
        }
        
        stage('Deploy') {
            steps {
                deployToKubernetes(
                    namespace: 'production',
                    deployment: 'my-app',
                    image: env.IMAGE_TAG,
                    replicas: 10
                )
            }
        }
    }
    
    post {
        always {
            sendSlackNotification(
                status: currentBuild.result,
                message: "Deployment ${currentBuild.result}"
            )
        }
    }
}
```

## 🧪 **Testing the Library**

### Unit Tests for Groovy Code
```groovy
// test/groovy/BuildDockerImageSpec.groovy
import spock.lang.*

class BuildDockerImageSpec extends Specification {
    def "should build docker image with correct tag"() {
        given:
        def config = [
            imageName: 'test-app',
            tag: 'v1.0.0'
        ]
        
        when:
        def result = buildDockerImage(config)
        
        then:
        result == 'test-app:v1.0.0'
    }
}
```

## 🎯 **Key Learnings**
- ✅ Create reusable pipeline components
- ✅ Implement custom Jenkins steps
- ✅ Version control pipeline logic
- ✅ Standardize CI/CD across teams
- ✅ Reduce code duplication

## 📊 **Benefits**
- **Code Reuse**: 80% reduction in Jenkinsfile code
- **Consistency**: Standardized deployments across 50+ projects
- **Maintenance**: Single source of truth for pipeline logic
- **Onboarding**: New teams can deploy in minutes

## 🚀 **Best Practices**
1. Version your library (use tags)
2. Document all functions
3. Keep functions focused and small
4. Provide sensible defaults
5. Make parameters configurable
6. Add error handling
7. Write unit tests

## 🔍 **Troubleshooting**
- **Issue**: Library not loading
  - **Solution**: Check library configuration in Jenkins settings
- **Issue**: Function not found
  - **Solution**: Ensure function is in `vars/` directory with `.groovy` extension
- **Issue**: Credentials not found
  - **Solution**: Verify credential IDs match Jenkins credential store

## 📚 **Additional Resources**
- [Jenkins Shared Libraries](https://www.jenkins.io/doc/book/pipeline/shared-libraries/)
- [Pipeline Development Tools](https://www.jenkins.io/doc/book/pipeline/development/)
- [Groovy Documentation](https://groovy-lang.org/documentation.html)
