# Version Management

This document describes how Go version is centrally managed across the project.

## Single Source of Truth: `.go-version`

The Go version is defined in a single file: **`.go-version`**

```
1.25.3
```

## How It's Used

### 1. GitHub Actions Workflows

All workflows read from `.go-version`:

```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version-file: .go-version
```

**Files:**
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/test.yml`
- `.github/workflows/security-scan.yml`

### 2. Dockerfiles

Dockerfiles use a build argument with default from `.go-version`:

```dockerfile
ARG GO_VERSION=1.25.3
FROM golang:${GO_VERSION} AS builder
```

**Files:**
- `Dockerfile` (controller)
- `build/collector/Dockerfile` (collector)

The release workflow extracts and passes the version:

```yaml
- name: Extract Go version
  id: go-version
  run: echo "GO_VERSION=$(cat .go-version)" >> $GITHUB_OUTPUT

- name: Build and push image
  uses: docker/build-push-action@v6
  with:
    build-args: |
      GO_VERSION=${{ steps.go-version.outputs.GO_VERSION }}
```

### 3. go.mod

The `go.mod` file also specifies the Go version:

```go
go 1.25.3
```

This is used by:
- Local development (Go toolchain)
- `go-version-file: go.mod` in workflows (alternative approach)

## Updating Go Version

To update the Go version across the entire project:

1. **Update `.go-version`:**
   ```bash
   echo "1.26.0" > .go-version
   ```

2. **Update `go.mod`:**
   ```bash
   go mod edit -go=1.26.0
   ```

3. **Update Dockerfile defaults (optional but recommended):**
   ```bash
   sed -i 's/ARG GO_VERSION=.*/ARG GO_VERSION=1.26.0/' Dockerfile
   sed -i 's/ARG GO_VERSION=.*/ARG GO_VERSION=1.26.0/' build/collector/Dockerfile
   ```

4. **Test locally:**
   ```bash
   make test
   make docker-build
   ```

5. **Commit and push:**
   ```bash
   git add .go-version go.mod Dockerfile build/collector/Dockerfile
   git commit -m "chore: Update Go version to 1.26.0"
   git push
   ```

## Benefits

✅ **Single Source of Truth**: One file (`.go-version`) controls the version  
✅ **Consistency**: Same Go version across CI/CD, Docker builds, and local dev  
✅ **Easy Updates**: Change one file, everything updates  
✅ **No Hardcoding**: Workflows and Dockerfiles read from `.go-version`  
✅ **Fail-Safe**: Dockerfiles have default values if build arg not provided  

## Verification

Check all Go versions are consistent:

```bash
# Check .go-version
cat .go-version

# Check go.mod
grep "^go " go.mod

# Check Dockerfiles
grep "ARG GO_VERSION" Dockerfile build/collector/Dockerfile

# Check workflows
grep -r "go-version" .github/workflows/
```

All should show the same version!
