#!/bin/sh

set -e

# Show help if requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ] || [ "$1" = "help" ]; then
    cat << EOF
Chaos Engineering Power Tool

Usage: 
  chaos [TYPE] [OPTIONS...]

Chaos Types:
  process     - Process manipulation (suspend, kill, restart)
  cpu         - CPU stress testing
  memory      - Memory pressure testing  
  storage     - Storage I/O chaos
  network     - Network connectivity chaos

Environment Variables:
  CHAOS_TYPE     - Type of chaos experiment (default: process)
  DURATION       - Duration of experiment (default: 30s)
  INTERVAL       - Interval between actions (default: 5s)
  TARGET_PID     - Target process ID (default: 1)
  OUTPUT_FILE    - Output file path (default: /tmp/chaos-TIMESTAMP.log)
  COLLECTOR_ENDPOINT - Collector endpoint for sending results

Examples:
  chaos process suspend
  chaos cpu 50
  chaos memory 80
  chaos network connectivity google.com 80
  chaos storage fill /tmp 1G

EOF
    exit 0
fi

# Parse command line arguments
if [ $# -gt 0 ]; then
    CHAOS_TYPE="$1"
    shift
else
    # If no command line args, try to read from TOOL_ARG_* env vars
    if [ -n "${TOOL_ARG_0:-}" ]; then
        CHAOS_TYPE="$TOOL_ARG_0"
        # Build args array from TOOL_ARG_* variables
        i=1
        while true; do
            var_name="TOOL_ARG_$i"
            eval "var_value=\${$var_name:-}"
            if [ -z "$var_value" ]; then
                break
            fi
            set -- "$@" "$var_value"
            i=$((i + 1))
        done
    fi
fi

# Default values
CHAOS_TYPE="${CHAOS_TYPE:-process}"
DURATION="${DURATION:-30s}"
INTERVAL="${INTERVAL:-5s}"
TARGET_PID="${TARGET_PID:-1}"
OUTPUT_FILE="${OUTPUT_FILE:-/tmp/chaos-${TARGET_POD_NAME:-unknown}-${HOSTNAME}-$(date +%Y%m%d-%H%M%S).log}"

echo "Starting chaos engineering experiment..."
echo "Type: $CHAOS_TYPE"
echo "Duration: $DURATION"
echo "Target PID: $TARGET_PID"
echo "Output: $OUTPUT_FILE"

# Parse duration to seconds
parse_duration() {
    local dur="$1"
    case "$dur" in
        *s) echo "${dur%s}" ;;
        *m) echo "$((${dur%m} * 60))" ;;
        *h) echo "$((${dur%h} * 3600))" ;;
        *) echo "$dur" ;;
    esac
}

DURATION_SEC=$(parse_duration "$DURATION")
INTERVAL_SEC=$(parse_duration "$INTERVAL")

# Execute chaos experiment based on type
case "$CHAOS_TYPE" in
    "process")
        echo "Executing process chaos experiment..."
        /chaos/process-chaos.sh "$TARGET_PID" "$DURATION_SEC" "$INTERVAL_SEC" "$@" 2>&1 | tee "$OUTPUT_FILE"
        ;;
    "cpu")
        echo "Executing CPU chaos experiment..."
        /chaos/cpu-chaos.sh "$DURATION_SEC" "$@" 2>&1 | tee "$OUTPUT_FILE"
        ;;
    "storage")
        echo "Executing storage chaos experiment..."
        /chaos/storage-chaos.sh "$DURATION_SEC" "$@" 2>&1 | tee "$OUTPUT_FILE"
        ;;
    "network")
        echo "Executing network chaos experiment..."
        /chaos/network-chaos.sh "$DURATION_SEC" "$@" 2>&1 | tee "$OUTPUT_FILE"
        ;;
    "memory")
        echo "Executing memory chaos experiment..."
        /chaos/memory-chaos.sh "$DURATION_SEC" "$@" 2>&1 | tee "$OUTPUT_FILE"
        ;;
    *)
        echo "Unknown chaos type: $CHAOS_TYPE"
        echo "Available types: process, cpu, storage, network, memory"
        echo "Use 'chaos --help' for more information"
        exit 1
        ;;
esac

echo "Chaos experiment completed"
echo "Results written to: $OUTPUT_FILE"

# Send results to collector if configured
if [ -n "${COLLECTOR_ENDPOINT:-}" ]; then
    echo "Sending results to collector..."
    send-profile.sh "$OUTPUT_FILE"
else
    echo "No collector configured, keeping results locally"
fi
