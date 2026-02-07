# Project 4: Declarative Jenkins Pipeline

## Overview
Build a complete Jenkins declarative pipeline with multiple stages, parallel execution, post-build actions, and artifact management.

## What You'll Learn
- Jenkins declarative syntax
- Pipeline stages and steps
- Parallel execution
- Credentials management
- Artifact archiving
- Email notifications

## Project Structure
```
04-jenkins-declarative/
├── Jenkinsfile
├── jenkins/
│   ├── docker-compose.yaml
│   └── plugins.txt
├── app/
│   ├── main.go
│   ├── main_test.go
│   └── Dockerfile
└── README.md
```

## Implementation

### 1. Jenkinsfile

**Jenkinsfile:**
```groovy
pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'docker.io'
        DOCKER_CREDENTIALS_ID = 'dockerhub-credentials'
        APP_NAME = 'library-manager'
        GO_VERSION = '1.19'
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timestamps()
        timeout(time: 1, unit: 'HOURS')
        disableConcurrentBuilds()
    }
    
    parameters {
        choice(name: 'ENVIRONMENT', choices: ['dev', 'staging', 'production'], description: 'Target environment')
        booleanParam(name: 'RUN_TESTS', defaultValue: true, description: 'Run tests')
        string(name: 'DOCKER_TAG', defaultValue: 'latest', description: 'Docker image tag')
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo "🔍 Checking out code..."
                checkout scm
                script {
                    env.GIT_COMMIT_SHORT = sh(
                        script: "git rev-parse --short HEAD",
                        returnStdout: true
                    ).trim()
                    env.BUILD_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT_SHORT}"
                }
            }
        }
        
        stage('Setup') {
            steps {
                echo "🔧 Setting up environment..."
                sh '''
                    go version
                    go env
                    mkdir -p bin coverage
                '''
            }
        }
        
        stage('Dependencies') {
            steps {
                echo "📦 Downloading dependencies..."
                sh '''
                    cd app
                    go mod download
                    go mod verify
                '''
            }
        }
        
        stage('Lint') {
            steps {
                echo "🔍 Running linters..."
                sh '''
                    cd app
                    go fmt ./...
                    go vet ./...
                    
                    # Install golangci-lint if not present
                    if ! command -v golangci-lint &> /dev/null; then
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                            sh -s -- -b $(go env GOPATH)/bin v1.54.2
                    fi
                    
                    golangci-lint run --timeout=5m
                '''
            }
        }
        
        stage('Test') {
            when {
                expression { params.RUN_TESTS == true }
            }
            parallel {
                stage('Unit Tests') {
                    steps {
                        echo "🧪 Running unit tests..."
                        sh '''
                            cd app
                            go test -v -race -coverprofile=../coverage/unit.out ./...
                            go tool cover -html=../coverage/unit.out -o ../coverage/unit.html
                        '''
                    }
                }
                
                stage('Integration Tests') {
                    steps {
                        echo "🔗 Running integration tests..."
                        sh '''
                            cd app
                            go test -v -tags=integration -coverprofile=../coverage/integration.out ./...
                        '''
                    }
                }
            }
            post {
                always {
                    junit testResults: '**/test-results/*.xml', allowEmptyResults: true
                    publishHTML([
                        reportDir: 'coverage',
                        reportFiles: 'unit.html',
                        reportName: 'Coverage Report'
                    ])
                }
            }
        }
        
        stage('Build') {
            steps {
                echo "🏗️ Building application..."
                sh '''
                    cd app
                    CGO_ENABLED=0 go build -v -o ../bin/${APP_NAME} \
                        -ldflags="-X main.version=${BUILD_TAG} -X main.commit=${GIT_COMMIT_SHORT}"
                '''
            }
        }
        
        stage('Docker Build') {
            steps {
                echo "🐳 Building Docker image..."
                script {
                    dockerImage = docker.build(
                        "${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_TAG}",
                        "-f app/Dockerfile ."
                    )
                }
            }
        }
        
        stage('Security Scan') {
            steps {
                echo "🔒 Scanning for vulnerabilities..."
                sh '''
                    # Install Trivy if not present
                    if ! command -v trivy &> /dev/null; then
                        wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
                        echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | \
                            sudo tee -a /etc/apt/sources.list.d/trivy.list
                        sudo apt-get update
                        sudo apt-get install trivy
                    fi
                    
                    trivy image --severity HIGH,CRITICAL \
                        --format json \
                        --output trivy-report.json \
                        ${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_TAG}
                '''
            }
            post {
                always {
                    archiveArtifacts artifacts: 'trivy-report.json', allowEmptyArchive: true
                }
            }
        }
        
        stage('Push Image') {
            when {
                branch 'main'
            }
            steps {
                echo "📤 Pushing Docker image..."
                script {
                    docker.withRegistry('https://' + DOCKER_REGISTRY, DOCKER_CREDENTIALS_ID) {
                        dockerImage.push(BUILD_TAG)
                        dockerImage.push('latest')
                    }
                }
            }
        }
        
        stage('Deploy') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            steps {
                echo "🚀 Deploying to ${params.ENVIRONMENT}..."
                script {
                    def deployEnv = params.ENVIRONMENT ?: (env.BRANCH_NAME == 'main' ? 'staging' : 'dev')
                    
                    withCredentials([
                        file(credentialsId: 'kubeconfig', variable: 'KUBECONFIG'),
                        string(credentialsId: "${deployEnv}-db-url", variable: 'DB_URL')
                    ]) {
                        sh """
                            kubectl set image deployment/${APP_NAME} \
                                ${APP_NAME}=${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_TAG} \
                                -n ${deployEnv}
                            
                            kubectl rollout status deployment/${APP_NAME} -n ${deployEnv} --timeout=5m
                        """
                    }
                }
            }
        }
        
        stage('Smoke Tests') {
            when {
                branch 'main'
            }
            steps {
                echo "💨 Running smoke tests..."
                sh '''
                    # Wait for deployment to stabilize
                    sleep 10
                    
                    # Run smoke tests
                    curl -f https://${ENVIRONMENT}.example.com/health || exit 1
                    
                    # Run API tests
                    newman run tests/api-tests.json \
                        --env-var base_url=https://${ENVIRONMENT}.example.com \
                        --reporters cli,junit \
                        --reporter-junit-export test-results/newman.xml
                '''
            }
        }
    }
    
    post {
        always {
            echo '🧹 Cleaning up...'
            cleanWs()
        }
        
        success {
            echo '✅ Pipeline succeeded!'
            emailext(
                subject: "✅ Build Successful: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                body: """
                    <h2>Build Successful</h2>
                    <p><strong>Job:</strong> ${env.JOB_NAME}</p>
                    <p><strong>Build Number:</strong> ${env.BUILD_NUMBER}</p>
                    <p><strong>Build Tag:</strong> ${BUILD_TAG}</p>
                    <p><strong>Commit:</strong> ${GIT_COMMIT_SHORT}</p>
                    <p><a href="${env.BUILD_URL}">View Build</a></p>
                """,
                to: '${DEFAULT_RECIPIENTS}',
                mimeType: 'text/html'
            )
        }
        
        failure {
            echo '❌ Pipeline failed!'
            emailext(
                subject: "❌ Build Failed: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                body: """
                    <h2>Build Failed</h2>
                    <p><strong>Job:</strong> ${env.JOB_NAME}</p>
                    <p><strong>Build Number:</strong> ${env.BUILD_NUMBER}</p>
                    <p><strong>Failed Stage:</strong> ${env.STAGE_NAME}</p>
                    <p><a href="${env.BUILD_URL}console">View Console Output</a></p>
                """,
                to: '${DEFAULT_RECIPIENTS}',
                mimeType: 'text/html'
            )
        }
        
        unstable {
            echo '⚠️ Pipeline unstable!'
        }
    }
}
```

