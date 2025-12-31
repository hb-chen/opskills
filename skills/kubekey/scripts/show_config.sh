#!/bin/bash

# Show and analyze cluster configuration
# Usage: ./show_config.sh <config-file>

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

echo "Cluster Configuration: $CONFIG_FILE"
echo "=================================="
echo ""

# Display full configuration
cat "$CONFIG_FILE"
echo ""

# Try to extract key information if yq is available
if command -v yq &> /dev/null; then
    echo "Configuration Summary:"
    echo "---------------------"
    
    # Kubernetes version
    K8S_VERSION=$(yq eval '.spec.kubernetes.version // "not specified"' "$CONFIG_FILE" 2>/dev/null || echo "N/A")
    echo "Kubernetes Version: $K8S_VERSION"
    
    # Count nodes
    MASTER_COUNT=$(yq eval '.spec.hosts | map(select(.role[] == "master")) | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
    WORKER_COUNT=$(yq eval '.spec.hosts | map(select(.role[] == "worker")) | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
    ETCD_COUNT=$(yq eval '.spec.hosts | map(select(.role[] == "etcd")) | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
    
    echo "Master Nodes: $MASTER_COUNT"
    echo "Worker Nodes: $WORKER_COUNT"
    echo "ETCD Nodes: $ETCD_COUNT"
    
    # Network plugin
    NETWORK_PLUGIN=$(yq eval '.spec.network.plugin // "not specified"' "$CONFIG_FILE" 2>/dev/null || echo "N/A")
    echo "Network Plugin: $NETWORK_PLUGIN"
    
    echo ""
fi

# Show cluster info using kk if available
if command -v kk &> /dev/null; then
    echo "KubeKey Cluster Information:"
    echo "----------------------------"
    kk cluster-info 2>/dev/null || echo "  Unable to get cluster info (cluster may not exist)"
    echo ""
fi

# If kubectl is available and cluster is accessible, show cluster info
if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null; then
    echo "Current Cluster Status (via kubectl):"
    echo "-------------------------------------"
    kubectl get nodes
    echo ""
    kubectl get pods -A | head -20
fi

