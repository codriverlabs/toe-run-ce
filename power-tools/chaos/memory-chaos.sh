#!/bin/sh

# Memory chaos script - memory pressure experiments in ephemeral containers
# Usage: memory-chaos.sh <duration_sec> [action] [memory_mb] [pattern]

DURATION="${1:-30}"
ACTION="${2:-pressure}"
MEMORY_MB="${3:-100}"
PATTERN="${4:-linear}"

echo "Memory Chaos Experiment Started"
echo "Duration: ${DURATION}s"
echo "Action: $ACTION"
echo "Memory target: ${MEMORY_MB}MB"
echo "Pattern: $PATTERN"
echo "Timestamp: $(date)"

# Get current memory info
if [ -f /proc/meminfo ]; then
    echo "Initial memory status:"
    grep -E "MemTotal|MemFree|MemAvailable" /proc/meminfo
fi

case "$ACTION" in
    "pressure")
        echo "Creating memory pressure with ${MEMORY_MB}MB allocation"
        
        case "$PATTERN" in
            "linear")
                echo "Linear memory allocation pattern"
                # Allocate memory in chunks over time
                CHUNK_SIZE=10  # 10MB chunks
                CHUNKS=$((MEMORY_MB / CHUNK_SIZE))
                INTERVAL=$((DURATION / CHUNKS))
                if [ $INTERVAL -lt 1 ]; then
                    INTERVAL=1
                fi
                
                i=0
                while [ $i -lt $CHUNKS ]; do
                    # Allocate memory chunk using dd
                    dd if=/dev/zero of="/tmp/mem_chunk_$i" bs=1M count=$CHUNK_SIZE 2>/dev/null &
                    echo "$(date): Allocated chunk $i (${CHUNK_SIZE}MB)"
                    i=$((i + 1))
                    sleep $INTERVAL
                done
                
                # Keep memory allocated for remaining duration
                REMAINING=$((DURATION - (CHUNKS * INTERVAL)))
                if [ $REMAINING -gt 0 ]; then
                    sleep $REMAINING
                fi
                ;;
                
            "spike")
                echo "Memory spike pattern"
                # Allocate all memory at once
                dd if=/dev/zero of="/tmp/mem_spike" bs=1M count=$MEMORY_MB 2>/dev/null &
                SPIKE_PID=$!
                echo "$(date): Memory spike allocated (${MEMORY_MB}MB)"
                sleep $DURATION
                kill $SPIKE_PID 2>/dev/null || true
                ;;
                
            "oscillating")
                echo "Oscillating memory pattern"
                CYCLES=5
                CYCLE_DURATION=$((DURATION / CYCLES))
                
                i=0
                while [ $i -lt $CYCLES ]; do
                    # Allocate
                    dd if=/dev/zero of="/tmp/mem_cycle_$i" bs=1M count=$MEMORY_MB 2>/dev/null &
                    CYCLE_PID=$!
                    echo "$(date): Cycle $i - Memory allocated"
                    sleep $((CYCLE_DURATION / 2))
                    
                    # Release
                    kill $CYCLE_PID 2>/dev/null || true
                    rm -f "/tmp/mem_cycle_$i" 2>/dev/null || true
                    echo "$(date): Cycle $i - Memory released"
                    sleep $((CYCLE_DURATION / 2))
                    
                    i=$((i + 1))
                done
                ;;
        esac
        
        # Cleanup
        echo "Cleaning up memory allocations..."
        rm -f /tmp/mem_chunk_* /tmp/mem_spike /tmp/mem_cycle_* 2>/dev/null || true
        ;;
    
    "leak")
        echo "Simulating memory leak pattern"
        LEAK_RATE=5  # MB per interval
        INTERVAL=5   # seconds
        
        END_TIME=$(($(date +%s) + DURATION))
        LEAK_COUNT=0
        
        while [ $(date +%s) -lt $END_TIME ]; do
            # Create a "leaked" memory file that grows over time
            dd if=/dev/zero of="/tmp/leak_$LEAK_COUNT" bs=1M count=$LEAK_RATE 2>/dev/null &
            echo "$(date): Memory leak simulation - allocated ${LEAK_RATE}MB (total: $((LEAK_COUNT * LEAK_RATE))MB)"
            
            LEAK_COUNT=$((LEAK_COUNT + 1))
            sleep $INTERVAL
        done
        
        echo "Memory leak simulation completed. Cleaning up..."
        rm -f /tmp/leak_* 2>/dev/null || true
        ;;
    
    "fragmentation")
        echo "Simulating memory fragmentation"
        # Allocate many small chunks to fragment memory
        FRAGMENT_SIZE=1  # 1MB fragments
        FRAGMENTS=$((MEMORY_MB / FRAGMENT_SIZE))
        
        echo "Creating $FRAGMENTS memory fragments of ${FRAGMENT_SIZE}MB each"
        
        i=0
        while [ $i -lt $FRAGMENTS ]; do
            dd if=/dev/zero of="/tmp/frag_$i" bs=1M count=$FRAGMENT_SIZE 2>/dev/null &
            
            # Randomly release some fragments to create fragmentation
            if [ $((i % 3)) -eq 0 ] && [ $i -gt 10 ]; then
                RELEASE_IDX=$((i - 5))
                rm -f "/tmp/frag_$RELEASE_IDX" 2>/dev/null || true
            fi
            
            i=$((i + 1))
            
            # Progress indicator
            if [ $((i % 20)) -eq 0 ]; then
                echo "$(date): Created $i/$FRAGMENTS fragments"
            fi
            
            sleep 0.1
        done
        
        echo "Fragmentation pattern active for ${DURATION}s"
        sleep $DURATION
        
        echo "Cleaning up fragmentation test..."
        rm -f /tmp/frag_* 2>/dev/null || true
        ;;
    
    "monitor")
        echo "Monitoring memory usage patterns"
        END_TIME=$(($(date +%s) + DURATION))
        
        while [ $(date +%s) -lt $END_TIME ]; do
            echo "$(date): Memory status:"
            if [ -f /proc/meminfo ]; then
                grep -E "MemTotal|MemFree|MemAvailable|Buffers|Cached" /proc/meminfo
            fi
            
            # Show process memory usage if available
            if [ -f /proc/self/status ]; then
                echo "Current process memory:"
                grep -E "VmSize|VmRSS|VmPeak" /proc/self/status
            fi
            
            sleep 10
        done
        ;;
    
    *)
        echo "Unknown memory action: $ACTION"
        echo "Available actions: pressure, leak, fragmentation, monitor"
        echo "Available patterns: linear, spike, oscillating"
        exit 1
        ;;
esac

# Final memory status
if [ -f /proc/meminfo ]; then
    echo "Final memory status:"
    grep -E "MemTotal|MemFree|MemAvailable" /proc/meminfo
fi

echo "Memory chaos experiment completed at $(date)"