### 2. Jenkins Setup with Docker Compose

**jenkins/docker-compose.yaml:**
```yaml
version: '3.8'

services:
  jenkins:
    image: jenkins/jenkins:lts
    container_name: jenkins
    privileged: true
    user: root
    ports:
      - "8080:8080"
      - "50000:50000"
    volumes:
      - jenkins_home:/var/jenkins_home
      - /var/run/docker.sock:/var/run/docker.sock
      - ./plugins.txt:/usr/share/jenkins/ref/plugins.txt
    environment:
      - JAVA_OPTS=-Djenkins.install.runSetupWizard=false
      - CASC_JENKINS_CONFIG=/var/jenkins_home/casc.yaml
    restart: unless-stopped

  jenkins-agent:
    image: jenkins/inbound-agent
    container_name: jenkins-agent
    depends_on:
      - jenkins
    environment:
      - JENKINS_URL=http://jenkins:8080
      - JENKINS_AGENT_NAME=docker-agent
      - JENKINS_SECRET=${JENKINS_AGENT_SECRET}
    restart: unless-stopped

volumes:
  jenkins_home:
    driver: local
```

### 3. Required Plugins

**jenkins/plugins.txt:**
```
workflow-aggregator:latest
docker-workflow:latest
kubernetes:latest
git:latest
github:latest
email-ext:latest
htmlpublisher:latest
junit:latest
jacoco:latest
sonar:latest
blueocean:latest
pipeline-stage-view:latest
credentials-binding:latest
timestamper:latest
ws-cleanup:latest
```

