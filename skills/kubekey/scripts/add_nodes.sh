#!/bin/bash

# Add nodes to a Kubernetes cluster using KubeKey
# Usage: ./add_nodes.sh <config-file>
# The config file should include all existing nodes plus the new nodes to add

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <config-file>"
    echo "Example: $0 ../examples/cluster-config.yaml"
    echo ""
    echo "Note: The config file should include:"
    echo "  - All existing nodes (from current cluster)"
    echo "  - New nodes to be added"
    echo ""
    echo "To get current cluster config, run:"
    echo "  kk create config --from-cluster -f <output-file>"
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

echo "Adding nodes to Kubernetes cluster with configuration: $CONFIG_FILE"
echo ""

# Show current cluster status
echo "Current cluster status:"
kubectl get nodes 2>/dev/null || echo "Warning: Could not get cluster status"
echo ""

# Add nodes
echo "Starting to add nodes..."
echo "This may take several minutes..."
echo ""

kk add nodes -f "$CONFIG_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Nodes added successfully!"
    echo ""
    echo "Updated cluster status:"
    kubectl get nodes
else
    echo ""
    echo "✗ Failed to add nodes"
    echo "Please check the error messages above"
    exit 1
fi

