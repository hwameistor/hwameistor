---
sidebar_position: 1
sidebar_label: "Node Expansion"
---

# Node Expansion

It's one of common requirements for a storage system to expand the capacity by adding a new storage node. HwameiStor provides this feature by the following simple procedure.

## Steps

### 1. Prepare a storage node (e.g. k8s-worker-4, /dev/sdb, SSD disk)

Add a node into the Kubernetes cluster, or select a Kubernetes node. And ensure the node has all the required items which are described in [Prerequisites](../install/prereq.md).

After the new node is already in the Kubernetes cluster, make sure the following HwameiStor pods are already running on this node.

```console
$ kubectl -n hwameistor get pod -o wide | grep k8s-worker-4
[root@demo-dev-master-01 ~]# k -n hwameistor get pod -o wide | grep demo-dev-worker-02
hwameistor-local-disk-manager-c86g5     2/2     Running   0     19h   10.6.182.105      k8s-worker-4   <none>  <none>
hwameistor-local-storage-s4zbw          2/2     Running   0     19h   192.168.140.82    k8s-worker-4   <none>  <none>

# check for the LSN which is for the metadata of the node's storage
$ kubectl get localstoragenode k8s-worker-4
NAME                 IP           ZONE      REGION    STATUS   AGE
k8s-worker-4   10.6.182.103       default   default   Ready    8d
```

### 2. Add the storage node into HwameiStor

Add the storage label into the node by adding a LocalStorageClaim CR as below:

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

Finally, check if the node has the storage pool setup by following:

```console
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
