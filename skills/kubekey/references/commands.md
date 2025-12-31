# KubeKey Commands Reference

This document provides a quick reference for KubeKey (kk) commands.

## Basic Commands

### Check Version
```bash
kk version
```
Display KubeKey client version information.

### Cluster Information
```bash
kk cluster-info
```
Display cluster information.

## Cluster Creation

### Create Cluster Configuration
```bash
kk create config --from-cluster -f <output-file>
```
Generate a configuration file from an existing cluster.

### Create Cluster
```bash
kk create cluster -f <config-file>
```
Create a Kubernetes cluster using the specified configuration file.

## Cluster Management

### Add Nodes
```bash
kk add nodes -f <config-file>
```
Add nodes to an existing cluster. The config file must include all existing nodes plus new ones.

### Delete Node
```bash
kk delete node <node-name>
```
Delete a node from the cluster. Ensure the node is drained first.

### Delete Cluster
```bash
kk delete cluster
```
Delete the entire cluster.

## Cluster Upgrade

### Upgrade Cluster
```bash
kk upgrade --with-kubernetes --kubernetes-version <version>
```
Upgrade Kubernetes to the specified version.

```bash
kk upgrade --with-kubernetes --with-kubesphere \
  --kubernetes-version <k8s-version> \
  --kubesphere-version <ks-version>
```
Upgrade both Kubernetes and KubeSphere.

## Environment Initialization

### Initialize Environment
```bash
kk init
```
Initialize the installation environment.

## Certificate Management

### Cluster Certificates
```bash
kk certs
```
Manage cluster certificates.

## Artifact Management

### Manage Artifacts
```bash
kk artifact
```
Manage KubeKey offline installation packages.

## Help

### Get Help
```bash
kk help
kk <command> --help
```
Get help for KubeKey or specific commands.

