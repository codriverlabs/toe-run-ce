# TODO

## CI/CD Issues

### GitHub Actions Test Failures
- **Issue**: Controller tests fail in public repository GitHub Actions but work in private repository
- **Error**: `fork/exec /usr/local/kubebuilder/bin/etcd: no such file or directory`
- **Root Cause**: envtest binaries not properly set up in public repo CI environment
- **Investigation Needed**: 
  - Compare private vs public repo runner configurations
  - Test `make setup-envtest` locally to simulate GitHub Actions environment
  - Verify if `setup-envtest` step completes successfully in CI
- **Potential Solutions**:
  - Add explicit kubebuilder installation step
  - Debug envtest binary setup process
  - Check if KUBEBUILDER_ASSETS environment variable is properly set
