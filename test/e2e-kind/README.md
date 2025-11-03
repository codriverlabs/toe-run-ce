# TOE Kind E2E Testing

## Overview

This directory contains comprehensive end-to-end tests for TOE using Kind (Kubernetes in Docker) clusters. These tests validate functionality that cannot be tested with envtest, particularly ephemeral container creation and real pod profiling.

**Kind Architecture:**
- Runs local Kubernetes clusters using Docker container "nodes"
- Each "node" is bootstrapped with kubeadm
- Designed for testing Kubernetes itself, local development, and CI
- Provides Go packages for cluster creation and image management

**Complete Isolation:** Both clusters and images use commit-based naming:
- Cluster: `toe-e2e-<commit-hash>`
- Image: `toe-controller:e2e-<commit-hash>`

## Quick Start

### Fully Automated (Recommended)
```bash
# Everything automated: build → deploy → test → cleanup
./test/e2e-kind/run-tests.sh
```

This single command:
- ✅ Creates Kind cluster
- ✅ Builds controller image
- ✅ Deploys TOE components
- ✅ Runs all tests
- ✅ Cleans up automatically

### Run Specific Phase
```bash
TEST_PHASE=phase1 ./test/e2e-kind/run-tests.sh
```

### Keep Cluster for Debugging
```bash
KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh
```

### Manual Steps (Advanced)
```bash
# 1. Setup cluster only
./test/e2e-kind/cluster/setup-cluster.sh

# 2. Build and deploy manually
make docker-build IMG=toe-controller:e2e
kind load docker-image toe-controller:e2e --name <cluster-name>
kubectl apply -f test/e2e-kind/manifests/

# 3. Run tests
go test -v -tags=e2ekind ./test/e2e-kind/...

# 4. Cleanup
./test/e2e-kind/cluster/teardown-cluster.sh
```

## Test Phases

### Phase 1: Ephemeral Container Tests ✅ IMPLEMENTED
- Ephemeral container creation in target pods
- Tool execution in ephemeral containers
- Container lifecycle management
- Data collection from ephemeral containers
- Multiple target pod scenarios

**File:** `ephemeral_containers_test.go`

### Phase 2: Real Workload Profiling 🔄 TODO
- Profile nginx under load
- Profile multi-container pods
- Profile redis/database workloads
- Profile batch jobs

**File:** `real_workloads_test.go` (to be created)

### Phase 3: Storage Integration 🔄 TODO
- PVC output mode
- Collector output mode
- Storage class selection
- Data persistence verification

**File:** `storage_integration_test.go` (to be created)

### Phase 4: Multi-Node Scenarios 🔄 TODO
- Cross-node profiling
- Node affinity constraints
- Network policies

**File:** `multi_node_test.go` (to be created)

### Phase 5: Security and RBAC 🔄 TODO
- Namespace isolation
- Security context enforcement
- Service account permissions

**File:** `security_rbac_test.go` (to be created)

### Phase 6: Failure Scenarios 🔄 TODO
- Pod lifecycle failures
- Controller failures
- Resource exhaustion

**File:** `failure_scenarios_test.go` (to be created)

## Cluster Management

### Cluster Naming
Clusters are named using commit hashes to enable parallel testing:
```
toe-e2e-<commit-hash>
```

Example: `toe-e2e-a1b2c3d4`

### Setup
```bash
./test/e2e-kind/cluster/setup-cluster.sh
```

Creates:
- Multi-node Kind cluster (1 control-plane + 2 workers)
- Installs TOE CRDs
- Configures RBAC
- Sets up networking and storage

### Teardown
```bash
./test/e2e-kind/cluster/teardown-cluster.sh
```

Performs:
- Artifact collection
- Resource cleanup
- Cluster deletion
- Container image pruning

## Reusable Code from Existing E2E Suite

### From `test/e2e/simple_utils.go`
All utility functions are reusable with minor timeout adjustments:
- Namespace management
- Pod creation and waiting
- PowerTool/PowerToolConfig creation
- Status checking and logging

