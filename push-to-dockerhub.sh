#!/bin/bash

# Check if Docker Hub username is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <dockerhub-username>"
    echo "Example: $0 myusername"
    exit 1
fi

DOCKER_HUB_USER=$1
IMAGE_NAME="$DOCKER_HUB_USER/circuit-breaker-controller:latest"

echo "🐳 Building Docker image for Docker Hub..."
docker build -t $IMAGE_NAME .

echo "📤 Pushing to Docker Hub..."
docker push $IMAGE_NAME

echo "✅ Successfully pushed $IMAGE_NAME to Docker Hub"

# Update Helm values
echo "🔧 Updating Helm values..."
sed -i '' "s|repository: .*|repository: $IMAGE_NAME|g" helm/circuit-breaker-controller/values.yaml

echo "🎯 Updated Helm chart to use: $IMAGE_NAME"
echo "📋 To deploy: make helm-uninstall && make helm-install"