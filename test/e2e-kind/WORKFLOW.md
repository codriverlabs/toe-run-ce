# Automated E2E Test Workflow

## Single Command Execution

```bash
./test/e2e-kind/run-tests.sh
```

## Complete Automation Flow

```
┌──────────────────────────────────────────────────────────────┐
│                  ./run-tests.sh                              │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ Step 1: Cluster Setup                                        │
│ ├─ Create cluster: toe-e2e-<commit-hash>                     │
│ ├─ Install networking (CNI)                                  │
│ ├─ Setup storage classes                                     │
│ └─ Configure RBAC                                            │
│                                                              │
│ Duration: ~2 minutes                                         │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ Step 2: Build & Load Images                                 │
│ ├─ Image: toe-controller:e2e-<commit-hash>                  │
│ ├─ Check if image exists                                     │
│ ├─ Build: make docker-build IMG=toe-controller:e2e-<hash>   │
│ ├─ Load: kind load docker-image toe-controller:e2e-<hash>   │
│ └─ Verify image in cluster                                   │
│                                                              │
│ Duration: ~3 minutes (first run), ~10s (cached)             │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ Step 3: Deploy TOE Components                               │
│ ├─ Install CRDs (PowerTool, PowerToolConfig)                │
│ ├─ Deploy controller (namespace, RBAC, deployment)          │
│ ├─ Wait for ready (300s timeout)                            │
│ └─ Verify deployment status                                  │
│                                                              │
│ Duration: ~30 seconds                                        │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ Step 4: Run Tests                                           │
│ ├─ Execute: go test -tags=e2ekind                           │
│ ├─ Run selected phase (or all)                              │
│ ├─ Collect results                                           │
│ └─ Generate reports                                          │
│                                                              │
│ Duration: ~5-15 minutes (depends on phase)                  │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ Step 5: Cleanup (unless KEEP_CLUSTER=true)                  │
│ ├─ Collect artifacts (logs, resources, events)              │
│ ├─ Delete PowerTools                                         │
│ ├─ Delete cluster                                            │
│ └─ Prune container images                                    │
│                                                              │
│ Duration: ~30 seconds                                        │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
                    ✅ Complete!
```

## What Gets Automated

### ✅ Infrastructure
- Kind cluster creation
- Networking setup
- Storage provisioning
- RBAC configuration

### ✅ Build Process
- Controller image build
- Image loading into cluster
- Build caching

### ✅ Deployment
- CRD installation
- Controller deployment
- Health checks
- Readiness verification

### ✅ Testing
- Test compilation
- Test execution
- Result collection
- Artifact generation

### ✅ Cleanup
- Resource deletion
- Cluster teardown
- Image pruning
- Artifact archival

## No Manual Steps Required

The entire workflow is **fully automated**:

```bash
# One command does everything
./test/e2e-kind/run-tests.sh

# Output:
# 🚀 Starting Kind E2E tests with cluster: toe-e2e-a1b2c3d4
# 📦 Step 1: Setting up Kind cluster...
# 🔨 Step 2: Building controller image...
# 🚀 Step 3: Deploying TOE components...
# 🧪 Step 4: Running E2E tests...
# ✅ All tests passed!
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: E2E Tests
  run: ./test/e2e-kind/run-tests.sh
```

### GitLab CI
```yaml
e2e-tests:
  script:
    - ./test/e2e-kind/run-tests.sh
```

### Jenkins
```groovy
stage('E2E Tests') {
    steps {
        sh './test/e2e-kind/run-tests.sh'
    }
}
```

## Customization

### Environment Variables
```bash
# Custom cluster name
CLUSTER_NAME=my-test ./test/e2e-kind/run-tests.sh

# Keep cluster after tests
KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh

# Run specific phase
TEST_PHASE=phase1 ./test/e2e-kind/run-tests.sh

# Custom timeout
TEST_TIMEOUT=45m ./test/e2e-kind/run-tests.sh
```

### Parallel Execution
```bash
# Multiple PRs on same runner - complete isolation
# PR #123 (commit: abc123)
./test/e2e-kind/run-tests.sh
# Cluster: toe-e2e-abc123
# Image: toe-controller:e2e-abc123

# PR #456 (commit: def456)
./test/e2e-kind/run-tests.sh
# Cluster: toe-e2e-def456
# Image: toe-controller:e2e-def456

# No conflicts - fully isolated!
```

## Error Handling

The script handles errors automatically:

```bash
# Build failure → stops execution
# Deployment failure → collects logs, exits
# Test failure → collects artifacts, exits with error code
# Cleanup always runs (unless KEEP_CLUSTER=true)
```

## Summary

**Before:** Manual 10-step process  
**After:** Single command

**Before:** ~30 minutes with manual steps  
**After:** ~10 minutes fully automated

**Before:** Error-prone manual deployment  
**After:** Consistent, repeatable automation

✅ **Zero manual intervention required**
