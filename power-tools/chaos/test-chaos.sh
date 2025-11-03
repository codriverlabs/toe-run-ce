#!/bin/sh

# Test script for chaos engineering power tool
# Usage: test-chaos.sh [test_type]

TEST_TYPE="${1:-all}"
TEST_DURATION="10s"  # Short duration for testing

echo "Chaos Engineering Power Tool Test Suite"
echo "========================================"
echo "Test type: $TEST_TYPE"
echo "Test duration: $TEST_DURATION"
echo "Timestamp: $(date)"
echo

# Test process chaos
test_process_chaos() {
    echo "Testing Process Chaos..."
    echo "------------------------"
    
    # Test suspend action (safest for testing)
    echo "Test 1: Process suspend/resume"
    CHAOS_TYPE=process DURATION=$TEST_DURATION ./entrypoint.sh suspend
    echo
    
    # Test with different target PID (current shell)
    echo "Test 2: Process chaos with specific PID"
    TARGET_PID=$$ CHAOS_TYPE=process DURATION=$TEST_DURATION ./entrypoint.sh suspend
    echo
}

# Test CPU chaos
test_cpu_chaos() {
    echo "Testing CPU Chaos..."
    echo "--------------------"
    
    echo "Test 1: CPU stress at 50%"
    CHAOS_TYPE=cpu DURATION=$TEST_DURATION ./entrypoint.sh 50
    echo
    
    echo "Test 2: CPU stress at 80%"
    CHAOS_TYPE=cpu DURATION=$TEST_DURATION ./entrypoint.sh 80
    echo
}

# Test storage chaos
test_storage_chaos() {
    echo "Testing Storage Chaos..."
    echo "------------------------"
    
    echo "Test 1: Storage fill at 10% (safe for testing)"
    CHAOS_TYPE=storage DURATION=$TEST_DURATION ./entrypoint.sh 10 /tmp
    echo
    
    echo "Test 2: Storage monitoring"
    CHAOS_TYPE=storage DURATION=$TEST_DURATION ./entrypoint.sh 5 /tmp
    echo
}

# Test network chaos
test_network_chaos() {
    echo "Testing Network Chaos..."
    echo "------------------------"
    
    echo "Test 1: Network connectivity test"
    CHAOS_TYPE=network DURATION=$TEST_DURATION ./entrypoint.sh connectivity google.com 80
    echo
    
    echo "Test 2: DNS resolution test"
    CHAOS_TYPE=network DURATION=$TEST_DURATION ./entrypoint.sh dns google.com
    echo
    
    echo "Test 3: Network latency test"
    CHAOS_TYPE=network DURATION=$TEST_DURATION ./entrypoint.sh latency google.com
    echo
    
    echo "Test 4: Network monitoring"
    CHAOS_TYPE=network DURATION=$TEST_DURATION ./entrypoint.sh monitor
    echo
}

# Test memory chaos
test_memory_chaos() {
    echo "Testing Memory Chaos..."
    echo "-----------------------"
    
    echo "Test 1: Memory pressure - linear pattern"
    CHAOS_TYPE=memory DURATION=$TEST_DURATION ./entrypoint.sh pressure 20 linear
    echo
    
    echo "Test 2: Memory pressure - spike pattern"
    CHAOS_TYPE=memory DURATION=$TEST_DURATION ./entrypoint.sh pressure 10 spike
    echo
    
    echo "Test 3: Memory monitoring"
    CHAOS_TYPE=memory DURATION=$TEST_DURATION ./entrypoint.sh monitor
    echo
    
    echo "Test 4: Memory fragmentation (small test)"
    CHAOS_TYPE=memory DURATION=$TEST_DURATION ./entrypoint.sh fragmentation 5
    echo
}

# Test environment variable handling
test_environment() {
    echo "Testing Environment Variables..."
    echo "-------------------------------"
    
    echo "Test 1: Custom output file"
    OUTPUT_FILE="/tmp/chaos-test-output.log" CHAOS_TYPE=cpu DURATION=5s ./entrypoint.sh 30
    if [ -f "/tmp/chaos-test-output.log" ]; then
        echo "✓ Custom output file created successfully"
        rm -f "/tmp/chaos-test-output.log"
    else
        echo "✗ Custom output file not created"
    fi
    echo
    
    echo "Test 2: Custom interval"
    INTERVAL=2s CHAOS_TYPE=process DURATION=8s ./entrypoint.sh suspend
    echo
}

# Test error handling
test_error_handling() {
    echo "Testing Error Handling..."
    echo "-------------------------"
    
    echo "Test 1: Invalid chaos type"
    CHAOS_TYPE=invalid DURATION=1s ./entrypoint.sh || echo "✓ Invalid chaos type handled correctly"
    echo
    
    echo "Test 2: Invalid process PID"
    TARGET_PID=99999 CHAOS_TYPE=process DURATION=1s ./entrypoint.sh suspend || echo "✓ Invalid PID handled correctly"
    echo
    
    echo "Test 3: Invalid network action"
    CHAOS_TYPE=network DURATION=1s ./entrypoint.sh invalid_action || echo "✓ Invalid network action handled correctly"
    echo
}

# Run tests based on type
case "$TEST_TYPE" in
    "process")
        test_process_chaos
        ;;
    "cpu")
        test_cpu_chaos
        ;;
    "storage")
        test_storage_chaos
        ;;
    "network")
        test_network_chaos
        ;;
    "memory")
        test_memory_chaos
        ;;
    "environment")
        test_environment
        ;;
    "errors")
        test_error_handling
        ;;
    "all")
        test_process_chaos
        test_cpu_chaos
        test_storage_chaos
        test_network_chaos
        test_memory_chaos
        test_environment
        test_error_handling
        ;;
    *)
        echo "Unknown test type: $TEST_TYPE"
        echo "Available types: process, cpu, storage, network, memory, environment, errors, all"
        exit 1
        ;;
esac

echo "Test Suite Completed"
echo "===================="
echo "Check output files in /tmp/chaos-* for detailed results"
echo "Timestamp: $(date)"
