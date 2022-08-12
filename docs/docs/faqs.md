---
sidebar_position: 8
sidebar_label: "FAQs"
---

# FAQs

### Q1: How does HwameiStor scheduler work in a Kubernetes platform? 

The HwameiStor scheduler is deployed as a pod in the HwameiStor namespace.

![img](img/clip_image002.png)

Once the applications (Deployment or StatefulSet) are created, the pod will be scheduled to the worker nodes on which HwameiStor is already configured.

### Q2: How does HwameiStor schedule applications with multi-replicas workloads and what are the differences compared to the traditional shared storage (NFS / block)?

We strongly recommend using StatefulSet for applications with multi-replica workloads.

StatefulSet will deploy replicas on the same worker node with the original pod, and will also create a PV data volume for each replica. If you need to deploy replicas on different worker nodes, you shall manually configure them with `pod affinity`.

![img](img/clip_image004.png)

We suggest using a single pod for deployment because the block data volumes can not be shared.

**For the traditional shared storage:**

StatefulSet will deploy replicas to other worker nodes for workload distribution and will also create a PV data volume for each replica.

Deployment will also deploy replicas to other worker nodes for workload distribution but will share the same PV data volume (only for NFS). We suggest using a single pod for block storage because the block data volumes can not be shared.
