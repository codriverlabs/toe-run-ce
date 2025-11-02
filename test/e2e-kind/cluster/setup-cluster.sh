#!/bin/bash
set -euo pipefail

# TOE Kind E2E Cluster Setup Script
# This script creates a Kind cluster and prepares it for TOE E2E testing

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Commit-based cluster naming for parallel test runs
COMMIT_HASH="${GITHUB_SHA:-$(git rev-parse --short HEAD 2>/dev/null || echo 'local')}"
CLUSTER_NAME="${CLUSTER_NAME:-toe-e2e-${COMMIT_HASH}}"

KUBECTL_VERSION="${KUBECTL_VERSION:-v1.34.0}"

echo "🚀 Setting up Kind cluster for TOE E2E testing..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for condition
wait_for_condition() {
    local condition="$1"
    local timeout="${2:-300}"
    local interval="${3:-5}"
    
    echo "⏳ Waiting for: $condition"
    for ((i=0; i<timeout; i+=interval)); do
        if eval "$condition"; then
            echo "✅ Condition met: $condition"
            return 0
        fi
        sleep "$interval"
    done
    echo "❌ Timeout waiting for: $condition"
    return 1
}

# Install Kind if not present
install_kind() {
    if ! command_exists kind; then
        echo "📦 Installing Kind..."
        if command_exists go; then
            echo "Using go install for latest version..."
            go install sigs.k8s.io/kind@latest
        else
            echo "Using binary download..."
            curl -Lo ./kind "https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64"
            chmod +x ./kind
            sudo mv ./kind /usr/local/bin/kind
        fi
    else
        echo "✅ Kind already installed: $(kind version)"
    fi
}

# Install kubectl if not present
install_kubectl() {
    if ! command_exists kubectl; then
        echo "📦 Installing kubectl $KUBECTL_VERSION..."
        curl -LO "https://dl.k8s.io/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/kubectl
    else
        echo "✅ kubectl already installed: $(kubectl version --client 2>/dev/null | head -1 || echo 'kubectl')"
    fi
}

# Setup Docker/Podman compatibility
setup_container_runtime() {
    if command_exists podman && [ ! -S /var/run/docker.sock ]; then
        echo "🔧 Setting up Podman Docker socket compatibility..."
        
        # Enable Podman socket
        systemctl --user enable --now podman.socket || true
        
        # Create Docker socket symlink
        sudo mkdir -p /var/run
        if [ -S "/run/user/$UID/podman/podman.sock" ]; then
            sudo ln -sf "/run/user/$UID/podman/podman.sock" /var/run/docker.sock
        elif [ -S "/run/podman/podman.sock" ]; then
            sudo ln -sf /run/podman/podman.sock /var/run/docker.sock
        fi
        
        # Set Docker host
        export DOCKER_HOST=unix:///var/run/docker.sock
        
        # Verify compatibility
        docker version >/dev/null 2>&1 || {
            echo "❌ Docker socket compatibility failed"
            exit 1
        }
        echo "✅ Podman Docker socket ready"
    elif command_exists docker; then
        echo "✅ Docker runtime detected"
    else
        echo "❌ No container runtime found (Docker or Podman required)"
        exit 1
    fi
}

# Create Kind cluster
create_cluster() {
    echo "🏗️ Creating Kind cluster: $CLUSTER_NAME"
    
    # Delete existing cluster if it exists
    if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
        echo "🧹 Deleting existing cluster: $CLUSTER_NAME"
        kind delete cluster --name "$CLUSTER_NAME"
    fi
    
    # Create new cluster (using default configuration for ARM64 compatibility)
    kind create cluster \
        --name "$CLUSTER_NAME" \
        --wait 300s
    
    # Set kubectl context
    kubectl cluster-info --context "kind-$CLUSTER_NAME"
}

