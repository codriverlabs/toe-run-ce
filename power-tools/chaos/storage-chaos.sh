#!/bin/sh

# Storage chaos script - fill up storage paths
# Usage: storage-chaos.sh <duration_sec> [fill_percent] [target_path]

DURATION="${1:-30}"
FILL_PERCENT="${2:-80}"
TARGET_PATH="${3:-/tmp}"

echo "Storage Chaos Experiment Started"
echo "Duration: ${DURATION}s"
echo "Fill percentage: ${FILL_PERCENT}%"
echo "Target path: $TARGET_PATH"
echo "Timestamp: $(date)"

# Discover storage paths if not specified
if [ "$TARGET_PATH" = "/tmp" ]; then
    echo "Discovering mounted storage paths..."
    
    # Look for common mount points
    for path in /tmp /var /opt /usr/local /home; do
        if [ -d "$path" ] && [ -w "$path" ]; then
            echo "Found writable path: $path"
        fi
    done
    
    # Use df to show filesystem usage
    if command -v df >/dev/null 2>&1; then
        echo "Current filesystem usage:"
        df -h
    fi
fi

# Validate target path
if [ ! -d "$TARGET_PATH" ]; then
    echo "ERROR: Target path $TARGET_PATH does not exist"
    exit 1
fi

if [ ! -w "$TARGET_PATH" ]; then
    echo "ERROR: Target path $TARGET_PATH is not writable"
    exit 1
fi

# Get available space
if command -v df >/dev/null 2>&1; then
    AVAILABLE_KB=$(df "$TARGET_PATH" | tail -1 | awk '{print $4}')
    FILL_KB=$((AVAILABLE_KB * FILL_PERCENT / 100))
    echo "Available space: ${AVAILABLE_KB}KB"
    echo "Will fill: ${FILL_KB}KB (${FILL_PERCENT}%)"
else
    # Fallback if df is not available
    FILL_KB=102400  # 100MB default
    echo "Cannot determine available space, using default: ${FILL_KB}KB"
fi

# Create chaos directory
CHAOS_DIR="$TARGET_PATH/chaos-storage-$$"
mkdir -p "$CHAOS_DIR"
echo "Created chaos directory: $CHAOS_DIR"

# Fill storage in chunks
CHUNK_SIZE=1024  # 1MB chunks
CHUNKS=$((FILL_KB / CHUNK_SIZE))
if [ $CHUNKS -lt 1 ]; then
    CHUNKS=1
fi

echo "Filling storage with $CHUNKS chunks of ${CHUNK_SIZE}KB each..."

# Start filling storage
i=0
while [ $i -lt $CHUNKS ]; do
    CHUNK_FILE="$CHAOS_DIR/fill_$i"
    
    # Create file filled with zeros
    if command -v dd >/dev/null 2>&1; then
        dd if=/dev/zero of="$CHUNK_FILE" bs=1024 count=$CHUNK_SIZE 2>/dev/null || break
    else
        # Fallback method using shell
        j=0
        while [ $j -lt $((CHUNK_SIZE * 1024)) ]; do
            printf "0" >> "$CHUNK_FILE" || break
            j=$((j + 1024))
        done
    fi
    
    i=$((i + 1))
    
    # Show progress every 10 chunks
    if [ $((i % 10)) -eq 0 ]; then
        echo "Created $i/$CHUNKS chunks..."
        if command -v df >/dev/null 2>&1; then
            df -h "$TARGET_PATH" | tail -1
        fi
    fi
    
    # Small delay to avoid overwhelming the system
    sleep 0.1
done

echo "Storage fill completed. Files created in $CHAOS_DIR"

# Monitor for the duration
echo "Monitoring storage chaos for ${DURATION}s..."
sleep "$DURATION"

# Cleanup
echo "Cleaning up chaos files..."
rm -rf "$CHAOS_DIR"
echo "Cleanup completed"

# Final storage status
if command -v df >/dev/null 2>&1; then
    echo "Final filesystem usage:"
    df -h "$TARGET_PATH"
fi

echo "Storage chaos experiment completed at $(date)"
