# KubeKey Configuration Options Reference

This document provides detailed information about all configuration options available in KubeKey cluster configuration files.

## Kubernetes Version

**Field**: `spec.kubernetes.version`

**Options**:
- `v1.28.0` - Latest stable (recommended)
- `v1.27.0` - Previous stable
- `v1.26.0` - Older stable
- `v1.25.0` - Legacy support

**Recommendation**: Use the latest stable version unless you have specific compatibility requirements.

## Container Runtime Interface (CRI)

**Field**: Configured separately, not directly in config file

**Options**:
- **containerd** (recommended)
  - CNCF standard
  - Lightweight and efficient
  - Default for Kubernetes 1.24+
  - Better resource usage

- **docker**
  - Traditional choice
  - Widely used and familiar
  - Requires Docker daemon
  - More resource intensive

- **cri-o**
  - OCI-compliant
  - Lightweight
  - Good for edge deployments

- **isula**
  - Huawei's container runtime
  - Less common

**Recommendation**: Use `containerd` for new clusters unless you have specific requirements.

## Network Plugin (CNI)

**Field**: `spec.network.plugin`

**Options**:

### Calico (Recommended for Production)
- **Best for**: Production environments, network policies
- **Features**:
  - Network policies support
  - BGP routing
  - IP-in-IP or VXLAN encapsulation
  - Good performance
- **Use case**: Production clusters requiring network segmentation

### Flannel
- **Best for**: Simple deployments, small clusters
- **Features**:
  - Simple configuration
  - VXLAN backend
  - Easy to set up
- **Use case**: Development, small clusters, simple networking needs

### Cilium
- **Best for**: High performance, advanced networking
- **Features**:
  - eBPF-based (kernel-level performance)
  - Advanced network policies
  - Service mesh integration
  - High throughput
- **Use case**: High-performance clusters, advanced networking requirements

### Kube-OVN
- **Best for**: Advanced networking features
- **Features**:
  - OVN-based networking
  - Advanced load balancing
  - Network isolation
  - QoS support
- **Use case**: Complex networking requirements

### Weave
- **Best for**: Simple mesh networking
- **Features**:
  - Automatic mesh networking
  - Simple setup
  - Encrypted networking option
- **Use case**: Small to medium clusters, simple requirements

**Recommendation**: 
- Production: `calico`
- Development/Simple: `flannel`
- High Performance: `cilium`

## Network CIDR Configuration

### Pod CIDR
**Field**: `spec.network.kubePodsCIDR`

**Default**: `10.233.64.0/18`

**Calculation Guide**:
- `/18` = 16,384 IPs (~16k pods)
- `/16` = 65,536 IPs (~65k pods)
- `/20` = 4,096 IPs (~4k pods)

**Formula**: 2^(32-mask) - 2 (subtract network and broadcast)

**Recommendation**: 
- Small cluster (< 10 nodes): `/18` or `/20`
- Medium cluster (10-50 nodes): `/18`
- Large cluster (50+ nodes): `/16`

### Service CIDR
**Field**: `spec.network.kubeServiceCIDR`

**Default**: `10.233.0.0/18`

**Important**: Must not overlap with Pod CIDR or node networks.

**Recommendation**: Use `/18` for most cases (16k services).

## Proxy Mode

**Field**: `spec.kubernetes.proxyMode`

**Options**:
- **iptables** (default)
  - Simpler
  - Good for small to medium clusters
  - Standard Linux tool

- **ipvs**
  - Better performance
  - Lower latency
  - Better for large clusters (100+ nodes)
  - Requires kernel modules

**Recommendation**: 
- Small/Medium: `iptables`
- Large: `ipvs`

## Node Configuration

### Host Definition
```yaml
hosts:
- {name: node1, address: 192.168.1.10, internalAddress: 192.168.1.10, user: root, password: "password"}
```

**Fields**:
- `name`: Unique node identifier
- `address`: External/public IP address
- `internalAddress`: Internal/private IP address (same as address if single network)
- `user`: SSH username (typically `root`)
- `password`: SSH password (or use `privateKeyPath` for key-based auth)

