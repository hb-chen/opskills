#!/bin/bash

# Check if KubeKey is installed
# Returns 0 if installed, 1 if not

set -e

echo "Checking KubeKey installation..."

# Check if kk command exists
if command -v kk &> /dev/null; then
    echo "✓ KubeKey is installed"
    
    # Try to get version
    if kk version &> /dev/null; then
        echo "Version information:"
        kk version
        echo ""
        
        # Try to get cluster info if cluster exists
        if kk cluster-info &> /dev/null; then
            echo "Cluster information:"
            kk cluster-info
        fi
    else
        echo "KubeKey found but version command failed"
    fi
    
    exit 0
else
    echo "✗ KubeKey is not installed"
    echo ""
    echo "To install KubeKey, run:"
    echo "  ./scripts/install_kubekey.sh [VERSION]"
    echo ""
    echo "Or visit: https://github.com/kubesphere/kubekey/releases"
    exit 1
fi

