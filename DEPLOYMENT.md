# TOE Deployment Guide

This guide covers different ways to deploy the Tactical Observability Engine (TOE) operator.

## 🚀 Quick Start

### Option 1: Direct YAML Installation (Recommended)

```bash
# Install the latest release
kubectl apply -f https://github.com/codriverlabs/toe/releases/latest/download/toe-operator-latest.yaml

# Or install a specific version
kubectl apply -f https://github.com/codriverlabs/toe/releases/download/v1.0.20-beta/toe-operator-1.0.20-beta.yaml
```

### Option 2: Helm Installation

```bash
# Add the repository (when available)
helm repo add toe https://codriverlabs.github.io/toe

# Or install directly from GitHub release
helm install toe-operator \
  https://github.com/codriverlabs/toe/releases/download/v1.0.20-beta/toe-operator-1.0.0.tgz
```

## 🔧 Development Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/codriverlabs/toe.git
cd toe

# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/codriverlabs/toe-controller:latest
```

### Local Development

```bash
# Run locally (requires kubeconfig)
make run
```

## 📦 Build Your Own Release

```bash
# Generate all release artifacts
make github-release VERSION=1.0.20-beta

# Controller only
make github-release-controller VERSION=1.0.20-beta

# Collector only  
make github-release-collector VERSION=1.0.20-beta

# Build all Docker images locally
make docker-build-all VERSION=1.0.20-beta

# Push all Docker images
make docker-push-all VERSION=1.0.20-beta
```

## 🎯 Container Images

The following container images are published with each release:

- **Controller**: `ghcr.io/codriverlabs/toe-controller:latest`
- **Collector**: `ghcr.io/codriverlabs/toe-collector:latest`  
- **Aperf Tool**: `ghcr.io/codriverlabs/toe-aperf:latest`

## 🎯 Deployment Options

### Controller + Collector (Full Stack)
```bash
kubectl apply -f toe-operator-1.0.20-beta.yaml
kubectl apply -f toe-collector-1.0.20-beta.yaml
```

### Controller Only
```bash
kubectl apply -f toe-operator-1.0.20-beta.yaml
```

### Custom Registry (ECR Example)
```bash
helm install toe-operator toe-operator-1.0.0.tgz \
  --set global.registry.type=ecr \
  --set controller.image.repository=123456789012.dkr.ecr.us-west-2.amazonaws.com/codriverlabs/toe-controller \
  --set collector.image.repository=123456789012.dkr.ecr.us-west-2.amazonaws.com/codriverlabs/toe-collector
```

## 🔍 Verification

```bash
# Check operator status
kubectl get pods -n toe-system

# Check CRDs
kubectl get crd | grep toe

# View logs
kubectl logs -n toe-system deployment/toe-operator-controller-manager
```

## 🗑️ Uninstallation

```bash
# Remove operator
kubectl delete -f toe-operator-1.0.20-beta.yaml

# Remove CRDs (this will delete all PowerTool resources!)
kubectl delete crd powertools.codriverlabs.ai.toe.run
kubectl delete crd powertoolconfigs.codriverlabs.ai.toe.run
```
