#!/bin/bash

# Script to setup image pull secrets at namespace level for target pods
# Usage: ./setup-namespace-secrets.sh <namespace> <ecr-registry> <region> [secret-name]

NAMESPACE=$1
ECR_REGISTRY=$2
REGION=$3
SECRET_NAME=${4:-"ecr-secret"}

if [ -z "$NAMESPACE" ] || [ -z "$ECR_REGISTRY" ] || [ -z "$REGION" ]; then
    echo "Usage: $0 <namespace> <ecr-registry> <region> [secret-name]"
    echo "Example: $0 default 123456789012.dkr.ecr.us-west-2.amazonaws.com us-west-2"
    exit 1
fi

echo "Setting up image pull secrets for namespace: $NAMESPACE"

# Create namespace if it doesn't exist
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create ECR secret in the target namespace
kubectl create secret docker-registry "$SECRET_NAME" \
  --docker-server="$ECR_REGISTRY" \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region "$REGION") \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Patch default service account to use the secret
kubectl patch serviceaccount default -n "$NAMESPACE" -p "{\"imagePullSecrets\":[{\"name\":\"$SECRET_NAME\"}]}"

echo "Setup complete! All pods in namespace '$NAMESPACE' will now have access to ECR images."
echo "The default service account has been configured with image pull secret: $SECRET_NAME"
