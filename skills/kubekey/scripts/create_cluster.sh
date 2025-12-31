#!/bin/bash

# Create a Kubernetes cluster using KubeKey
# Usage: ./create_cluster.sh <config-file>

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <config-file>"
    echo "Example: $0 ../examples/cluster-config.yaml"
    exit 1
fi

CONFIG_FILE="$1"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Check if KubeKey is installed
if ! command -v kk &> /dev/null; then
    echo "Error: KubeKey is not installed"
    echo "Please install it first: ./scripts/install_kubekey.sh"
    exit 1
fi

echo "Creating Kubernetes cluster with configuration: $CONFIG_FILE"
echo ""

# Basic YAML syntax check
if ! command -v yq &> /dev/null && ! command -v python3 &> /dev/null; then
    echo "Note: YAML validation skipped (yq or python3 not available)"
else
    # Try to validate YAML syntax
    if command -v yq &> /dev/null; then
        if ! yq eval '.' "$CONFIG_FILE" &> /dev/null; then
            echo "Warning: YAML syntax validation failed"
        fi
    elif command -v python3 &> /dev/null; then
        if ! python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>/dev/null; then
            echo "Warning: YAML syntax validation failed"
        fi
    fi
fi

echo ""

# Create cluster
echo "Starting cluster creation..."
echo "This may take several minutes..."
echo ""

kk create cluster -f "$CONFIG_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Cluster created successfully!"
    echo ""
    echo "To access the cluster, run:"
    echo "  export KUBECONFIG=~/.kube/config"
    echo "  kubectl get nodes"
else
    echo ""
    echo "✗ Cluster creation failed"
    echo "Please check the error messages above"
    exit 1
fi

