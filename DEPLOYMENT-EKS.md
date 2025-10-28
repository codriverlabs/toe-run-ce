# TOE Operator - EKS Deployment Guide

This guide describes how to deploy the TOE (Tactical Operations Engine) operator on Amazon EKS using ECR for container images.

## Prerequisites

- EKS cluster with Kubernetes v1.33+
- `kubectl` configured to access your EKS cluster
- `helm` v3.0+ installed
- AWS CLI configured with appropriate permissions
- cert-manager installed in your EKS cluster

### Install cert-manager (if not already installed)

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true

# Wait for cert-manager to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s
```

## Deployment Steps

### 1. Download and Extract Release

Download the latest release from GitHub and extract the Helm chart:

```bash
# Download the release (replace with actual version)
wget https://github.com/codriverlabs/toe/releases/download/v1.1.0/toe-operator-1.1.0.tgz

# Extract the Helm chart
tar -xf ./toe-operator-*.tgz
```

### 2. Configure Environment Variables

Set your AWS account details and desired version:

```bash
export AWS_ACCOUNT_ID=864899852480
export AWS_REGION=ap-south-1
export VERSION=1.1.0
```

### 3. Sync Images to ECR

Run the included script to sync container images from GHCR to your ECR:

```bash
./toe-operator/scripts/sync-images-from-ghcr-to-ecr.sh \
  --account-id $AWS_ACCOUNT_ID \
  --region $AWS_REGION \
  --image-version $VERSION
```

This script will:
- Authenticate with both GHCR and ECR
- Pull images from GitHub Container Registry
- Tag and push them to your ECR repositories

### 4. Deploy TOE Operator

Install the TOE operator using Helm with ECR image references:

```bash
ECR_REGISTRY_PREFIX=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/codriverlabs/ce

helm install --create-namespace --namespace toe-system \
  --set-string global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set-string controller.image.tag=$VERSION \
  --set-string collector.image.tag=$VERSION \
  --set-string collector.storage.storageClass=efs-sc \
  toe-operator-$VERSION ./toe-operator
```

### 5. Verify Deployment

Check that all components are running:

```bash
# Check pods
kubectl get pods -n toe-system

# Check services
kubectl get services -n toe-system

# Check CRDs
kubectl get crd | grep codriverlabs

# Expected output:
# powertoolconfigs.codriverlabs.ai.toe.run
# powertools.codriverlabs.ai.toe.run
```

## Configuration Options

### Debug Mode

For troubleshooting, add the `--debug` flag:

```bash
helm install --debug --create-namespace --namespace toe-system \
  --set-string global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set-string controller.image.tag=$VERSION \
  --set-string collector.image.tag=$VERSION \
  toe-operator-$VERSION ./toe-operator
```

## Troubleshooting

### cert-manager Issues

If you encounter cert-manager webhook errors:

```bash
# Check cert-manager status
kubectl get pods -n cert-manager

# Scale up cert-manager if needed
kubectl scale deployment cert-manager --replicas=1 -n cert-manager
kubectl scale deployment cert-manager-cainjector --replicas=1 -n cert-manager  
kubectl scale deployment cert-manager-webhook --replicas=1 -n cert-manager
```

### ECR Authentication

Ensure your EKS nodes have the `AmazonEC2ContainerRegistryPullOnly` policy attached, or configure image pull secrets if using IRSA.

### Certificate Issues

If TLS certificate creation fails:

```bash
# Check certificate status
kubectl get certificates -n toe-system

# Check issuer status
kubectl get issuers -n toe-system

# Delete and recreate if needed
kubectl delete certificate toe-collector-cert -n toe-system
kubectl delete secret toe-collector-certs -n toe-system
```

## Uninstallation

To completely remove the TOE operator:

```bash
# Uninstall Helm release
helm uninstall toe-operator-$VERSION -n toe-system

# Delete namespace
kubectl delete namespace toe-system

# Clean up CRDs
kubectl delete crd powertoolconfigs.codriverlabs.ai.toe.run
kubectl delete crd powertools.codriverlabs.ai.toe.run
```

## Next Steps

After successful deployment, you can create PowerTool resources to start profiling your applications. See the [examples](examples/) directory for sample configurations.
