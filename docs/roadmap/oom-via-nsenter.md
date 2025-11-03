# OOM via nsenter - Experimental Approach

## Status: 🔬 Experimental / Not Yet Implemented

## Concept

Use `nsenter` from a privileged ephemeral container to execute memory allocation commands using binaries that already exist in the target container. This may allow memory to be allocated in the target container's cgroup.

## Theory

### Current Problem
- Ephemeral containers have separate cgroups from target containers
- Memory allocated by ephemeral container doesn't count against target's limits
- gdb injection still allocates memory in caller's cgroup

### Proposed Solution

1. **Enter target's mount namespace** using `nsenter`
2. **Execute binaries from target container** (not from ephemeral container)
3. **Allocate memory using target's binaries**

The hypothesis: If we execute a binary that's already in the target container's filesystem, it might inherit the target's cgroup membership.

## Implementation Approach

### Step 1: Use nsenter to Enter Target Namespaces

```bash
# From privileged ephemeral container
nsenter --target $TARGET_PID --mount --pid --cgroup /bin/sh -c "COMMAND"
```

### Step 2: Execute Memory Allocation Using Target's Binaries

Use common utilities that exist in most containers:

**Option A: Using head/tail/sleep**
```bash
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'head -c 500M /dev/zero | tail -c +1 & sleep 60'
```

**Option B: Using dd**
```bash
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'dd if=/dev/zero of=/dev/null bs=1M count=500 & sleep 60'
```

**Option C: Using yes (if available)**
```bash
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'yes | head -c 500M > /dev/null & sleep 60'
```

### Step 3: Monitor for OOM

```bash
# Check if target process gets killed
while kill -0 $TARGET_PID 2>/dev/null; do
    sleep 1
done
echo "Target process terminated (OOM killed)"
```

## Requirements

### Capabilities Needed
- `SYS_ADMIN` - Required for nsenter
- `SYS_PTRACE` - May be needed for process inspection
- `allowPrivileged: true` - Full privileges

### Container Requirements
Target container must have:
- `/bin/sh` or `/bin/bash`
- Basic utilities: `head`, `tail`, `sleep` OR `dd`
- `/dev/zero` device

## Expected Behavior

### If It Works ✅
- Memory allocation happens in target's cgroup
- Target container hits its memory limit
- OOM killer selects target process
- Container restarts (if restart policy allows)

### If It Doesn't Work ❌
- Memory still allocated in ephemeral container's cgroup
- Ephemeral container gets OOM killed
- Target container unaffected

## Testing Plan

### Test 1: Verify Cgroup Membership

```bash
# Execute command via nsenter
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'echo $$ > /tmp/test_pid; sleep 10' &

# Check which cgroup the process belongs to
cat /proc/$(cat /tmp/test_pid)/cgroup
```

**Expected:** Should show target container's cgroup, not ephemeral's

### Test 2: Small Memory Allocation

```bash
# Allocate 50MB (less than limits)
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'head -c 50M /dev/zero | tail -c +1 & sleep 30'

# Monitor memory usage
watch -n 1 'cat /proc/1/status | grep VmRSS'
```

### Test 3: Trigger OOM

```bash
# Allocate 200MB (exceeds 128Mi limit)
nsenter --target 1 --mount --pid --cgroup \
  /bin/sh -c 'head -c 200M /dev/zero | tail -c +1 & sleep 60'

# Monitor for OOM
kubectl get pod $POD_NAME -w
```

## Implementation in Chaos Tool

### Modified process-chaos.sh

```bash
"oom-nsenter")
    echo "Triggering OOM via nsenter approach..."
    
    # Check if nsenter is available
    if ! command -v nsenter >/dev/null 2>&1; then
        echo "ERROR: nsenter not available"
        exit 1
    fi
    
    # Calculate memory to allocate (150MB to exceed 128Mi limit)
    MEMORY_MB=150
    
    echo "Allocating ${MEMORY_MB}MB via target container's binaries..."
    
    # Use nsenter to execute memory allocation in target's context
    nsenter --target $TARGET_PID --mount --pid --cgroup \
        /bin/sh -c "head -c ${MEMORY_MB}M /dev/zero | tail -c +1 & sleep $DURATION" &
    
    NSENTER_PID=$!
    
    # Monitor target process
    START_TIME=$(date +%s)
    while [ $(($(date +%s) - START_TIME)) -lt $DURATION ]; do
        if ! kill -0 "$TARGET_PID" 2>/dev/null; then
            echo "SUCCESS: Target process $TARGET_PID was OOM killed"
            kill -9 $NSENTER_PID 2>/dev/null || true
            exit 0
        fi
        sleep 1
    done
    
    # Cleanup
    kill -9 $NSENTER_PID 2>/dev/null || true
    
    if kill -0 "$TARGET_PID" 2>/dev/null; then
        echo "Target process survived - approach may not work"
    fi
    ;;
```

## Potential Issues

### 1. Cgroup Membership Still Wrong
Even with nsenter, the process may still be created in the ephemeral container's cgroup because:
- Process creation happens from the ephemeral container's context
- Kernel assigns cgroup at fork/exec time
- nsenter only changes namespace view, not cgroup membership

### 2. Binary Compatibility
Target container may not have required binaries:
- Minimal images (distroless, scratch) have no shell
- Alpine uses different paths (`/bin/sh` vs `/bin/bash`)
- Some containers have no `/dev/zero`

### 3. Permission Issues
Even with privileged mode:
- SELinux/AppArmor may block nsenter
- Container runtime may restrict namespace operations
- Cgroup filesystem may be read-only

## Alternative: Direct Cgroup Manipulation

If nsenter doesn't work, try direct cgroup manipulation:

```bash
# Find target's cgroup
TARGET_CGROUP=$(cat /proc/$TARGET_PID/cgroup | cut -d: -f3)

# Spawn memory consumer
head -c 200M /dev/zero | tail -c +1 &
CONSUMER_PID=$!

# Try to move it to target's cgroup
echo $CONSUMER_PID > /sys/fs/cgroup/$TARGET_CGROUP/cgroup.procs
```

**Likely result:** Permission denied or operation not permitted

## Success Criteria

This approach is considered successful if:
1. ✅ Memory allocation happens in target's cgroup (verified via `/proc/PID/cgroup`)
2. ✅ Target container hits OOM and gets killed
3. ✅ Ephemeral container survives
4. ✅ Works consistently across different container images

## Next Steps

1. **Implement experimental version** in chaos tool
2. **Test with various container images**:
   - nginx (full OS)
   - alpine (minimal)
   - distroless (no shell)
3. **Verify cgroup membership** of spawned processes
4. **Document results** and update this file
5. **If successful**: Promote to stable feature
6. **If unsuccessful**: Document why and mark as not feasible

## References

- [nsenter man page](https://man7.org/linux/man-pages/man1/nsenter.1.html)
- [Linux cgroup v2 documentation](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)
- [Kubernetes ephemeral containers](https://kubernetes.io/docs/concepts/workloads/pods/ephemeral-containers/)

## Related Issues

- Original OOM implementation: Uses stress-ng in ephemeral container (doesn't work)
- gdb injection attempt: Memory still counted in caller's cgroup (doesn't work)
- Sidecar approach: Would work but requires pod modification (not ephemeral)
