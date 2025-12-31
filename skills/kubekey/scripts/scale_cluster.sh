#!/bin/bash

# Scale a Kubernetes cluster using KubeKey
# This is a helper script that guides you through adding or removing nodes
# Usage: ./scale_cluster.sh [add|delete] [config-file|node-name]

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <operation> [arguments]"
    echo ""
    echo "Operations:"
    echo "  add <config-file>    - Add nodes to cluster"
    echo "  delete <node-name>   - Delete a node from cluster"
    echo ""
    echo "Examples:"
    echo "  $0 add cluster-config.yaml"
    echo "  $0 delete worker1"
    echo ""
    echo "For more control, use:"
    echo "  ./add_nodes.sh <config-file>"
    echo "  ./delete_node.sh <node-name>"
    exit 1
fi

OPERATION="$1"

case "$OPERATION" in
    add)
        if [ $# -lt 2 ]; then
            echo "Error: Config file required for 'add' operation"
            echo "Usage: $0 add <config-file>"
            exit 1
        fi
        exec "$(dirname "$0")/add_nodes.sh" "$2"
        ;;
    delete)
        if [ $# -lt 2 ]; then
            echo "Error: Node name required for 'delete' operation"
            echo "Usage: $0 delete <node-name>"
            exit 1
        fi
        exec "$(dirname "$0")/delete_node.sh" "$2"
        ;;
    *)
        echo "Error: Unknown operation: $OPERATION"
        echo "Valid operations: add, delete"
        exit 1
        ;;
esac

