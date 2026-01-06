# Quickstart (Prototype)

This guide explains how to test **LAs** within your own environment.

- **LAs** will be used in the `latest` version  
- **Cluster Autoscaler** will be used in version `1.30.4`, available for download [here](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.30.4)

There is no maximum number of clusters supported.  
However, for each cluster, you must extract its **resources** and **kubeconfig**, and apply the modifications described below.

## Required Modifications on code

There are three main changes to apply:

1. **Discovery Server Configuration**  
   For every remote cluster, add the corresponding information inside a secret named as the remote cluster, as shown in the secret.yaml example in this folder.  
   > The automatic agent that gathers this data is not implemented yet.

2. **Node Manager Templates**  
   Inside the [`node-manager`](../node-manager/util/) component, you can define **node group templates** based on your requirements.  
   Since this is a PoC, only the following resources can currently vary:
   - CPU  
   - GPU  
   - Number of nodes  
   - RAM  

3. **Local Node Name**  
   Change the name of the local node to your own, so that it is **not** managed by **Cluster Autoscaler** during scaling operations.

## Installing Instructions

There are two ways to deploy the three components:

### 1. Using Helm
From each component’s chart folder, run:

```bash
helm install server-grpc .
helm install node-manager .
helm install discovery .
```
## 2. Using Kubernetes YAML

Apply the manifests directly with `kubectl`:

```bash
kubectl apply -f deployment.yaml
kubectl apply -f rbac.yaml
```

**NOTE**: for testing purpose each component uses the same secret, named tls-cert-secret. It could be generated from the san.cnf in the certifacte folder.


Once all three components are running, launch Cluster Autoscaler v1.30.4 (preferably the binary version). The following is the most basic command:
 `cluster-autoscaler --cloud-provider=externalgrpc --cloud-config=configGrpc.yaml --kubeconfig=path/to/kubeconfig`

Where:
1. **--cloud-provider=externalgrpc** tells the Autoscaler to which it needs to connect, in this case the gRPC server.
2. **--kubeconfig=path/to/kubeconfig** specifies which cluster the Autoscaler should manage.
3. **--cloud-config=configGrpc.yaml** provides additional configuration options. the only required field is the endpoint where the Autoscaler connects to the gRPC server, as shown in the example below.

Example of `configGrpc.yaml`:
```
address: "localhost:9007"
grpc_timeout: 4m
```

