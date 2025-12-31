#!/bin/bash

# Delete a node from a Kubernetes cluster using KubeKey
# Usage: ./delete_node.sh <node-name>
# Note: This will delete the node from the cluster. Make sure to drain the node first if needed.

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <node-name>"
    echo "Example: $0 worker1"
    echo ""
    echo "Note: Before deleting a node:"
    echo "  1. Ensure workloads are migrated or stopped"
    echo "  2. Drain the node: kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data"
    echo "  3. Then run this script to remove the node"
    exit 1
fi

NODE_NAME="$1"

# Check if KubeKey is installed
if ! command -v kk &> /dev/null; then
    echo "Error: KubeKey is not installed"
    echo "Please install it first: ./scripts/install_kubekey.sh"
    exit 1
fi

# Check if node exists
if ! kubectl get node "$NODE_NAME" &> /dev/null; then
    echo "Error: Node '$NODE_NAME' not found in cluster"
    echo ""
    echo "Available nodes:"
    kubectl get nodes
    exit 1
fi

echo "Deleting node: $NODE_NAME"
echo ""

# Show current cluster status
echo "Current cluster status:"
kubectl get nodes
echo ""

# Confirm deletion
read -p "Are you sure you want to delete node '$NODE_NAME'? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Deletion cancelled"
    exit 0
fi

# Delete node
echo "Deleting node..."
echo "This may take several minutes..."
echo ""

kk delete node "$NODE_NAME"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Node deleted successfully!"
    echo ""
    echo "Updated cluster status:"
    kubectl get nodes
else
    echo ""
    echo "✗ Failed to delete node"
    echo "Please check the error messages above"
    exit 1
fi

