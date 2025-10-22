# 🚀 Quickstart [version 0.0.1]

## 🧩 Introduction
This guide explains how to test **LAs** within your own environment.

- **LAs** will be used in the `latest` version  
- **Cluster Autoscaler** will be used in version `1.30.4`, available for download [here](#)

There is no maximum number of clusters supported.  
However, for each cluster, you must extract its **resources** and **kubeconfig**, and apply the modifications described below.

---

## ⚙️ Required Modifications

There are three main changes to apply:

1. **Discovery Server Configuration**  
   For every remote cluster, add the corresponding information inside the [`discovery-server`](./path/to/discovery-server) component.  
   > The automatic agent that gathers this data is not implemented yet.

2. **Node Manager Templates**  
   Inside the [`node-manager`](./path/to/node-manager) component, you can define **node group templates** based on your requirements.  
   Since this is a PoC, only the following resources can currently vary:
   - CPU  
   - GPU  
   - Number of nodes  
   - RAM  

3. **Local Node Name**  
   Change the name of the local node to your own, so that it is **not** managed by **Cluster Autoscaler** during scaling operations.

---

## 🧠 Launch Instructions

Start the following Go components (in any order):

```bash
go run discovery.go
go run server.go
go run node-manager.go
```
Once all three components are running, launch Cluster Autoscaler v1.30.4 (preferably the binary version) on the cluster you want to manage.