# TOE Deployment Guide

This guide covers different ways to deploy the Tactical Observability Engine (TOE) operator.

## 🚀 Quick Start

### Option 1: Direct YAML Installation (Recommended)

```bash
# Install the latest release
kubectl apply -f https://github.com/codriverlabs/toe-run-ce/releases/latest/download/toe-operator-v1.1.0-public-preview.yaml

# Or install a specific version
kubectl apply -f https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.yaml
```

### Option 2: Helm Installation

```bash
# Install directly from GitHub release
helm install toe-operator \
  https://github.com/codriverlabs/toe-run-ce/releases/latest/download/toe-operator-v1.1.0-public-preview.tgz

# Or with custom version
helm install toe-operator \
  https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.tgz \
  --set global.version=v1.1.0 \
  --set global.registry.repository=your-registry.com/toe
```

## 📦 Container Images

The following container images are published with each release:

- **Controller**: `ghcr.io/codriverlabs/ce/toe-controller:v1.1.0`
- **Collector**: `ghcr.io/codriverlabs/ce/toe-collector:v1.1.0`
- **Aperf Tool**: `ghcr.io/codriverlabs/ce/toe-aperf:v1.1.0`
- **Tcpdump Tool**: `ghcr.io/codriverlabs/ce/toe-tcpdump:v1.1.0`
- **Chaos Tool**: `ghcr.io/codriverlabs/ce/toe-chaos:v1.1.0`

## 🎯 What's Included in Each Release

### YAML Installer (`toe-operator-v1.1.0.yaml`)
- Complete operator deployment with CRDs
- Controller and RBAC configurations
- Webhook configurations
- Namespace setup (toe-system)

### Helm Chart (`toe-operator-v1.1.0.tgz`)
- Configurable Helm chart with values
- PowerTool configurations (aperf, tcpdump, chaos)
- ECR sync scripts for private registries
- Complete examples directory
- Support for different registry types (GHCR, ECR, local)

### Included Examples
- PowerTool configurations for all tools
- Target pod examples (StatefulSet, Pod with PVC)
- Testing configurations (multi-container, non-root)
- Output modes (ephemeral, PVC, collector)

## 🔧 Development Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/codriverlabs/toe-run-ce.git
cd toe

# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/codriverlabs/ce/toe-controller:latest

# Deploy PowerTool configurations
make deploy-configs
```

### Local Development

```bash
# Generate configs with current images
make generate-configs

# Run locally (requires kubeconfig)
make run
```

## 🏗️ Build Your Own Release

```bash
# Generate all release artifacts
make github-release VERSION=v1.1.0

# Build and push all Docker images
make docker-build-all VERSION=v1.1.0
make docker-push-all VERSION=v1.1.0

# Generate Helm chart only
make helm-chart VERSION=v1.1.0
```

## 🎯 Deployment Options

### Standard Deployment (Controller + Collector + PowerTools)
```bash
# Using YAML (includes both controller and collector)
kubectl apply -f https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.yaml

# Using Helm with PowerTools enabled
helm install toe-operator \
  https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.tgz \
  --set powertools.enabled=true
```

### Without PowerTools (Controller + Collector only)
```bash
helm install toe-operator \
  https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.tgz \
  --set powertools.enabled=false
```

### Custom Registry (ECR Example)
```bash
helm install toe-operator \
  https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.tgz \
  --set global.version=v1.1.0 \
  --set global.registry.repository=123456789012.dkr.ecr.us-west-2.amazonaws.com/codriverlabs/ce \
  --set ecr.accountId=123456789012 \
  --set ecr.region=us-west-2
```

## 🔍 Verification

```bash
# Check operator status
kubectl get pods -n toe-system

# Check CRDs
kubectl get crd | grep codriverlabs

# Check PowerTool configurations
kubectl get powertoolconfigs -n toe-system

# View controller logs
kubectl logs -n toe-system deployment/toe-operator-controller-manager

# View collector logs (if enabled)
kubectl logs -n toe-system deployment/toe-collector
```

## 🛠️ PowerTool Configuration

After deployment, PowerTool configurations are automatically created:

```bash
# List available tools
kubectl get powertoolconfigs -n toe-system

# Expected output:
# NAME           AGE
# aperf-config   1m
# chaos-config   1m
# tcpdump-config 1m
```

Create a PowerTool to use these configurations:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerTool
metadata:
  name: profile-my-app
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-application
  tool:
    name: "aperf"
    duration: "30s"
  output:
    mode: "ephemeral"
```

## 🗑️ Uninstallation

```bash
# Using Helm
helm uninstall toe-operator -n toe-system

# Using YAML (delete in reverse order)
kubectl delete -f https://github.com/codriverlabs/toe-run-ce/releases/download/v1.1.0/toe-operator-v1.1.0.yaml

# Remove CRDs (this will delete all PowerTool resources!)
kubectl delete crd powertools.codriverlabs.ai.toe.run
kubectl delete crd powertoolconfigs.codriverlabs.ai.toe.run

# Remove namespace
kubectl delete namespace toe-system
```

## 📚 Next Steps

- See [DEPLOYMENT-EKS.md](DEPLOYMENT-EKS.md) for EKS-specific deployment
- Check [examples/](examples/) for PowerTool usage examples
- Review [docs/security/](docs/security/) for security considerations
