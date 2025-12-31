# KubeKey Skill

A skill for managing Kubernetes clusters using [KubeKey](https://github.com/kubesphere/kubekey), a tool for deploying Kubernetes clusters efficiently.

## Features

- ✅ Check KubeKey installation and version
- ✅ Install KubeKey with specified version
- ✅ **Generate cluster configurations** based on user requirements
  - Interactive configuration generator
  - Support for all major options (K8s version, CNI, CRI, etc.)
  - Smart defaults and recommendations
- ✅ Create Kubernetes clusters from configuration files
- ✅ Scale clusters by adding or removing nodes
- ✅ **Upgrade clusters** to newer Kubernetes/KubeSphere versions
- ✅ View and analyze cluster configurations

## Quick Start

### Check Installation

```bash
./scripts/check_kubekey.sh
```

### Install KubeKey

```bash
# Install latest version
./scripts/install_kubekey.sh

# Install specific version
./scripts/install_kubekey.sh v3.0.0
```

### Generate Cluster Configuration

You can generate a cluster configuration interactively or manually:

**Interactive Generation:**
```bash
./scripts/generate_config.sh [output-file]
```

This will prompt you for:
- Kubernetes version
- CNI plugin (Calico, Flannel, Cilium, etc.)
- CRI (containerd, Docker, etc.)
- Network CIDRs
- Node information (IPs, credentials, roles)
- Advanced options

**Manual Creation:**
1. Review the example: `examples/cluster-config.yaml`
2. See detailed options: `examples/config-options.md`
3. Create your configuration file

### Create a Cluster

1. Generate or create a cluster configuration file
2. Review and validate the configuration
3. Run the create script:

```bash
./scripts/create_cluster.sh cluster-config.yaml
```

### Add Nodes to Cluster

1. Get current cluster configuration:
   ```bash
   kk create config --from-cluster -f current-cluster.yaml
   ```

2. Edit the configuration file to add new nodes to `spec.hosts` and `roleGroups`

3. Add nodes:
   ```bash
   ./scripts/add_nodes.sh current-cluster.yaml
   ```
   Or use: `kk add nodes -f current-cluster.yaml`

### Delete Nodes from Cluster

1. Drain the node (recommended):
   ```bash
   kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
   ```

2. Delete the node:
   ```bash
   ./scripts/delete_node.sh <node-name>
   ```
   Or use: `kk delete node <node-name>`

**Note**: The `scale_cluster.sh` script is a convenience wrapper that calls `add_nodes.sh` or `delete_node.sh` based on the operation.

### Upgrade Cluster

Upgrade your Kubernetes cluster to a newer version:

**Interactive Mode** (Recommended):
```bash
./scripts/upgrade_cluster.sh
```

**Command Line Options**:
```bash
# Upgrade Kubernetes only
./scripts/upgrade_cluster.sh --k8s-version v1.28.0

# Upgrade Kubernetes and KubeSphere
./scripts/upgrade_cluster.sh --k8s-version v1.28.0 --ks-version v3.4.1 --with-kubesphere
```

**Important Notes**:
- Always backup before upgrading
- Upgrade one minor version at a time (e.g., 1.27 → 1.28)
- Ensure all nodes have synchronized time (NTP)
- Perform during low-traffic periods
- Monitor the upgrade process (takes 30-60 minutes)

### View Configuration

```bash
./scripts/show_config.sh examples/cluster-config.yaml
```

## Configuration

Cluster configurations are defined in YAML files. The AI agent can help you generate configurations based on your requirements.

### Key Configuration Options

**Kubernetes Version**:
- Latest stable: `v1.28.0` (recommended)
- Previous versions: `v1.27.0`, `v1.26.0`, etc.

**CNI Plugin** (Network):
- **Calico**: Production-ready, supports network policies (recommended)
- **Flannel**: Simple, good for small clusters
- **Cilium**: eBPF-based, high performance
- **Kube-OVN**: Advanced networking features
- **Weave**: Simple mesh networking

**CRI** (Container Runtime):
- **containerd**: Recommended, CNCF standard, lightweight
- **docker**: Traditional, widely used
- **cri-o**: Lightweight, OCI-compliant

**Network CIDRs**:
- Pod CIDR: Default `10.233.64.0/18` (16k pods)
- Service CIDR: Default `10.233.0.0/18` (16k services)
- Must not overlap with each other or node networks

**Proxy Mode**:
- `iptables`: Simple, good for small/medium clusters
- `ipvs`: Better performance for large clusters (100+ nodes)

### Configuration Files

- `examples/cluster-config.yaml` - Complete example configuration
- `examples/config-options.md` - Detailed reference for all options
- `scripts/generate_config.sh` - Interactive configuration generator

### Configuration Sections

- **hosts**: Define your cluster nodes (IP addresses, credentials, roles)
- **roleGroups**: Assign nodes to etcd, master, and worker roles
- **kubernetes**: Kubernetes version and settings
- **network**: Network plugin and CIDR ranges
- **registry**: Container registry configuration

## Prerequisites

- Linux or macOS
- SSH access to target nodes
- Root or sudo privileges (for installation)
- Network connectivity

## Scripts

All scripts are located in the `scripts/` directory:

- `check_kubekey.sh` - Check if KubeKey is installed
- `install_kubekey.sh` - Install KubeKey tool
- `generate_config.sh` - **Interactive cluster configuration generator**
- `create_cluster.sh` - Create a new Kubernetes cluster
- `add_nodes.sh` - Add nodes to an existing cluster
- `delete_node.sh` - Delete a node from cluster
- `upgrade_cluster.sh` - **Upgrade Kubernetes/KubeSphere cluster**
- `scale_cluster.sh` - Convenience wrapper for add/delete operations
- `show_config.sh` - Display and analyze cluster configuration

## Resources

- [KubeKey GitHub](https://github.com/kubesphere/kubekey)
- [KubeKey Documentation](https://kubesphere.io/docs/installing-on-linux/introduction/kubekey/)
- [Agent Skills Specification](https://agentskills.io/)

## License

MIT

