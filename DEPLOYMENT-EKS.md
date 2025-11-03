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

## 📦 Container Images for EKS

The following images need to be synced to your ECR:

- **Controller**: `ghcr.io/codriverlabs/ce/toe-controller:v1.1.0`
- **Collector**: `ghcr.io/codriverlabs/ce/toe-collector:v1.1.0`
- **Aperf Tool**: `ghcr.io/codriverlabs/ce/toe-aperf:v1.1.0`
- **Tcpdump Tool**: `ghcr.io/codriverlabs/ce/toe-tcpdump:v1.1.0`
- **Chaos Tool**: `ghcr.io/codriverlabs/ce/toe-chaos:v1.1.0`

## Deployment Steps

### 1. Download and Extract Release

Download the latest release from GitHub and extract the Helm chart:

```bash
# Download the release
VERSION=v1.1.0
wget https://github.com/codriverlabs/toe-run-ce/releases/download/$VERSION/toe-operator-$VERSION.tgz

# Extract the Helm chart
tar -xf toe-operator-$VERSION.tgz
```

### 2. Configure Environment Variables

Set your AWS account details and desired version:

```bash
export AWS_ACCOUNT_ID=123456789012
export AWS_REGION=us-west-2
export VERSION=v1.1.0
export ECR_REGISTRY_PREFIX=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/codriverlabs/ce
```

### 3. Sync Images to ECR

The Helm chart includes a script to sync all container images from GHCR to your ECR:

```bash
# Make the script executable
chmod +x ./toe-operator/scripts/sync-images-from-ghcr-to-ecr.sh

# Run the sync script
./toe-operator/scripts/sync-images-from-ghcr-to-ecr.sh \
  --account-id $AWS_ACCOUNT_ID \
  --region $AWS_REGION \
  --image-version $VERSION
```

This script will:
- Authenticate with both GHCR and ECR
- Pull all 5 container images from GitHub Container Registry
- Create ECR repositories if they don't exist
- Tag and push images to your ECR repositories

### 4. Deploy TOE Operator (Standard Deployment)

Install the TOE operator with controller and collector (PowerTools optional):

```bash
helm install toe-operator ./toe-operator \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set powertools.enabled=true \
  --set collector.storage.storageClass=gp2
```

### 5. Alternative: Without PowerTools

For deployment without PowerTool configurations:

```bash
helm install toe-operator ./toe-operator \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set powertools.enabled=false
```

### 6. Verify Deployment

Check that all components are running:

```bash
# Check pods
kubectl get pods -n toe-system

# Expected output (standard deployment):
# NAME                                        READY   STATUS    RESTARTS   AGE
# toe-collector-xxx                          1/1     Running   0          2m
# toe-operator-controller-manager-xxx        2/2     Running   0          2m

# Check services
kubectl get services -n toe-system

# Check CRDs
kubectl get crd | grep codriverlabs

# Expected output:
# powertoolconfigs.codriverlabs.ai.toe.run
# powertools.codriverlabs.ai.toe.run

# Check PowerTool configurations (if enabled)
kubectl get powertoolconfigs -n toe-system

# Expected output:
# NAME           AGE
# aperf-config   2m
# chaos-config   2m
# tcpdump-config 2m
```

## 🎯 EKS-Specific Configuration Options

### Using EFS for Collector Storage

```bash
# First, create EFS storage class
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: efs-sc
provisioner: efs.csi.aws.com
parameters:
  provisioningMode: efs-ap
  fileSystemId: fs-xxxxxxxxx
  directoryPerms: "0755"
EOF

# Deploy with EFS storage
helm install toe-operator ./toe-operator \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set collector.storage.storageClass=efs-sc \
  --set collector.storage.size=100Gi
```

### Using IRSA (IAM Roles for Service Accounts)

```bash
# Create IRSA role first (replace with your cluster name and account)
eksctl create iamserviceaccount \
  --cluster=my-cluster \
  --namespace=toe-system \
  --name=toe-operator-controller-manager \
  --attach-policy-arn=arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
  --approve

# Deploy with IRSA
helm install toe-operator ./toe-operator \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set ecr.useIRSA=true \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::$AWS_ACCOUNT_ID:role/eksctl-my-cluster-addon-iamserviceaccount-toe-system-toe-operator-controller-manager-Role1-xxx
```