# Load container images
load_images() {
    echo "📦 Loading TOE container images..."
    
    cd "$PROJECT_ROOT"
    
    # Build images if they don't exist
    if ! docker images | grep -q "toe-controller"; then
        echo "🔨 Building TOE controller image..."
        make docker-build IMG=toe-controller:e2e
    fi
    
    if ! docker images | grep -q "toe-collector"; then
        echo "🔨 Building TOE collector image..."
        make collector-build IMG=toe-collector:e2e
    fi
    
    # Load images into Kind cluster
    echo "⬆️ Loading images into Kind cluster..."
    kind load docker-image toe-controller:e2e --name "$CLUSTER_NAME"
    kind load docker-image toe-collector:e2e --name "$CLUSTER_NAME"
    
    # Verify images are loaded
    docker exec -it "${CLUSTER_NAME}-control-plane" crictl images | grep toe || {
        echo "❌ Failed to load TOE images"
        exit 1
    }
    
    echo "✅ TOE images loaded successfully"
}

# Setup cluster networking
setup_networking() {
    echo "🌐 Setting up cluster networking..."
    
    # Wait for CNI to be ready
    wait_for_condition "kubectl get nodes | grep -q Ready" 180
    
    # Install ingress controller if needed
    if ! kubectl get pods -n ingress-nginx | grep -q ingress-nginx-controller; then
        echo "📡 Installing NGINX Ingress Controller..."
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
        
        # Wait for ingress controller to be ready
        wait_for_condition "kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=90s" 120
    fi
    
    echo "✅ Networking setup complete"
}

# Setup storage
setup_storage() {
    echo "💾 Setting up storage classes..."
    
    # Verify default storage class exists
    if ! kubectl get storageclass | grep -q "(default)"; then
        echo "⚠️ No default storage class found, creating one..."
        kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: rancher.io/local-path
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
EOF
    fi
    
    echo "✅ Storage setup complete"
}

# Setup RBAC and security
setup_security() {
    echo "🔐 Setting up RBAC and security..."
    
    # Create TOE system namespace
    kubectl create namespace toe-system --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply basic RBAC (will be enhanced by TOE manifests)
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: toe-e2e-sa
  namespace: toe-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: toe-e2e-role
rules:
- apiGroups: [""]
  resources: ["pods", "pods/ephemeralcontainers", "pods/status"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["codriverlabs.ai.toe.run"]
  resources: ["powertools", "powertoolconfigs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: toe-e2e-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: toe-e2e-role
subjects:
- kind: ServiceAccount
  name: toe-e2e-sa
  namespace: toe-system
EOF
    
    echo "✅ Security setup complete"
}

# Verify cluster health
verify_cluster() {
    echo "🔍 Verifying cluster health..."
    
    # Check node status
    kubectl get nodes -o wide
    
    # Check system pods
    kubectl get pods -A
    
    # Verify API server is responsive
    kubectl version 2>/dev/null | head -2 || echo "kubectl version check"
    
    # Check cluster info
    kubectl cluster-info
    
    echo "✅ Cluster verification complete"
}

# Main execution
main() {
    echo "🎯 TOE Kind E2E Cluster Setup"
    echo "Cluster Name: $CLUSTER_NAME"
    echo "Kubectl Version: $KUBECTL_VERSION"
    echo ""
    
    install_kind
    install_kubectl
    setup_container_runtime
    create_cluster
    setup_networking
    setup_storage
    setup_security
    verify_cluster
    
    echo ""
    echo "🎉 Kind cluster setup complete!"
    echo "Cluster Name: $CLUSTER_NAME"
    echo "Context: kind-$CLUSTER_NAME"
    echo ""
    echo "Next steps:"
    echo "  1. Deploy TOE components: kubectl apply -f test/e2e-kind/manifests/"
    echo "  2. Run E2E tests: make test-e2e-kind"
    echo "  3. Cleanup: ./test/e2e-kind/cluster/teardown-cluster.sh"
}

# Execute main function
main "$@"