### Role Groups

**Master Nodes**:
- Run control plane components (API server, etcd, controller manager, scheduler)
- Minimum: 1 (single master) or 3+ (HA)
- Recommended: 3 for production HA

**Worker Nodes**:
- Run application workloads
- Minimum: 1
- Recommended: 2+ for redundancy

**ETCD Nodes**:
- Store cluster state
- Can be co-located with masters or separate
- For HA: 3 or 5 nodes (odd number)

## Control Plane Endpoint

**Field**: `spec.controlPlaneEndpoint`

**Single Master**:
```yaml
controlPlaneEndpoint:
  domain: ""
  address: ""
  port: 6443
```

**HA with Load Balancer**:
```yaml
controlPlaneEndpoint:
  domain: k8s-api.example.com
  address: ""  # Or specific LB IP
  port: 6443
```

## Registry Configuration

**Private Registry**:
```yaml
registry:
  privateRegistry: "registry.example.com"
  namespaceOverride: "kubesphere"
```

**Registry Mirrors** (for faster downloads):
```yaml
registry:
  registryMirrors:
  - "https://dockerhub.azk8s.cn"
  - "https://reg-mirror.qiniu.com"
```

**Insecure Registries**:
```yaml
registry:
  insecureRegistries:
  - "registry.local:5000"
```

## Advanced Settings

### Max Pods per Node
**Field**: `spec.kubernetes.maxPods`

**Default**: `110`

**Calculation**: Based on available IPs in node CIDR
- Node CIDR `/24`: 254 IPs - 1 = 253 max pods
- Node CIDR `/25`: 126 IPs - 1 = 125 max pods

### Node CIDR Mask Size
**Field**: `spec.kubernetes.nodeCidrMaskSize`

**Default**: `24`

**Effect**: Determines subnet size allocated to each node
- `/24`: 254 IPs per node
- `/25`: 126 IPs per node
- `/26`: 62 IPs per node

## Configuration Checklist

Before creating a cluster, verify:

- [ ] All node IPs are reachable
- [ ] SSH credentials are correct
- [ ] Pod CIDR doesn't overlap with service CIDR
- [ ] CIDRs don't overlap with node networks
- [ ] Kubernetes version is supported
- [ ] CNI plugin is appropriate for use case
- [ ] Control plane endpoint is configured (for HA)
- [ ] Registry settings are correct (if using private registry)
- [ ] Node roles are correctly assigned
- [ ] Sufficient resources on nodes (CPU, memory, disk)

## References

- [KubeKey Documentation](https://kubesphere.io/docs/installing-on-linux/introduction/kubekey/)
- [Kubernetes Networking](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
- [Calico Network Policies](https://docs.tigera.io/calico/latest/network-policy/)
- [Cilium Documentation](https://docs.cilium.io/)
- [Containerd Configuration](https://containerd.io/docs/)
- [Kubernetes Cluster Networking](https://kubernetes.io/docs/concepts/cluster-administration/networking/)

## Example Configurations

### Minimal Development Cluster
```yaml
kubernetes:
  version: v1.28.0
network:
  plugin: flannel
  kubePodsCIDR: 10.233.64.0/18
  kubeServiceCIDR: 10.233.0.0/18
kubernetes:
  proxyMode: iptables
```

### Production Cluster
```yaml
kubernetes:
  version: v1.28.0
network:
  plugin: calico
  kubePodsCIDR: 10.233.64.0/18
  kubeServiceCIDR: 10.233.0.0/18
kubernetes:
  proxyMode: ipvs
  maxPods: 110
```

### High Performance Cluster
```yaml
kubernetes:
  version: v1.28.0
network:
  plugin: cilium
  kubePodsCIDR: 10.233.64.0/16
  kubeServiceCIDR: 10.233.0.0/18
kubernetes:
  proxyMode: ipvs
  maxPods: 250
```