## 🔧 Configuration Customization

### Custom PowerTool Settings

```bash
# Deploy with custom PowerTool configurations
helm install toe-operator ./toe-operator \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX \
  --set powertools.aperf.allowedNamespaces="{production,staging}" \
  --set powertools.chaos.resources.limits.memory=512Mi \
  --set powertools.tcpdump.defaultArgs="{-i,eth0,-c,1000}"
```

### Debug Mode

For troubleshooting, enable debug mode:

```bash
helm install toe-operator ./toe-operator \
  --debug \
  --create-namespace \
  --namespace toe-system \
  --set global.version=$VERSION \
  --set global.registry.repository=$ECR_REGISTRY_PREFIX
```

## 🔍 Validation and Testing

### Test PowerTool Functionality

```bash
# Create a test pod
kubectl apply -f ./toe-operator/examples/targets/target-pod-with-pvc.yaml

# Create a PowerTool to profile it
kubectl apply -f ./toe-operator/examples/powertool-aperf-ephemeral.yaml

# Check PowerTool status
kubectl get powertools -A

# View PowerTool logs
kubectl describe powertool profile-my-app
```

## 🛠️ Troubleshooting

### cert-manager Issues

If you encounter cert-manager webhook errors:

```bash
# Check cert-manager status
kubectl get pods -n cert-manager

# Restart cert-manager if needed
kubectl rollout restart deployment cert-manager -n cert-manager
kubectl rollout restart deployment cert-manager-webhook -n cert-manager
```

### ECR Authentication Issues

```bash
# Verify ECR repositories exist
aws ecr describe-repositories --region $AWS_REGION | grep codriverlabs

# Test ECR authentication
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Check node IAM permissions
kubectl describe node | grep ProviderID
```

### Image Pull Issues

```bash
# Check image pull secrets
kubectl get secrets -n toe-system

# Verify image references in pods
kubectl describe pod -n toe-system -l app.kubernetes.io/name=toe-operator

# Check PowerTool configurations
kubectl get powertoolconfigs -n toe-system -o yaml
```

### Certificate Issues

```bash
# Check certificate status
kubectl get certificates -n toe-system

# Check issuer status
kubectl get issuers -n toe-system

# Recreate certificates if needed
kubectl delete certificate toe-collector-cert -n toe-system
kubectl delete secret toe-collector-certs -n toe-system
helm upgrade toe-operator ./toe-operator --reuse-values
```

## 🗑️ Uninstallation

To completely remove the TOE operator:

```bash
# Delete all PowerTool resources first
kubectl delete powertools --all -A

# Uninstall Helm release
helm uninstall toe-operator -n toe-system

# Delete namespace
kubectl delete namespace toe-system

# Clean up CRDs
kubectl delete crd powertoolconfigs.codriverlabs.ai.toe.run
kubectl delete crd powertools.codriverlabs.ai.toe.run

# Optional: Remove ECR repositories
aws ecr delete-repository --repository-name codriverlabs/ce/toe-controller --region $AWS_REGION --force
aws ecr delete-repository --repository-name codriverlabs/ce/toe-collector --region $AWS_REGION --force
aws ecr delete-repository --repository-name codriverlabs/ce/toe-aperf --region $AWS_REGION --force
aws ecr delete-repository --repository-name codriverlabs/ce/toe-tcpdump --region $AWS_REGION --force
aws ecr delete-repository --repository-name codriverlabs/ce/toe-chaos --region $AWS_REGION --force
```

## 📚 Next Steps

After successful deployment:

1. **Create PowerTools**: Use the examples in `./toe-operator/examples/` to create PowerTool resources
2. **Monitor Performance**: Check controller and collector logs for performance insights  
3. **Scale as Needed**: Adjust replica counts and resource limits based on your workload
4. **Security Review**: Review the security documentation in `docs/security/`

For more examples and advanced configurations, see the [main deployment guide](DEPLOYMENT.md).
