# Quickstart

This guide explains how to test **LAs** within your own environment.

- **LAs** will be used in the `latest` version  
- **Cluster Autoscaler** will be used in version `1.30.4`, available for download [here](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.30.4)

There is no maximum number of clusters supported.  
However, for each cluster, you must extract its **resources** and **kubeconfig**, and apply the modifications described below.

## Required Modifications

There are three main changes to apply:

1. **Discovery Server Configuration**  
   For every remote cluster, add the corresponding information inside the [`discovery-server`](../discovery-server/functions) component.  
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

## Launch Instructions

Start the following Go components (in any order):

```
go run discovery.go
go run server.go
go run node-manager.go
```
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