## Setting Up Jenkins

### 1. Start Jenkins
```bash
cd jenkins
docker-compose up -d

# Get initial admin password
docker exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword
```

### 2. Install Plugins
```bash
# Install plugins from file
docker exec jenkins jenkins-plugin-cli --plugin-file /usr/share/jenkins/ref/plugins.txt
docker restart jenkins
```

### 3. Configure Credentials
Navigate to: Manage Jenkins → Credentials

Add:
- **dockerhub-credentials**: Username/Password for Docker Hub
- **kubeconfig**: Secret file for Kubernetes
- **dev-db-url**, **staging-db-url**, **production-db-url**: Secret text

### 4. Create Pipeline Job
1. New Item → Pipeline
2. Configure → Pipeline Definition: Pipeline script from SCM
3. SCM: Git
4. Repository URL: Your repo URL
5. Script Path: `Jenkinsfile`

## Advanced Features

### Shared Libraries

**vars/deployApp.groovy:**
```groovy
def call(Map config) {
    withCredentials([file(credentialsId: 'kubeconfig', variable: 'KUBECONFIG')]) {
        sh """
            kubectl set image deployment/${config.appName} \
                ${config.appName}=${config.image} \
                -n ${config.namespace}
            kubectl rollout status deployment/${config.appName} -n ${config.namespace}
        """
    }
}
```

**Usage in Jenkinsfile:**
```groovy
@Library('shared-library') _

deployApp(
    appName: 'library-manager',
    image: "${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_TAG}",
    namespace: 'production'
)
```

### Matrix Builds

```groovy
matrix {
    axes {
        axis {
            name 'GO_VERSION'
            values '1.19', '1.20', '1.21'
        }
        axis {
            name 'PLATFORM'
            values 'linux', 'darwin'
        }
    }
    stages {
        stage('Build') {
            steps {
                sh "GOOS=${PLATFORM} go build ."
            }
        }
    }
}
```

## Best Practices

✅ Use declarative pipeline syntax for clarity  
✅ Implement proper error handling  
✅ Archive artifacts and test results  
✅ Use credentials securely  
✅ Implement timeouts to prevent hanging builds  
✅ Clean workspace after builds  
✅ Use shared libraries for reusable code  
✅ Implement proper notifications  

## Troubleshooting

**Issue:** Docker commands fail  
**Solution:** Ensure Jenkins container has access to Docker socket

**Issue:** kubectl not found  
**Solution:** Install kubectl in Jenkins container or use agent with kubectl

**Issue:** Tests not showing in UI  
**Solution:** Ensure JUnit plugin is installed and test results path is correct

## Next Steps

- Implement custom shared libraries
- Add SonarQube integration
- Set up Jenkins Configuration as Code (JCasC)
- Implement dynamic agent provisioning
- Add performance testing stage

## Resources

- [Jenkins Pipeline Syntax](https://www.jenkins.io/doc/book/pipeline/syntax/)
- [Jenkins Plugins](https://plugins.jenkins.io/)
- [Jenkins Best Practices](https://www.jenkins.io/doc/book/pipeline/pipeline-best-practices/)
