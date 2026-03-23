pipeline {
    agent any

    environment {
        AWS_REGION     = 'us-east-1'
        AWS_ACCOUNT_ID = '100984278168'
        ECR_REPO       = "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/movievault/inventory-service"
        IMAGE_TAG      = "${env.GIT_COMMIT[0..7]}"
        SONAR_URL      = 'http://3.237.177.157:9000'
    }

    stages {

        stage('Checkout') {
            steps {
                checkout scm
                echo "Building commit: ${env.GIT_COMMIT}"
            }
        }

    stage('Unit Tests') {
     steps {
         sh '''
             export GOROOT=/usr/local/go
             export PATH=$GOROOT/bin:$PATH
             export HOME=/tmp
             export GOPATH=/tmp/gopath
             export GOCACHE=/tmp/gocache
             mkdir -p /tmp/gopath /tmp/gocache
             
             go version
             go env GOROOT GOTOOLDIR
             
             go mod tidy
             go test ./... -v -coverprofile=coverage.out
         '''
     }
    post {
        always {
            archiveArtifacts artifacts: 'coverage.out',
                             allowEmptyArchive: true
        }
    }
}
      stage('SonarQube Analysis') {
    steps {
        withSonarQubeEnv('sonarqube') {
            sh '''
                export GOROOT=/usr/local/go
                export PATH=$GOROOT/bin:/opt/sonar-scanner/bin:$PATH
                export HOME=/tmp
                export GOPATH=/tmp/gopath
                export GOCACHE=/tmp/gocache
                sonar-scanner \
                  -Dsonar.projectKey=movievault-inventory \
                  -Dsonar.projectName=movievault-inventory \
                  -Dsonar.sources=. \
                  -Dsonar.go.coverage.reportPaths=coverage.out \
                  -Dsonar.host.url=http://3.237.177.157:9000
            '''
        }
    }
}

        stage('Quality Gate') {
            steps {
                timeout(time: 5, unit: 'MINUTES') {
                    waitForQualityGate abortPipeline: true
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                sh "docker build -t movievault-inventory:${IMAGE_TAG} ."
                echo "Image built: movievault-inventory:${IMAGE_TAG}"
            }
        }

        stage('Trivy Security Scan') {
            steps {
                sh '''
                    trivy image \
                      --exit-code 0 \
                      --severity CRITICAL,HIGH \
                      --no-progress \
                      movievault-inventory:${IMAGE_TAG}
                '''
                // exit-code 0 means report but don't fail yet
                // change to 1 when ready to enforce blocking
            }
        }

        stage('Push to ECR') {
            steps {
                withCredentials([[
                    $class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId: 'aws-credentials'
                ]]) {
                    sh '''
                        aws ecr get-login-password --region ${AWS_REGION} | \
                        docker login --username AWS --password-stdin \
                        ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

                        docker tag movievault-inventory:${IMAGE_TAG} ${ECR_REPO}:${IMAGE_TAG}
                        docker tag movievault-inventory:${IMAGE_TAG} ${ECR_REPO}:latest

                        docker push ${ECR_REPO}:${IMAGE_TAG}
                        docker push ${ECR_REPO}:latest

                        echo "Pushed ${ECR_REPO}:${IMAGE_TAG}"
                    '''
                }
            }
        }

        stage('Update Manifests') {
            steps {
                withCredentials([string(credentialsId: 'github-token', variable: 'GH_TOKEN')]) {
                    sh '''
                        rm -rf movievault-manifests
                        git clone https://${GH_TOKEN}@github.com/cliffdoyle/movievault-manifests

                        cd movievault-manifests

                        sed -i "s|inventory-service:.*|inventory-service:${IMAGE_TAG}|g" \
                          k8s/03-inventory-service.yaml

                        git config user.email "jenkins@movievault.com"
                        git config user.name "Jenkins CI"
                        git add .
                        git diff --cached --quiet || git commit -m "ci: inventory-service → ${IMAGE_TAG}"
                        git push

                        echo "Manifests updated — Argo CD will deploy within 3 minutes"
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "Pipeline PASSED — ${IMAGE_TAG} deployed via Argo CD"
        }
        failure {
            echo "Pipeline FAILED — check the stage above"
        }
        always {
            sh "docker rmi movievault-inventory:${IMAGE_TAG} || true"
        }
    }
}