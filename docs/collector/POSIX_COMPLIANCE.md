# POSIX Compliance for Path Names

## Issue

Using `=` in filesystem paths can cause issues:
- Not POSIX-compliant for directory names
- Can break shell scripts and tools
- Not DNS-safe for potential future use
- Inconsistent with Kubernetes naming conventions

## Solution

Use hyphen (`-`) instead of equals (`=`) to separate label keys and values.

### Format

**Before:** `key=value`  
**After:** `key-value`

### Examples

| Label Selector | Old Path | New Path |
|----------------|----------|----------|
| `app: nginx` | `/data/default/app=nginx/...` | `/data/default/app-nginx/...` |
| `env: production` | `/data/prod/env=production/...` | `/data/prod/env-production/...` |
| `tier: backend` | `/data/default/tier=backend/...` | `/data/default/tier-backend/...` |

## Benefits

✅ **POSIX-compliant:** Safe for all filesystem operations  
✅ **DNS-safe:** Can be used in hostnames if needed  
✅ **Shell-friendly:** No escaping needed in scripts  
✅ **Kubernetes-aligned:** Matches K8s naming conventions  
✅ **Uniform:** Consistent with other path components

## Implementation

Single line change in `extractMatchingLabels()`:

```go
// Before
labels = append(labels, fmt.Sprintf("%s=%s", key, value))

// After
labels = append(labels, fmt.Sprintf("%s-%s", key, value))
```

## Compatibility

This is a breaking change for existing deployments. Profiles will be stored in new paths:

**Migration:**
```bash
# Move existing profiles to new structure
cd /data/default
for dir in *=*; do
  newdir=$(echo "$dir" | tr '=' '-')
  mv "$dir" "$newdir"
done
```

Or simply redeploy and start fresh with the new structure.

## Standards Reference

- **POSIX.1-2017:** Portable Filename Character Set: `A-Z a-z 0-9 . _ -`
- **RFC 1123:** DNS label format: `[a-z0-9]([-a-z0-9]*[a-z0-9])?`
- **Kubernetes:** Label values must match `(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?`

Using hyphen ensures compatibility across all these standards.
