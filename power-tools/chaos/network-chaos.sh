#!/bin/sh

# Network chaos script - simulate network issues in Kubernetes ephemeral containers
# Usage: network-chaos.sh <duration_sec> [action] [target_host] [port]

DURATION="${1:-30}"
ACTION="${2:-connectivity}"
TARGET_HOST="${3:-8.8.8.8}"
PORT="${4:-80}"

echo "Network Chaos Experiment Started"
echo "Duration: ${DURATION}s"
echo "Action: $ACTION"
echo "Target: $TARGET_HOST:$PORT"
echo "Timestamp: $(date)"

# Since we're in an ephemeral container sharing network namespace,
# we can observe network behavior but have limited ability to modify it

case "$ACTION" in
    "connectivity")
        echo "Testing network connectivity patterns"
        END_TIME=$(($(date +%s) + DURATION))
        SUCCESS=0
        FAILURES=0
        
        while [ $(date +%s) -lt $END_TIME ]; do
            if command -v nc >/dev/null 2>&1; then
                if nc -z -w 2 "$TARGET_HOST" "$PORT" 2>/dev/null; then
                    SUCCESS=$((SUCCESS + 1))
                    echo "$(date): Connection to $TARGET_HOST:$PORT - SUCCESS"
                else
                    FAILURES=$((FAILURES + 1))
                    echo "$(date): Connection to $TARGET_HOST:$PORT - FAILED"
                fi
            elif command -v wget >/dev/null 2>&1; then
                if wget -q --spider --timeout=2 "http://$TARGET_HOST:$PORT" 2>/dev/null; then
                    SUCCESS=$((SUCCESS + 1))
                    echo "$(date): HTTP check to $TARGET_HOST:$PORT - SUCCESS"
                else
                    FAILURES=$((FAILURES + 1))
                    echo "$(date): HTTP check to $TARGET_HOST:$PORT - FAILED"
                fi
            else
                echo "$(date): No network tools available for testing"
                sleep 5
            fi
            sleep 3
        done
        
        echo "Network connectivity summary:"
        echo "Successful connections: $SUCCESS"
        echo "Failed connections: $FAILURES"
        TOTAL=$((SUCCESS + FAILURES))
        if [ $TOTAL -gt 0 ]; then
            FAILURE_RATE=$((FAILURES * 100 / TOTAL))
            echo "Failure rate: ${FAILURE_RATE}%"
        fi
        ;;
    
    "dns")
        echo "Testing DNS resolution patterns"
        END_TIME=$(($(date +%s) + DURATION))
        
        while [ $(date +%s) -lt $END_TIME ]; do
            if command -v nslookup >/dev/null 2>&1; then
                echo "$(date): DNS lookup for $TARGET_HOST"
                nslookup "$TARGET_HOST" 2>&1 | head -10
            elif command -v ping >/dev/null 2>&1; then
                echo "$(date): Ping-based DNS test for $TARGET_HOST"
                ping -c 1 -W 2 "$TARGET_HOST" 2>&1 | head -3
            else
                echo "$(date): No DNS tools available"
            fi
            sleep 10
        done
        ;;
    
    "latency")
        echo "Measuring network latency patterns"
        END_TIME=$(($(date +%s) + DURATION))
        LATENCIES=""
        
        while [ $(date +%s) -lt $END_TIME ]; do
            if command -v ping >/dev/null 2>&1; then
                RESULT=$(ping -c 1 -W 2 "$TARGET_HOST" 2>/dev/null | grep "time=" | sed 's/.*time=\([0-9.]*\).*/\1/')
                if [ -n "$RESULT" ]; then
                    echo "$(date): Latency to $TARGET_HOST: ${RESULT}ms"
                    LATENCIES="$LATENCIES $RESULT"
                else
                    echo "$(date): Ping to $TARGET_HOST failed"
                fi
            else
                echo "$(date): Ping not available for latency measurement"
            fi
            sleep 5
        done
        
        if [ -n "$LATENCIES" ]; then
            echo "Latency measurements collected: $LATENCIES"
        fi
        ;;
    
    "bandwidth")
        echo "Testing bandwidth/throughput patterns"
        END_TIME=$(($(date +%s) + DURATION))
        
        while [ $(date +%s) -lt $END_TIME ]; do
            if command -v wget >/dev/null 2>&1; then
                echo "$(date): Testing download speed from $TARGET_HOST"
                # Try to download a small file to test bandwidth
                START_TIME=$(date +%s)
                wget -q -O /dev/null --timeout=10 "http://$TARGET_HOST/" 2>/dev/null
                END_TIME_DL=$(date +%s)
                DURATION_DL=$((END_TIME_DL - START_TIME))
                echo "Download test completed in ${DURATION_DL}s"
            else
                echo "$(date): No bandwidth testing tools available"
            fi
            sleep 15
        done
        ;;
    
    "monitor")
        echo "Monitoring network interfaces and connections"
        END_TIME=$(($(date +%s) + DURATION))
        
        while [ $(date +%s) -lt $END_TIME ]; do
            echo "$(date): Network interface status:"
            if [ -f /proc/net/dev ]; then
                cat /proc/net/dev | head -5
            fi
            
            echo "$(date): Active connections:"
            if [ -f /proc/net/tcp ]; then
                wc -l /proc/net/tcp
            fi
            
            sleep 10
        done
        ;;
    
    *)
        echo "Unknown network action: $ACTION"
        echo "Available actions: connectivity, dns, latency, bandwidth, monitor"
        echo "Note: Running in ephemeral container - limited network modification capabilities"
        exit 1
        ;;
esac

echo "Network chaos experiment completed at $(date)"