### From `test/e2e/*_test.go`
Controller logic tests are fully reusable:
- Target pod discovery
- Reconciliation logic
- Resource lifecycle
- Error handling
- Validation tests

## New Test Capabilities

### What Kind Enables (vs Envtest)
✅ **Actual ephemeral container creation**
✅ **Real kubelet interaction**
✅ **Container runtime integration**
✅ **Network policies**
✅ **Persistent volumes**
✅ **Multi-node scenarios**
✅ **Real workload profiling**

### What Envtest Already Covers
✅ Controller reconciliation logic
✅ CRD validation
✅ RBAC rules
✅ Status updates
✅ Label selectors
✅ Error handling

## Environment Variables

- `CLUSTER_NAME`: Override cluster name (default: `toe-e2e-<commit>`)
- `KEEP_CLUSTER`: Keep cluster after tests (default: `false`)
- `TEST_PHASE`: Run specific phase (default: `all`)
- `TEST_TIMEOUT`: Test timeout (default: `30m`)
- `GITHUB_SHA`: Commit hash for CI (auto-detected locally)

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run Kind E2E Tests
  env:
    GITHUB_SHA: ${{ github.sha }}
  run: ./test/e2e-kind/run-tests.sh
```

### Artifacts
Test artifacts are automatically collected in `test-artifacts-<timestamp>/`:
- Cluster info dump
- PowerTool/PowerToolConfig resources
- Controller and collector logs
- Events
- Node descriptions

## Development Workflow

### 1. Local Testing
```bash
# Run tests with cluster kept for debugging
KEEP_CLUSTER=true TEST_PHASE=phase1 ./test/e2e-kind/run-tests.sh

# Inspect cluster
kubectl get all -A
kubectl get powertools -A
kubectl logs -n toe-system -l app=toe-controller

# Cleanup when done
kind delete cluster --name toe-e2e-local
```

### 2. Adding New Tests
1. Create test file in `test/e2e-kind/`
2. Add build tag: `//go:build e2e-kind`
3. Use existing utilities from `utils.go`
4. Follow Ginkgo/Gomega patterns
5. Update this README

### 3. Debugging Failed Tests
```bash
# Check cluster state
kubectl get all -A

# Check TOE resources
kubectl get powertools -A -o yaml
kubectl get powertoolconfigs -A -o yaml

# Check logs
kubectl logs -n toe-system -l app=toe-controller --tail=100
kubectl logs -n toe-system -l app=toe-collector --tail=100

# Check events
kubectl get events -A --sort-by='.lastTimestamp'
```

## Performance Benchmarks

- **Cluster Setup**: ~2 minutes
- **Test Suite (Phase 1)**: ~5 minutes
- **Full Suite (All Phases)**: ~15 minutes
- **Resource Usage**: <4GB RAM, <2 CPU cores

## Success Criteria

- ✅ All tests pass consistently (>95% success rate)
- ✅ No resource leaks after test completion
- ✅ Proper cleanup of ephemeral containers
- ✅ Actual profiling data collection verified
- ✅ Multi-pod scenarios work correctly

## Next Steps

1. ✅ Phase 1 implementation complete
2. 🔄 Implement Phase 2 (Real Workloads)
3. 🔄 Implement Phase 3 (Storage Integration)
4. 🔄 Implement Phase 4 (Multi-Node)
5. 🔄 Implement Phase 5 (Security/RBAC)
6. 🔄 Implement Phase 6 (Failure Scenarios)
7. 🔄 CI/CD integration
8. 🔄 Performance optimization

## References

- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Ephemeral Containers](https://kubernetes.io/docs/concepts/workloads/pods/ephemeral-containers/)
- [Ginkgo Testing Framework](https://onsi.github.io/ginkgo/)
- [TOE E2E Strategy](../docs/e2e-kind-strategy.md)
