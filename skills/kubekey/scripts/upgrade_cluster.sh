#!/bin/bash

# Upgrade a Kubernetes cluster using KubeKey
# Usage: ./upgrade_cluster.sh [options]
# Options:
#   --k8s-version <version>    Target Kubernetes version (e.g., v1.28.0)
#   --ks-version <version>      Target KubeSphere version (optional)
#   --config <file>             Use configuration file for upgrade
#   --with-kubesphere           Upgrade KubeSphere along with Kubernetes

set -e

K8S_VERSION=""
KS_VERSION=""
CONFIG_FILE=""
WITH_KS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --k8s-version)
            K8S_VERSION="$2"
            shift 2
            ;;
        --ks-version)
            KS_VERSION="$2"
            WITH_KS=true
            shift 2
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --with-kubesphere)
            WITH_KS=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--k8s-version <version>] [--ks-version <version>] [--config <file>] [--with-kubesphere]"
            exit 1
            ;;
    esac
done

# Check if KubeKey is installed
if ! command -v kk &> /dev/null; then
    echo "Error: KubeKey is not installed"
    echo "Please install it first: ./scripts/install_kubekey.sh"
    exit 1
fi

echo "KubeKey Cluster Upgrade"
echo "======================"
echo ""

# Show current cluster status
echo "Current cluster status:"
if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null; then
    echo "Kubernetes version:"
    kubectl version --short 2>/dev/null || echo "  Unable to get version"
    echo ""
    echo "Nodes:"
    kubectl get nodes
    echo ""
else
    echo "Warning: Could not get cluster status (kubectl not available or cluster not accessible)"
    echo ""
fi

# Interactive mode if no arguments provided
if [ -z "$K8S_VERSION" ] && [ -z "$CONFIG_FILE" ]; then
    echo "Interactive upgrade mode"
    echo "----------------------"
    
    # Get current Kubernetes version
    CURRENT_K8S=$(kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print $3}' || echo "unknown")
    echo "Current Kubernetes version: $CURRENT_K8S"
    echo ""
    
    # Ask for target Kubernetes version
    read -p "Target Kubernetes version (e.g., v1.28.0): " K8S_VERSION
    if [ -z "$K8S_VERSION" ]; then
        echo "Error: Kubernetes version is required"
        exit 1
    fi
    
    # Ask about KubeSphere
    read -p "Upgrade KubeSphere? (yes/no) [default: no]: " UPGRADE_KS
    if [ "${UPGRADE_KS,,}" = "yes" ] || [ "${UPGRADE_KS,,}" = "y" ]; then
        WITH_KS=true
        read -p "Target KubeSphere version (e.g., v3.4.1): " KS_VERSION
    fi
    
    echo ""
fi

# Validate Kubernetes version format
if [ -n "$K8S_VERSION" ] && [[ ! "$K8S_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    echo "Warning: Kubernetes version should be in format vX.Y.Z (e.g., v1.28.0)"
    read -p "Continue anyway? (yes/no): " CONTINUE
    if [ "${CONTINUE,,}" != "yes" ] && [ "${CONTINUE,,}" != "y" ]; then
        exit 0
    fi
fi

# Build upgrade command
UPGRADE_CMD="kk upgrade --with-kubernetes --kubernetes-version $K8S_VERSION"

if [ "$WITH_KS" = true ]; then
    if [ -n "$KS_VERSION" ]; then
        UPGRADE_CMD="$UPGRADE_CMD --with-kubesphere --kubesphere-version $KS_VERSION"
    else
        UPGRADE_CMD="$UPGRADE_CMD --with-kubesphere"
    fi
fi

if [ -n "$CONFIG_FILE" ]; then
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Error: Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
    UPGRADE_CMD="$UPGRADE_CMD -f $CONFIG_FILE"
fi

# Show upgrade plan
echo "Upgrade Plan:"
echo "-------------"
echo "Kubernetes: $K8S_VERSION"
if [ "$WITH_KS" = true ]; then
    if [ -n "$KS_VERSION" ]; then
        echo "KubeSphere: $KS_VERSION"
    else
        echo "KubeSphere: latest compatible version"
    fi
fi
echo ""

# Confirm upgrade
read -p "Proceed with upgrade? (yes/no): " CONFIRM
if [ "${CONFIRM,,}" != "yes" ] && [ "${CONFIRM,,}" != "y" ]; then
    echo "Upgrade cancelled"
    exit 0
fi

echo ""
echo "Starting cluster upgrade..."
echo "This may take 30-60 minutes depending on cluster size..."
echo ""

# Execute upgrade
eval $UPGRADE_CMD

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Cluster upgraded successfully!"
    echo ""
    echo "Updated cluster status:"
    kubectl get nodes
    echo ""
    kubectl version --short
else
    echo ""
    echo "✗ Cluster upgrade failed"
    echo "Please check the error messages above"
    echo ""
    echo "Troubleshooting:"
    echo "  - Check node connectivity and resources"
    echo "  - Verify version compatibility"
    echo "  - Review KubeKey logs for details"
    exit 1
fi

