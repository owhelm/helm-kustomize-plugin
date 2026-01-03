#!/bin/bash
set -euo pipefail

# Integration test for helm-kustomize-plugin
# Tests the plugin against the simple-app example

echo "Running integration test..."

# Check if yq is installed
if ! command -v yq &> /dev/null; then
  echo "Error: yq is not installed. Please install yq to run integration tests."
  echo "See: https://github.com/mikefarah/yq#install"
  exit 1
fi

# Ensure plugin is installed
if ! helm plugin list | grep -q kustomize; then
  echo "Error: kustomize plugin not installed"
  echo "Run 'make install' first"
  exit 1
fi

# Run helm template with the plugin
echo "Testing simple-app example..."
OUTPUT=$(helm template examples/simple-app --post-renderer kustomize)

# Check for expected transformations
FAILED=0

# Check 1: Deployment should have 3 replicas (patched from 1)
REPLICAS=$(echo "$OUTPUT" | yq eval 'select(.kind == "Deployment") | .spec.replicas' -)
if [ "$REPLICAS" = "3" ]; then
  echo "✓ Deployment replicas patched to 3"
else
  echo "✗ FAILED: Deployment replicas not set to 3 (got: $REPLICAS)"
  FAILED=1
fi

# Check 2: Deployment should have environment: production label
ENV_LABEL=$(echo "$OUTPUT" | yq eval 'select(.kind == "Deployment") | .metadata.labels.environment' -)
if [ "$ENV_LABEL" = "production" ]; then
  echo "✓ Environment label added"
else
  echo "✗ FAILED: Environment label not set to production (got: $ENV_LABEL)"
  FAILED=1
fi

# Check 3: KustomizePluginData resource should NOT be in output
PLUGIN_DATA_COUNT=$(echo "$OUTPUT" | yq eval 'select(.kind == "KustomizePluginData") | .kind' - | wc -l | tr -d ' ')
if [ "$PLUGIN_DATA_COUNT" = "0" ]; then
  echo "✓ KustomizePluginData resource removed from output"
else
  echo "✗ FAILED: KustomizePluginData resource found in output (should be removed)"
  FAILED=1
fi

# Check 4: Service should still be present
SERVICE_COUNT=$(echo "$OUTPUT" | yq eval 'select(.kind == "Service") | .kind' - | wc -l | tr -d ' ')
if [ "$SERVICE_COUNT" -gt "0" ]; then
  echo "✓ Service resource present"
else
  echo "✗ FAILED: Service resource not found"
  FAILED=1
fi

# Check 5: Deployment should still be present
DEPLOYMENT_COUNT=$(echo "$OUTPUT" | yq eval 'select(.kind == "Deployment") | .kind' - | wc -l | tr -d ' ')
if [ "$DEPLOYMENT_COUNT" -gt "0" ]; then
  echo "✓ Deployment resource present"
else
  echo "✗ FAILED: Deployment resource not found"
  FAILED=1
fi

echo ""
if [ $FAILED -eq 0 ]; then
  echo "✅ All integration tests passed!"
  exit 0
else
  echo "❌ Some tests failed"
  exit 1
fi
