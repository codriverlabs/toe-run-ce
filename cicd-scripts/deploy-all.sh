#!/bin/bash
# TOE Complete Deployment Script
# Usage: ./deploy-all.sh [local|ecr]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Override registry type from command line
if [ $# -gt 0 ]; then
    REGISTRY_TYPE="$1"
fi

# Validate registry type
case "$REGISTRY_TYPE" in
    "local"|"ecr")
        ;;
    *)
        log_error "Invalid registry type '$REGISTRY_TYPE'. Use 'local' or 'ecr'"
        echo "Usage: $0 [local|ecr]"
        exit 1
        ;;
esac

log_info "Starting TOE deployment with registry: $REGISTRY_TYPE, version: $VERSION"

# Function to run script with error handling
run_script() {
    local script_name="$1"
    local description="$2"
    local args="${3:-}"
    
    if [ ! -f "$script_name" ]; then
        log_error "Script not found: $script_name"
        exit 1
    fi
    
    if [ ! -x "$script_name" ]; then
        log_error "Script not executable: $script_name"
        exit 1
    fi
    
    log_info "Running: $description"
    echo "----------------------------------------"
    
    if [ -n "$args" ]; then
        if ./"$script_name" "$REGISTRY_TYPE" $args; then
            log_success "$description completed"
            echo ""
        else
            log_error "$description failed"
            exit 1
        fi
    else
        if ./"$script_name" "$REGISTRY_TYPE"; then
            log_success "$description completed"
            echo ""
        else
            log_error "$description failed"
            exit 1
        fi
    fi
}

# Function to cleanup existing deployment
cleanup_deployment() {
    log_info "Performing automatic cleanup of existing deployment..."
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Found existing $NAMESPACE namespace. Cleaning up..."
        
        # Run cleanup script
        if [ -f "clean-controller.sh" ]; then
            run_script "clean-controller.sh" "Cleanup existing controller deployment"
        else
            log_warning "clean-controller.sh not found, performing manual cleanup"
            kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
            kubectl delete clusterrole manager-role --ignore-not-found=true
            kubectl delete clusterrolebinding manager-rolebinding --ignore-not-found=true
            kubectl delete crd powertools.codriverlabs.ai.toe.run --ignore-not-found=true
            kubectl delete crd powertoolconfigs.codriverlabs.ai.toe.run --ignore-not-found=true
        fi
        
        # Wait for namespace deletion
        log_info "Waiting for namespace deletion to complete..."
        while kubectl get namespace "$NAMESPACE" &> /dev/null; do
            echo "    Namespace $NAMESPACE still terminating, waiting..."
            sleep 3
        done
        log_success "Cleanup completed"
    else
        log_info "No existing deployment found"
    fi
}

# Main deployment function
main() {
    cd "$SCRIPT_DIR"
    
    echo "========================================"
    echo "TOE Complete Deployment Script"
    echo "========================================"
    echo "Registry Type: $REGISTRY_TYPE"
    echo "Version: $VERSION"
    echo "Namespace: $NAMESPACE"
    echo ""
    
    # Automatic cleanup (enforced)
    cleanup_deployment
    
    echo ""
    log_info "Starting build phase..."
    echo ""
    
    # Build Phase: Build all components first
    log_info "Phase 1: Building Controller"
    run_script "build-controller.sh" "Build controller image"
    
    log_info "Phase 2: Building Collector"
    run_script "build-collector.sh" "Build collector image"
    
    log_info "Phase 3: Building PowerTool Images"
    run_script "build-powertool-tool.sh" "Build PowerTool images" "aperf"
    
    echo ""
    log_info "All builds completed successfully! Starting deployment phase..."
    echo ""
    
    # Deploy Phase: Deploy all components
    log_info "Phase 4: Deploying Controller"
    run_script "deploy-controller.sh" "Deploy controller to cluster"
    
    log_info "Phase 5: Deploying Collector"
    run_script "deploy-collector.sh" "Deploy collector to cluster"
    
    echo ""
    echo "========================================"
    log_success "TOE Deployment Completed Successfully!"
    echo "========================================"
    echo ""
    
    # Deployment verification
    log_info "Verifying deployment..."
    echo ""
    
    log_info "Checking controller status:"
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=toe
    echo ""
    
    log_info "Checking collector status:"
    kubectl get pods -n "$NAMESPACE" -l app=toe-collector
    echo ""
    
    log_info "Checking CRDs:"
    kubectl get crds | grep codriverlabs.ai.toe.run
    echo ""
    
    log_info "Checking ServiceAccounts:"
    kubectl get serviceaccounts -n "$NAMESPACE"
    echo ""
    
    echo "========================================"
    log_success "Deployment verification completed!"
    echo ""
    log_info "Next steps:"
    echo "1. Create PowerToolConfig resources for your profiling tools"
    echo "2. Deploy target applications to profile"
    echo "3. Create PowerTool resources to start profiling"
    echo ""
    log_info "For examples, see the main project README.md"
    echo "========================================"
}

# Trap to handle script interruption
trap 'log_error "Deployment interrupted"; exit 1' INT TERM

# Run main function
main "$@"
