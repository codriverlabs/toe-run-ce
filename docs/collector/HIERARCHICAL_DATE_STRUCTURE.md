# Hierarchical Date Structure

## Overview

The collector uses a hierarchical date structure with separate folders for year, month, and day:

```
/data/{namespace}/{labels}/{powertool}/{year}/{month}/{day}/{filename}
```

Example:
```
/data/default/app-nginx/profile-job/2025/10/30/output.txt
```

## Why Hierarchical Dates?

### 1. Filesystem Performance

**Problem with flat structure:**
```
/data/default/app-nginx/profile-job/
├── 2025-01-01-file1.txt
├── 2025-01-01-file2.txt
├── 2025-01-02-file1.txt
├── 2025-01-02-file2.txt
...
└── 2025-12-31-file100.txt  (36,500+ files in one directory!)
```

**With hierarchical structure:**
```
/data/default/app-nginx/profile-job/
├── 2025/
│   ├── 01/
│   │   ├── 01/  (100 files max per day)
│   │   └── 02/
│   └── 12/
│       └── 31/
```

**Benefits:**
- Faster directory listings (100 files vs 36,500)
- Better inode cache utilization
- Reduced filesystem metadata overhead
- Improved backup/restore performance

### 2. Natural Partitioning

Easy to implement retention policies:

```bash
# Delete profiles older than 90 days
find /data -type d -path "*/2024/07/*" -exec rm -rf {} +

# Archive last month
tar -czf archive-2025-09.tar.gz /data/*/*/*/2025/09/

# Query specific date range
find /data -path "*/2025/10/[01-15]/*"
```

### 3. Intuitive Navigation

```bash
# Browse by year
ls /data/default/app-nginx/profile-job/2025/

# Browse by month
ls /data/default/app-nginx/profile-job/2025/10/

# Browse by day
ls /data/default/app-nginx/profile-job/2025/10/30/
```

### 4. Query Efficiency

```bash
# All profiles from October 2025
find /data -path "*/2025/10/*"

# All profiles from specific day
find /data -path "*/2025/10/30/*"

# All profiles from last 7 days
for i in {0..6}; do
  date=$(date -d "$i days ago" +%Y/%m/%d)
  find /data -path "*/$date/*"
done
```

### 5. Storage Optimization

- Easy to move old data to cheaper storage tiers
- Simple to implement compression by date range
- Natural boundaries for backup strategies

## Configuration

### Default (Recommended)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: collector-config
data:
  dateFormat: "2006/01/02"  # year/month/day
```

Result: `/data/.../2025/10/30/file.txt`

### With Hour (High-Volume)

```yaml
data:
  dateFormat: "2006/01/02/15"  # year/month/day/hour
```

Result: `/data/.../2025/10/30/14/file.txt`

Use when generating 1000+ profiles per day.

### With Hour and Minute (Very High-Volume)

```yaml
data:
  dateFormat: "2006/01/02/15/04"  # year/month/day/hour/minute
```

Result: `/data/.../2025/10/30/14/25/file.txt`

Use when generating 10,000+ profiles per day.

## Performance Comparison

### Flat Structure (Single Date Folder)

```
Format: "2006-01-02"
Path: /data/.../2025-10-30/

Files per directory: 1,000 - 100,000+
ls command: 1-5 seconds
find command: 5-30 seconds
Backup: Slow (single large directory)
```

### Hierarchical Structure (Year/Month/Day)

```
Format: "2006/01/02"
Path: /data/.../2025/10/30/

Files per directory: 10 - 1,000
ls command: <100ms
find command: <1 second
Backup: Fast (parallel by date)
```

## Best Practices

### 1. Retention Policies

```bash
#!/bin/bash
# Delete profiles older than 90 days
CUTOFF_DATE=$(date -d "90 days ago" +%Y/%m/%d)
find /data -type d -path "*/${CUTOFF_DATE%/*/*}/*" | while read dir; do
  if [[ "$dir" < "/data/any/path/$CUTOFF_DATE" ]]; then
    rm -rf "$dir"
  fi
done
```

### 2. Archival

```bash
# Archive by month
LAST_MONTH=$(date -d "last month" +%Y/%m)
tar -czf "profiles-${LAST_MONTH}.tar.gz" /data/*/*/*/${LAST_MONTH/\//-}/
```

### 3. Monitoring

```bash
# Count profiles per day
find /data -type f -path "*/2025/10/30/*" | wc -l

# Disk usage by month
du -sh /data/*/*/*/*/10/
```

## Migration from Flat Structure

If you have existing flat date structure:

```bash
#!/bin/bash
# Migrate from 2025-10-30 to 2025/10/30

cd /data
find . -type d -name "20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]" | while read dir; do
  # Extract date components
  date=$(basename "$dir")
  year=${date:0:4}
  month=${date:5:2}
  day=${date:8:2}
  
  # Create new structure
  parent=$(dirname "$dir")
  mkdir -p "$parent/$year/$month"
  
  # Move files
  mv "$dir" "$parent/$year/$month/$day"
done
```

## Summary

Hierarchical date structure is the right approach because:

✅ **Performance:** Faster operations with many files  
✅ **Scalability:** Handles millions of profiles efficiently  
✅ **Maintainability:** Easy retention and archival  
✅ **Intuitive:** Natural navigation and querying  
✅ **Standard:** Industry best practice for time-series data

**Recommendation:** Always use `2006/01/02` format unless you have specific requirements.
