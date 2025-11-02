# Collector Path Structure

## Path Format

```
{basePath}/{namespace}/{matching-labels}/{powertool-name}/{year}/{month}/{day}/{filename}
```

## Components

| Component | Source | Default | Example |
|-----------|--------|---------|---------|
| **basePath** | Collector config | `/data` | `/data` |
| **namespace** | Target pod namespace | - | `default` |
| **matching-labels** | First label from selector that matched | `unknown` | `app-nginx` or `env-prod` |
| **powertool-name** | PowerTool resource name | - | `profile-job-123` |
| **year** | Current year | - | `2025` |
| **month** | Current month | - | `10` |
| **day** | Current day | - | `30` |
| **filename** | Original file or generated | `{powertool}.profile` | `aperf-output.txt` |

## Examples

### App Label Selector
```yaml
targets:
  labelSelector:
    matchLabels:
      app: nginx
```
Path: `/data/default/app-nginx/profile-nginx/2025/10/30/output.txt`

### Environment Label Selector
```yaml
targets:
  labelSelector:
    matchLabels:
      env: production
```
Path: `/data/default/env-production/profile-prod/2025/10/30/output.txt`

### Custom Label Selector
```yaml
targets:
  labelSelector:
    matchLabels:
      tier: backend
      version: v2
```
Path: `/data/default/tier-backend/profile-backend/2025/10/30/output.txt`
(Uses first matching label)

## Date Format Configuration

**Required:** ConfigMap must be created before deploying collector.

Default format creates hierarchical date structure:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: collector-config
  namespace: toe-system
data:
  dateFormat: "2006/01/02"  # Creates: year/month/day folders
```

### Common Formats

| Format | Output | Structure |
|--------|--------|-----------|
| `2006/01/02` | 2025/10/30 | Separate folders (default) |
| `2006-01-02` | 2025-10-30 | Single folder |
| `2006/01/02/15` | 2025/10/30/14 | With hour |
| `2006/01/02/15/04` | 2025/10/30/14/25 | With hour/minute |

**Recommendation:** Use `2006/01/02` for better organization and performance with large datasets.

## Metadata Flow

```
Target Pod Labels
    ↓
Controller (buildPowerToolEnvVars)
    ↓
Ephemeral Container Env Vars
    ↓
Power-Tool Script (send-profile.sh)
    ↓
HTTP Headers
    ↓
Collector Server (handleProfile)
    ↓
Storage Manager (SaveProfile)
    ↓
Hierarchical Path
```

## HTTP Headers

Power-tools send these headers:

```http
POST /api/v1/profile HTTP/1.1
Authorization: Bearer {token}
X-PowerTool-Namespace: default
X-PowerTool-Matching-Labels: app-nginx
X-PowerTool-Job-ID: profile-job-123
X-PowerTool-Filename: aperf-output.txt
Content-Type: application/octet-stream
```

The `X-PowerTool-Matching-Labels` header contains the first label from the PowerTool's `labelSelector.matchLabels` that matched the target pod, in `key-value` format (POSIX-compliant).

## Querying Profiles

### By Namespace
```bash
kubectl exec -n toe-system deployment/toe-collector -- ls /data/default/
```

### By Application
```bash
kubectl exec -n toe-system deployment/toe-collector -- ls /data/default/app-nginx/
```

### By Date (Year)
```bash
kubectl exec -n toe-system deployment/toe-collector -- find /data -type d -name "2025"
```

### By Date (Month)
```bash
kubectl exec -n toe-system deployment/toe-collector -- find /data -path "*/2025/10/*"
```

### By Date (Specific Day)
```bash
kubectl exec -n toe-system deployment/toe-collector -- find /data -path "*/2025/10/30/*"
```

### By PowerTool
```bash
kubectl exec -n toe-system deployment/toe-collector -- find /data -type d -name "profile-job-123"
```

### Recent Profiles (Today)
```bash
TODAY=$(date +%Y/%m/%d)
kubectl exec -n toe-system deployment/toe-collector -- find /data -path "*/$TODAY/*"
```

## Troubleshooting

### Missing matching labels
**Symptom:** Files saved under `/data/{namespace}/unknown/`

**Solution:** Ensure PowerTool has valid `labelSelector.matchLabels` that match target pods:
```yaml
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-application  # This label must exist on target pods
```

### Wrong date format
**Symptom:** Dates appear as `%!d(string=2006)` or similar

**Solution:** Use valid Go time format in ConfigMap

### Flat structure (old behavior)
**Symptom:** Files saved directly in `/data/`

**Solution:** Ensure:
1. Collector rebuilt with new code
2. Power-tools rebuilt with updated scripts
3. Controller redeployed with app label extraction
