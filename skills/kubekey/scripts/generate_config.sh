#!/bin/bash

# Interactive script to generate KubeKey cluster configuration
# Usage: ./generate_config.sh [output-file]

set -e

OUTPUT_FILE="${1:-cluster-config.yaml}"

echo "KubeKey Cluster Configuration Generator"
echo "======================================"
echo ""

# Cluster name
read -p "Cluster name [default: sample]: " CLUSTER_NAME
CLUSTER_NAME="${CLUSTER_NAME:-sample}"

# Kubernetes version
echo ""
echo "Kubernetes Version Options:"
echo "  1) v1.28.0 (latest stable, recommended)"
echo "  2) v1.27.0"
echo "  3) v1.26.0"
echo "  4) Custom version"
read -p "Select Kubernetes version [1-4, default: 1]: " K8S_VER_CHOICE
case $K8S_VER_CHOICE in
    2) K8S_VERSION="v1.27.0" ;;
    3) K8S_VERSION="v1.26.0" ;;
    4) read -p "Enter Kubernetes version (e.g., v1.25.0): " K8S_VERSION ;;
    *) K8S_VERSION="v1.28.0" ;;
esac

# CNI Plugin
echo ""
echo "CNI Plugin Options:"
echo "  1) calico (recommended for production, supports network policies)"
echo "  2) flannel (simple, good for small clusters)"
echo "  3) cilium (eBPF-based, high performance)"
echo "  4) kube-ovn (advanced networking features)"
echo "  5) weave (simple, automatic mesh)"
read -p "Select CNI plugin [1-5, default: 1]: " CNI_CHOICE
case $CNI_CHOICE in
    2) CNI_PLUGIN="flannel" ;;
    3) CNI_PLUGIN="cilium" ;;
    4) CNI_PLUGIN="kube-ovn" ;;
    5) CNI_PLUGIN="weave" ;;
    *) CNI_PLUGIN="calico" ;;
esac

# CRI
echo ""
echo "Container Runtime Interface (CRI) Options:"
echo "  1) containerd (recommended, CNCF standard)"
echo "  2) docker (traditional, widely used)"
echo "  3) cri-o"
read -p "Select CRI [1-3, default: 1]: " CRI_CHOICE
case $CRI_CHOICE in
    2) CRI="docker" ;;
    3) CRI="cri-o" ;;
    *) CRI="containerd" ;;
esac

# Network CIDRs
echo ""
read -p "Pod CIDR [default: 10.233.64.0/18]: " POD_CIDR
POD_CIDR="${POD_CIDR:-10.233.64.0/18}"

read -p "Service CIDR [default: 10.233.0.0/18]: " SERVICE_CIDR
SERVICE_CIDR="${SERVICE_CIDR:-10.233.0.0/18}"

# Proxy mode
echo ""
read -p "Proxy mode [iptables/ipvs, default: ipvs]: " PROXY_MODE
PROXY_MODE="${PROXY_MODE:-ipvs}"

# Control plane endpoint
echo ""
read -p "Control plane endpoint domain (leave empty for single master): " CP_DOMAIN

# Nodes
echo ""
echo "Node Configuration"
echo "------------------"
read -p "Number of master nodes: " MASTER_COUNT
MASTER_COUNT=${MASTER_COUNT:-1}

read -p "Number of worker nodes: " WORKER_COUNT
WORKER_COUNT=${WORKER_COUNT:-2}

HOSTS=()
ROLE_GROUPS_ETCD=()
ROLE_GROUPS_MASTER=()
ROLE_GROUPS_WORKER=()

# Master nodes
for i in $(seq 1 $MASTER_COUNT); do
    echo ""
    echo "Master Node $i:"
    read -p "  Node name [default: master$i]: " NODE_NAME
    NODE_NAME="${NODE_NAME:-master$i}"
    read -p "  External IP address: " EXTERNAL_IP
    read -p "  Internal IP address [default: $EXTERNAL_IP]: " INTERNAL_IP
    INTERNAL_IP="${INTERNAL_IP:-$EXTERNAL_IP}"
    read -p "  SSH user [default: root]: " SSH_USER
    SSH_USER="${SSH_USER:-root}"
    read -sp "  SSH password: " SSH_PASS
    echo ""
    
    HOSTS+=("  - {name: $NODE_NAME, address: $EXTERNAL_IP, internalAddress: $INTERNAL_IP, user: $SSH_USER, password: \"$SSH_PASS\"}")
    ROLE_GROUPS_ETCD+=("    - $NODE_NAME")
    ROLE_GROUPS_MASTER+=("    - $NODE_NAME")
done

# Worker nodes
for i in $(seq 1 $WORKER_COUNT); do
    echo ""
    echo "Worker Node $i:"
    read -p "  Node name [default: worker$i]: " NODE_NAME
    NODE_NAME="${NODE_NAME:-worker$i}"
    read -p "  External IP address: " EXTERNAL_IP
    read -p "  Internal IP address [default: $EXTERNAL_IP]: " INTERNAL_IP
    INTERNAL_IP="${INTERNAL_IP:-$EXTERNAL_IP}"
    read -p "  SSH user [default: root]: " SSH_USER
    SSH_USER="${SSH_USER:-root}"
    read -sp "  SSH password: " SSH_PASS
    echo ""
    
    HOSTS+=("  - {name: $NODE_NAME, address: $EXTERNAL_IP, internalAddress: $INTERNAL_IP, user: $SSH_USER, password: \"$SSH_PASS\"}")
    ROLE_GROUPS_WORKER+=("    - $NODE_NAME")
done

# Registry
echo ""
read -p "Private registry URL (leave empty if not using): " PRIVATE_REGISTRY

# Generate config file
cat > "$OUTPUT_FILE" <<EOF
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: $CLUSTER_NAME
spec:
  hosts:
$(printf '%s\n' "${HOSTS[@]}")
  roleGroups:
    etcd:
$(printf '%s\n' "${ROLE_GROUPS_ETCD[@]}")
    master:
$(printf '%s\n' "${ROLE_GROUPS_MASTER[@]}")
    worker:
$(printf '%s\n' "${ROLE_GROUPS_WORKER[@]}")
  controlPlaneEndpoint:
    domain: $CP_DOMAIN
    address: ""
    port: 6443
  kubernetes:
    version: $K8S_VERSION
    imageRepo: kubesphere
    clusterName: cluster.local
    masqueradeAll: false
    maxPods: 110
    nodeCidrMaskSize: 24
    proxyMode: $PROXY_MODE
  network:
    plugin: $CNI_PLUGIN
    kubePodsCIDR: $POD_CIDR
    kubeServiceCIDR: $SERVICE_CIDR
    multusCNI:
      enabled: false
  registry:
    privateRegistry: "$PRIVATE_REGISTRY"
    namespaceOverride: ""
    registryMirrors: []
    insecureRegistries: []
  addons: []
EOF

if [ -n "$CRI" ] && [ "$CRI" != "containerd" ]; then
    echo ""
    echo "Note: CRI setting ($CRI) needs to be configured separately in KubeKey."
    echo "Refer to KubeKey documentation for CRI configuration."
fi

echo ""
echo "✓ Configuration file generated: $OUTPUT_FILE"
echo ""
echo "Review the configuration and then create the cluster with:"
echo "  kk create cluster -f $OUTPUT_FILE"

