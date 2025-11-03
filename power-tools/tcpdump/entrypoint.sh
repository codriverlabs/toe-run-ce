#!/bin/bash

set -euo pipefail

# Default values
DURATION="${DURATION:-30s}"
OUTPUT_FILE="${OUTPUT_FILE:-/tmp/tcpdump-${TARGET_POD_NAME:-unknown}-${HOSTNAME}-$(date +%Y%m%d-%H%M%S).pcap}"

echo "Starting tcpdump capture..."
echo "Duration: $DURATION"
echo "Output: $OUTPUT_FILE"
echo "Arguments: $*"

# Parse duration to timeout value
if [[ "$DURATION" =~ ^([0-9]+)([smh])$ ]]; then
    NUM="${BASH_REMATCH[1]}"
    UNIT="${BASH_REMATCH[2]}"
    
    case "$UNIT" in
        s) TIMEOUT="$NUM" ;;
        m) TIMEOUT=$((NUM * 60)) ;;
        h) TIMEOUT=$((NUM * 3600)) ;;
    esac
else
    echo "Invalid duration format: $DURATION (use format like 30s, 5m, 1h)"
    exit 1
fi

# Build tcpdump command with provided arguments
TCPDUMP_CMD="tcpdump -w $OUTPUT_FILE $*"

# Run tcpdump with timeout
echo "Running: timeout $TIMEOUT $TCPDUMP_CMD"
if timeout "$TIMEOUT" $TCPDUMP_CMD; then
    echo "Tcpdump capture completed successfully"
else
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 124 ]; then
        echo "Tcpdump capture completed (timeout reached)"
    else
        echo "Tcpdump capture failed with exit code $EXIT_CODE"
        exit $EXIT_CODE
    fi
fi

# Check if output file was created
if [ ! -f "$OUTPUT_FILE" ]; then
    echo "Error: Output file $OUTPUT_FILE was not created"
    exit 1
fi

echo "Capture file created: $OUTPUT_FILE"
ls -lh "$OUTPUT_FILE"

# Send profile data to collector if configured
if [ -n "${COLLECTOR_ENDPOINT:-}" ]; then
    echo "Sending capture to collector..."
    send-profile.sh "$OUTPUT_FILE"
else
    echo "No collector configured, keeping file locally"
fi

echo "Tcpdump capture process completed"
