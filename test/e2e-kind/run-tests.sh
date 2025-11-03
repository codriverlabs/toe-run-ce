#!/bin/bash
set -euo pipefail

# TOE Kind E2E Test Runner
# Orchestrates cluster setup, test execution, and cleanup

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Configuration
COMMIT_HASH="${GITHUB_SHA:-$(git rev-parse --short HEAD 2>/dev/null || echo 'local')}"
export CLUSTER_NAME="${CLUSTER_NAME:-toe-e2e-${COMMIT_HASH}}"
export IMAGE_TAG="e2e-${COMMIT_HASH}"
export IMAGE_NAME="toe-controller:${IMAGE_TAG}"
KEEP_CLUSTER="${KEEP_CLUSTER:-false}"
TEST_PHASE="${TEST_PHASE:-all}"
TEST_TIMEOUT="${TEST_TIMEOUT:-30m}"

echo "🎯 TOE Kind E2E Test Runner"
echo "Cluster Name: $CLUSTER_NAME"
echo "Commit Hash: $COMMIT_HASH"
echo "Image Tag: $IMAGE_TAG"
echo "Test Phase: $TEST_PHASE"
echo "Keep Cluster: $KEEP_CLUSTER"
echo ""

# Cleanup function
cleanup() {
    local exit_code=$?
    
    if [ "$KEEP_CLUSTER" = "true" ]; then
        echo "⚠️ KEEP_CLUSTER=true, skipping cleanup"
        echo "To cleanup manually: kind delete cluster --name $CLUSTER_NAME"
    else
        echo "🧹 Running cleanup..."
        "$SCRIPT_DIR/cluster/teardown-cluster.sh" || true
    fi
    
    exit $exit_code
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

# Step 1: Setup cluster
echo "📦 Step 1: Setting up Kind cluster..."
"$SCRIPT_DIR/cluster/setup-cluster.sh"

# Step 2: Build and load controller image
echo "🔨 Step 2: Building controller image..."
cd "$PROJECT_ROOT"

if ! docker images | grep -q "toe-controller.*${IMAGE_TAG}"; then
    echo "Building controller image: $IMAGE_NAME"
    make docker-build IMG="$IMAGE_NAME"
else
    echo "✅ Controller image already exists: $IMAGE_NAME"
fi

echo "📦 Loading image into Kind cluster..."
kind load docker-image "$IMAGE_NAME" --name "$CLUSTER_NAME"

# Step 3: Deploy TOE components
echo "🚀 Step 3: Deploying TOE components..."

# Install CRDs
kubectl apply -f config/crd/bases/

# Deploy controller with dynamic image
cat test/e2e-kind/manifests/toe-controller.yaml | \
    sed "s|image: toe-controller:e2e|image: $IMAGE_NAME|g" | \
    kubectl apply -f -

# Wait for deployment
echo "⏳ Waiting for TOE controller to be ready..."
kubectl wait --for=condition=available --timeout=300s \
    deployment/toe-controller-manager -n toe-system

# Step 4: Run tests
echo "🧪 Step 4: Running E2E tests..."

TEST_ARGS="-v -tags=e2ekind"
GINKGO_ARGS="-ginkgo.v -ginkgo.progress -ginkgo.show-node-events"

case "$TEST_PHASE" in
    phase1|ephemeral)
        echo "Running Phase 1: Ephemeral Container Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Ephemeral"
        ;;
    phase2|workloads)
        echo "Running Phase 2: Real Workload Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Real Workload"
        ;;
    phase3|storage)
        echo "Running Phase 3: Storage Integration Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Storage Integration"
        ;;
    phase4|multinode)
        echo "Running Phase 4: Multi-Node Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Multi-Node"
        ;;
    phase5|security)
        echo "Running Phase 5: Security and RBAC Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Security and RBAC"
        ;;
    phase6|failures)
        echo "Running Phase 6: Failure Scenario Tests"
        TEST_ARGS="$TEST_ARGS -ginkgo.focus=Failure Scenarios"
        ;;
    all)
        echo "Running All Test Phases"
        ;;
    *)
        echo "❌ Unknown test phase: $TEST_PHASE"
        echo "Valid phases: phase1, phase2, phase3, phase4, phase5, phase6, all"
        exit 1
        ;;
esac

# Execute tests
cd "$SCRIPT_DIR"
go test $TEST_ARGS $GINKGO_ARGS -timeout="$TEST_TIMEOUT" ./... || {
    TEST_EXIT_CODE=$?
    echo "❌ Tests failed with exit code: $TEST_EXIT_CODE"
    
    # Collect debug information
    echo "📋 Collecting debug information..."
    kubectl get all -A
    kubectl get powertools -A
    kubectl get powertoolconfigs -A
    kubectl describe pods -n toe-system
    
    exit $TEST_EXIT_CODE
}

echo ""
echo "✅ All tests passed!"
echo ""
echo "Summary:"
echo "  Cluster: $CLUSTER_NAME"
echo "  Image: $IMAGE_NAME"
echo "  Phase: $TEST_PHASE"
echo "  Duration: $(date)"
echo ""
echo "To cleanup image: docker rmi $IMAGE_NAME"
