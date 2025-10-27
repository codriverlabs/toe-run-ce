#!/bin/bash
# Helper script to inspect collector PVC contents

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSPECTOR_YAML="$SCRIPT_DIR/toe-collector-pvc-inspector.yaml"

echo "🔍 TOE Collector PVC Inspector"
echo "=============================="

if ! kubectl get pod pvc-inspector -n toe-system >/dev/null 2>&1; then
    echo "❌ Inspector pod not found. Creating..."
    kubectl apply -f "$INSPECTOR_YAML"
    kubectl wait --for=condition=Ready pod/pvc-inspector -n toe-system --timeout=60s
fi

echo ""
echo "📁 PVC Contents:"
kubectl exec -n toe-system pvc-inspector -- ls -lah /data

echo ""
echo "💡 Usage Examples:"
echo "  # List all files:"
echo "  kubectl exec -n toe-system pvc-inspector -- ls -la /data"
echo ""
echo "  # View file contents:"
echo "  kubectl exec -n toe-system pvc-inspector -- cat /data/filename"
echo ""
echo "  # Interactive shell:"
echo "  kubectl exec -it -n toe-system pvc-inspector -- /bin/sh"
echo ""
echo "  # Copy file to local:"
echo "  kubectl cp toe-system/pvc-inspector:/data/filename ./filename"
echo ""
echo "  # Remove inspector pod:"
echo "  kubectl delete -f $INSPECTOR_YAML"
