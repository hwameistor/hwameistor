---
sidebar_position: 1
sidebar_label: "Node Expansion"
---

# Node Expansion

A storage system is usually expected to expand its capacity by adding a new storage node. In HwameiStor, it can be done with the following steps.

## Steps

### 1. Prepare a new storage node

Add the node into the Kubernetes cluster, or select a Kubernetes node. The node should have all the required items described in [Prerequisites](../install/prereq.md).

For example, the new node and disk information are as follows:

- name: k8s-worker-4
- devPath: /dev/sdb
- diskType: SSD disk

After the new node is already added into the Kubernetes cluster, make sure the following HwameiStor pods are already running on this node.

```console
$ kubectl get node
NAME           STATUS   ROLES            AGE     VERSION
k8s-master-1   Ready    master           96d     v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-3   Ready    worker           96d     v1.24.3-2+63243a96d1c393
k8s-worker-4   Ready    worker           1h      v1.24.3-2+63243a96d1c393

$ kubectl -n hwameistor get pod -o wide | grep k8s-worker-4
hwameistor-local-disk-manager-c86g5     2/2     Running   0     19h   10.6.182.105      k8s-worker-4   <none>  <none>
hwameistor-local-storage-s4zbw          2/2     Running   0     19h   192.168.140.82    k8s-worker-4   <none>  <none>

# check if LocalStorageNode exists
$ kubectl get localstoragenode k8s-worker-4
NAME                 IP           ZONE      REGION    STATUS   AGE
k8s-worker-4   10.6.182.103       default   default   Ready    8d
```

### 2. Add the storage node into HwameiStor

Construct the storage pool of the node by adding a LocalStorageClaim CR as below:

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-4
spec:
  nodeName: k8s-worker-4
  description:
    diskType: SSD
EOF
```

### 3. Post check

Finally, check if the node has constructed the storage pool by checking the LocalStorageNode CR:

```console
$ kubectl get localstoragenode k8s-worker-4
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  name: k8s-worker-4
spec:
  hostname: k8s-worker-4
  storageIP: 10.6.182.103
  topogoly:
    region: default
    zone: default
status:
  pools:
    LocalStorage_PoolSSD:
      class: SSD
      disks:
      - capacityBytes: 214744170496
        devPath: /dev/sdb
        state: InUse
        type: SSD
      freeCapacityBytes: 214744170496
      freeVolumeCount: 1000
      name: LocalStorage_PoolSSD
      totalCapacityBytes: 214744170496
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 214744170496
      volumes:
  state: Ready
```
