#!/bin/bash
# Deploy script for toe-k8s-operator
# Usage: ./deploy.sh [local|ecr] [--clean]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Validation functions
validate_tools() {
    echo "Validating required tools..."
    
    if ! command -v kubectl &> /dev/null; then
        echo "❌ Error: 'kubectl' is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v make &> /dev/null; then
        echo "❌ Error: 'make' is not installed or not in PATH"
        exit 1
    fi
    
    echo "✅ All required tools are available"
}

validate_cluster_access() {
    echo "Validating Kubernetes cluster access..."
    
    if ! kubectl cluster-info &> /dev/null; then
        echo "❌ Error: Cannot connect to Kubernetes cluster"
        echo "   Make sure kubectl is configured and cluster is accessible"
        exit 1
    fi
    
    echo "✅ Kubernetes cluster is accessible"
}

validate_helper_scripts() {
    if [ "$REGISTRY_TYPE" = "ecr" ]; then
        echo "Validating helper scripts for ECR..."
        
        if [ ! -f "$PROJECT_ROOT/configure-image-pull-secrets.sh" ]; then
            echo "❌ Error: configure-image-pull-secrets.sh not found"
            exit 1
        fi
        
        if [ ! -f "$PROJECT_ROOT/helper_scripts/setup-namespace-docker-secrets.sh" ]; then
            echo "❌ Error: helper_scripts/setup-namespace-docker-secrets.sh not found"
            exit 1
        fi
        
        echo "✅ Helper scripts are available"
    fi
}

CLEAN_DEPLOY=false
REGISTRY_TYPE="${REGISTRY_TYPE:-local}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        local|ecr)
            REGISTRY_TYPE="$1"
            shift
            ;;
        --clean)
            CLEAN_DEPLOY=true
            shift
            ;;
        *)
            echo "❌ Error: Invalid argument '$1'"
            echo "Usage: $0 [local|ecr] [--clean]"
            exit 1
            ;;
    esac
done

# Set image based on registry type
case "$REGISTRY_TYPE" in
    "local")
        IMAGE="$LOCAL_REGISTRY/codriverlabs/$PROJECT_NAME"
        ;;
    "ecr")
        IMAGE="$ECR_REGISTRY/codriverlabs/$PROJECT_NAME"
        ;;
    *)
        echo "❌ Error: Invalid registry type '$REGISTRY_TYPE'. Use 'local' or 'ecr'"
        exit 1
        ;;
esac

cd "$PROJECT_ROOT"

echo "=== Deploying toe-k8s-operator ==="
echo "Image: $IMAGE:$VERSION"
echo "Registry Type: $REGISTRY_TYPE"
echo "Clean Deploy: $CLEAN_DEPLOY"

# Validate environment
validate_tools
validate_cluster_access
validate_helper_scripts

# Step 1: Clean deployment if requested
if [ "$CLEAN_DEPLOY" = true ]; then
    echo "Step 1: Cleaning existing deployment..."
    
    echo "  Undeploying operator..."
    if ! make undeploy-controller-only; then
        echo "⚠️  Warning: Undeploy failed (may not exist)"
    else
        echo "✅ Operator undeployed"
    fi
    
    echo "  Uninstalling CRDs..."
    if ! make uninstall; then
        echo "⚠️  Warning: Uninstall failed (may not exist)"
    else
        echo "✅ CRDs uninstalled"
    fi
    
    # Brief pause to allow resources to be cleaned up
    echo "  Allowing time for resource cleanup..."
    sleep 2
    echo "✅ Cleanup completed"
fi

# Step 2: Setup secrets for ECR
if [ "$REGISTRY_TYPE" = "ecr" ]; then
    echo "Step 2: Setting up ECR secrets..."
    
    echo "  Configuring image pull secrets for operator..."
    if ! ./configure-image-pull-secrets.sh "$SECRET_NAME"; then
        echo "❌ Error: Failed to configure image pull secrets"
        exit 1
    fi
    echo "✅ Image pull secrets configured"
    
    echo "  Creating namespace..."
    if ! kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -; then
        echo "❌ Error: Failed to create namespace"
        exit 1
    fi
    echo "✅ Namespace created/updated"
    
    echo "  Setting up namespace secrets..."
    if ! ./helper_scripts/setup-namespace-docker-secrets.sh "$NAMESPACE" "$ECR_REGISTRY" "$ECR_REGION"; then
        echo "❌ Error: Failed to setup namespace secrets"
        exit 1
    fi
    echo "✅ Namespace secrets configured"
fi

# Step 3: Install CRDs
echo "Step 3: Installing CRDs..."
if ! make install; then
    echo "❌ Error: Failed to install CRDs"
    exit 1
fi
echo "✅ CRDs installed successfully"

# Step 4: Deploy operator
echo "Step 4: Deploying operator..."
if ! make deploy IMG="$IMAGE:$VERSION"; then
    echo "❌ Error: Failed to deploy operator"
    exit 1
fi
echo "✅ Operator deployed successfully"

# Step 5: Verify deployment
echo "Step 5: Verifying deployment..."
sleep 5
if ! kubectl get deployment -n "$NAMESPACE" | grep -q "toe-controller-manager"; then
    echo "⚠️  Warning: Operator deployment not found, checking pods..."
    kubectl get pods -n "$NAMESPACE" || true
else
    echo "✅ Operator deployment verified"
fi

echo ""
echo "🎉 Deployment completed successfully!"
echo "🔍 Check status: kubectl get pods -n $NAMESPACE"
echo ""
echo "📝 Example PowerTool:"
cat << 'EOF'
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerTool
metadata:
  name: my-powertool
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "aperf"
    duration: "30s"
  output:
    mode: "ephemeral"
EOF
