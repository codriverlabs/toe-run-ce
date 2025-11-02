#!/bin/sh

# Process chaos script - terminate, OOM, suspend processes
# Usage: process-chaos.sh <target_pid> <duration_sec> <interval_sec> [action]

TARGET_PID="${1:-1}"
DURATION="${2:-30}"
INTERVAL="${3:-5}"
ACTION="${4:-suspend}"

echo "Process Chaos Experiment Started"
echo "Target PID: $TARGET_PID"
echo "Duration: ${DURATION}s"
echo "Interval: ${INTERVAL}s"
echo "Action: $ACTION"
echo "Timestamp: $(date)"

# Validate target process exists
if ! kill -0 "$TARGET_PID" 2>/dev/null; then
    echo "ERROR: Process $TARGET_PID not found or not accessible"
    exit 1
fi

case "$ACTION" in
    "terminate-graceful")
        echo "Sending SIGTERM to process $TARGET_PID"
        kill -TERM "$TARGET_PID"
        echo "Graceful termination signal sent"
        ;;
    
    "terminate-force")
        echo "Sending SIGKILL to process $TARGET_PID"
        kill -KILL "$TARGET_PID"
        echo "Force termination signal sent"
        ;;
    
    "oom")
        echo "Triggering OOM condition for target process $TARGET_PID..."
        
        # Get target process info
        if [ -f "/proc/$TARGET_PID/status" ]; then
            echo "Target process info:"
            grep -E "Name|VmSize|VmRSS" /proc/$TARGET_PID/status || true
        fi
        
        echo "Using gdb to inject memory allocation into target process..."
        echo "This will make the target process allocate memory in its own cgroup"
        
        # Create a gdb script that allocates and touches memory
        cat > /tmp/oom-inject.gdb <<'EOF'
set pagination off
set confirm off
# Allocate 150MB and write to it to force physical allocation
call (void*)malloc(157286400)
set $ptr = (char*)$1
# Touch every page (4KB) to force actual allocation
set $i = 0
while $i < 157286400
  set *($ptr + $i) = 1
  set $i = $i + 4096
end
detach
quit
EOF
        
        echo "Injecting 150MB memory allocation into PID $TARGET_PID..."
        gdb -p $TARGET_PID -batch -x /tmp/oom-inject.gdb 2>&1 | grep -v "^warning:" | tail -20 &
        GDB_PID=$!
        
        # Monitor target process
        for i in 1 2 3 4 5 6 7 8 9 10; do
            sleep 2
            if ! kill -0 "$TARGET_PID" 2>/dev/null; then
                echo "SUCCESS: Target process $TARGET_PID was OOM killed after ${i} checks"
                kill -9 $GDB_PID 2>/dev/null || true
                exit 0
            fi
            if [ $i -eq 5 ]; then
                echo "Still injecting memory... (this may take time)"
            fi
        done
        
        wait $GDB_PID 2>/dev/null || true
        
        if ! kill -0 "$TARGET_PID" 2>/dev/null; then
            echo "SUCCESS: Target process $TARGET_PID was OOM killed"
        else
            echo "Target process survived memory injection"
            if [ -f "/proc/$TARGET_PID/status" ]; then
                echo "Final memory usage:"
                grep -E "VmSize|VmRSS" /proc/$TARGET_PID/status || true
            fi
        fi
        ;;
    
    "suspend")
        echo "Starting suspend/resume cycle for ${DURATION}s"
        END_TIME=$(($(date +%s) + DURATION))
        
        while [ $(date +%s) -lt $END_TIME ]; do
            if kill -0 "$TARGET_PID" 2>/dev/null; then
                echo "$(date): Suspending process $TARGET_PID"
                kill -STOP "$TARGET_PID"
                sleep "$INTERVAL"
                
                if kill -0 "$TARGET_PID" 2>/dev/null; then
                    echo "$(date): Resuming process $TARGET_PID"
                    kill -CONT "$TARGET_PID"
                fi
                sleep "$INTERVAL"
            else
                echo "Target process $TARGET_PID no longer exists"
                break
            fi
        done
        
        # Ensure process is resumed if still exists
        if kill -0 "$TARGET_PID" 2>/dev/null; then
            kill -CONT "$TARGET_PID" 2>/dev/null || true
            echo "Final resume signal sent to process $TARGET_PID"
        fi
        ;;
    
    *)
        echo "Unknown action: $ACTION"
        echo "Available actions: terminate-graceful, terminate-force, oom, suspend"
        exit 1
        ;;
esac

echo "Process chaos experiment completed at $(date)"
