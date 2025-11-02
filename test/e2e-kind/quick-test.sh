#!/bin/bash
set -euo pipefail

# Quick test script for Phase 1 using existing cluster

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🧪 Quick Phase 1 Test (using existing cluster)"
echo ""

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "toe-test-e2e"; then
    echo "❌ No cluster found. Please run setup-cluster.sh first"
    exit 1
fi

# Set context
kubectl config use-context kind-toe-test-e2e

# Verify cluster is accessible
echo "✅ Cluster: $(kubectl config current-context)"
echo "✅ Nodes:"
kubectl get nodes

# Check if CRDs are installed
echo ""
echo "📋 Checking CRDs..."
if ! kubectl get crd powertools.codriverlabs.ai.toe.run 2>/dev/null; then
    echo "⚠️ Installing CRDs..."
    kubectl apply -f "$PROJECT_ROOT/config/crd/bases/"
fi

# Check if controller is running
echo ""
echo "🔍 Checking TOE controller..."
if ! kubectl get deployment -n toe-system toe-controller-manager 2>/dev/null; then
    echo "⚠️ Controller not found - tests will run without controller"
    echo "   (This is OK for basic API tests)"
fi

# Run Phase 1 tests
echo ""
echo "🚀 Running Phase 1: Ephemeral Container Tests"
echo ""

cd "$SCRIPT_DIR"
go test -v -tags=e2ekind -timeout=10m ./... \
    -ginkgo.v \
    -ginkgo.progress \
    -ginkgo.focus="Ephemeral Container Profiling" || {
    
    TEST_EXIT=$?
    echo ""
    echo "❌ Tests failed. Debug info:"
    echo ""
    kubectl get all -A
    echo ""
    kubectl get powertools -A 2>/dev/null || echo "No PowerTools found"
    echo ""
    exit $TEST_EXIT
}

echo ""
echo "✅ Phase 1 tests completed!"
