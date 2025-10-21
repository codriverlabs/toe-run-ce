#!/bin/bash
set -e

# Use environment variables from PowerTool controller
DURATION=${PROFILER_DURATION:-30s}
WARMUP=${PROFILER_WARMUP:-0s}
TARGET_PID=${TARGET_PID:-1}
PROFILE_TYPE=${PROFILE_TYPE:-cpu}

# Determine output directory based on output mode
if [ "$OUTPUT_MODE" = "pvc" ]; then
    OUTPUT_DIR="/mnt/profiling-storage/${PVC_PATH:-profiles/aperf/}"
elif [ "$OUTPUT_MODE" = "s3" ]; then
    OUTPUT_DIR="/tmp/"
else
    OUTPUT_DIR="/tmp/"
fi

RUN_NAME="profile-${TARGET_POD_NAME}-${HOSTNAME}-$(date +%Y%m%d-%H%M%S)"

echo "Starting AWS aperf profiling..."
echo "Target Pod: ${TARGET_POD_NAME}"
echo "Target Container: ${TARGET_CONTAINER:-default}"
echo "Duration: ${DURATION}"
echo "Warmup: ${WARMUP}"
echo "Profile Type: ${PROFILE_TYPE}"
echo "Run Name: ${RUN_NAME}"
echo "Output Directory: ${OUTPUT_DIR}"

# Create output directory if needed
mkdir -p "$OUTPUT_DIR"

# Run warmup if specified
if [ "$WARMUP" != "0s" ] && [ -n "$WARMUP" ]; then
    echo "Warming up for ${WARMUP}..."
    sleep "${WARMUP%s}"
fi

# Run AWS aperf with specified parameters
aperf -vv record \
  --period="${DURATION%s}" \
  --run-name="$RUN_NAME" \
  --profile

# Find and copy aperf output files to the desired location
echo "Searching for aperf output files..."

# Look for the run directory and tar.gz file in root filesystem
APERF_DIR=$(find / -name "$RUN_NAME" -type d 2>/dev/null | head -1)
APERF_TAR=$(find / -name "${RUN_NAME}.tar.gz" -type f 2>/dev/null | head -1)

if [ -n "$APERF_DIR" ] && [ -d "$APERF_DIR" ]; then
    echo "Found aperf directory: $APERF_DIR"
    cp -r "$APERF_DIR"/* "$OUTPUT_DIR/" 2>/dev/null || true
fi

if [ -n "$APERF_TAR" ] && [ -f "$APERF_TAR" ]; then
    echo "Found aperf archive: $APERF_TAR"
    cp "$APERF_TAR" "$OUTPUT_DIR/" 2>/dev/null || true
fi

echo "Profiling completed. Output saved to $OUTPUT_DIR"

# List generated files
echo "Generated files:"
ls -la "$OUTPUT_DIR"

# Send to collector if configured
if [ -n "$COLLECTOR_ENDPOINT" ] && [ -n "$COLLECTOR_TOKEN" ]; then
    echo "Sending profiles to collector..."
    
    # Find and send only the tar.gz archive
    TAR_FILE=$(find "$OUTPUT_DIR" -name "*.tar.gz" -type f | head -1)
    if [ -n "$TAR_FILE" ] && [ -f "$TAR_FILE" ]; then
        echo "Sending archive: $TAR_FILE"
        if send-profile.sh "$TAR_FILE"; then
            echo "✅ Archive uploaded successfully"
        else
            echo "❌ Failed to upload archive"
        fi
    else
        echo "⚠️  No tar.gz archive found to upload"
    fi
    echo "Collector upload completed"
fi
