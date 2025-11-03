#!/bin/sh

# CPU chaos script - spike CPU usage
# Usage: cpu-chaos.sh <duration_sec> [cpu_percent]

DURATION="${1:-30}"
CPU_PERCENT="${2:-80}"
CORES=$(nproc 2>/dev/null || echo "1")

echo "CPU Chaos Experiment Started"
echo "Duration: ${DURATION}s"
echo "Target CPU: ${CPU_PERCENT}%"
echo "Available cores: $CORES"
echo "Timestamp: $(date)"

# Calculate number of workers based on CPU percentage
# stress-ng will distribute load across workers
WORKERS=$((CORES * CPU_PERCENT / 100))
if [ $WORKERS -lt 1 ]; then
    WORKERS=1
fi

echo "Starting stress-ng with $WORKERS CPU workers..."

# Use stress-ng for CPU stress
stress-ng --cpu $WORKERS --timeout ${DURATION}s --verbose

echo "CPU chaos experiment completed at $(date)"
