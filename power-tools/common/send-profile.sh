#!/bin/sh

# Helper script to send profile data to collector
# Usage: send-profile.sh <profile-file>

if [ $# -ne 1 ]; then
    echo "Usage: $0 <profile-file>"
    exit 1
fi

PROFILE_FILE="$1"
FILENAME=$(basename "$PROFILE_FILE")

# Validate required environment variables
if [ -z "$COLLECTOR_TOKEN" ]; then
    echo "Error: COLLECTOR_TOKEN not set"
    exit 1
fi

if [ -z "$POWERTOOL_JOB_ID" ]; then
    echo "Error: POWERTOOL_JOB_ID not set"
    exit 1
fi

if [ -z "$TARGET_NAMESPACE" ]; then
    echo "Error: TARGET_NAMESPACE not set"
    exit 1
fi

# Handle TLS certificate
CURL_OPTS=""
if [ -n "$COLLECTOR_CA_CERT" ]; then
    # Use provided CA certificate
    CA_CERT_FILE=$(mktemp)
    echo "$COLLECTOR_CA_CERT" > "$CA_CERT_FILE"
    CURL_OPTS="--cacert $CA_CERT_FILE"
elif [[ "$COLLECTOR_ENDPOINT" == https://* ]]; then
    # For HTTPS without CA cert, skip verification (not recommended for production)
    echo "Warning: No CA certificate provided, skipping TLS verification"
    CURL_OPTS="--insecure"
fi

# Send profile data with metadata headers
curl -X POST \
    $CURL_OPTS \
    -H "Authorization: Bearer $COLLECTOR_TOKEN" \
    -H "X-PowerTool-Job-ID: $POWERTOOL_JOB_ID" \
    -H "X-PowerTool-Namespace: $TARGET_NAMESPACE" \
    -H "X-PowerTool-Matching-Labels: ${POD_MATCHING_LABELS:-unknown}" \
    -H "X-PowerTool-Filename: $FILENAME" \
    -H "Content-Type: application/octet-stream" \
    --data-binary "@$PROFILE_FILE" \
    "$COLLECTOR_ENDPOINT/api/v1/profile"

CURL_EXIT_CODE=$?

# Clean up temporary CA cert file
if [ -n "$CA_CERT_FILE" ]; then
    rm -f "$CA_CERT_FILE"
fi

exit $CURL_EXIT_CODE
