---
sidebar_position: 3
sidebar_label: "Disk Storage Node"
---

# Disk Storage Node

Raw disk storage nodes provide applications with raw disk data volumes and 
maintain the mapping between raw disks and raw disk data volumes on the storage node.

## Steps

### 1. Prepare a disk storage node

Add the node to the Kubernetes cluster or select a Kubernetes node.

For example, suppose you have a new node with the following information:

- name: k8s-worker-2
- devPath: /dev/sdb
- diskType: SSD disk

After the new node is already added into the Kubernetes cluster,
make sure the following HwameiStor pods are already running on this node.

```bash
$ kubectl get node
NAME           STATUS   ROLES            AGE     VERSION
k8s-master-1   Ready    master           96d     v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker           96h     v1.24.3-2+63243a96d1c393

$ kubectl -n hwameistor get pod -o wide | grep k8s-worker-2
hwameistor-local-disk-manager-sfsf1     2/2     Running   0     19h   10.6.128.150      k8s-worker-2   <none>  <none>

# check LocalDiskNode resource
$ kubectl get localdisknode k8s-worker-2
NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
k8s-worker-2                                              Ready    21d
```

### 2. Add the storage node into HwameiStor

First, change the `owner` information of the disk sdb to local-disk-manager as below:

```console
$ kubectl edit ld localdisk-2307de2b1c5b5d051058bc1d54b41d5c
apiVersion: hwameistor.io/v1alpha1
kind: LocalDisk
metadata:
  name: localdisk-2307de2b1c5b5d051058bc1d54b41d5c
spec:
  devicePath: /dev/sdb
  nodeName: k8s-worker-2
+ owner: local-disk-manager
...
```

Create the storage pool of the node by adding a LocalStorageClaim CR as below:

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-2
spec:
  nodeName: k8s-worker-2
  owner: local-disk-manager
  description:
    diskType: SSD
EOF
```

### 3. Post check

Finally, check if the node has created the storage pool by checking the LocalDiskNode CR.

```bash
kubectl get localstoragenode k8s-worker-2 -o yaml
```

The output may look like:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskNode
metadata:
  name: k8s-worker-2
spec:
  nodeName: k8s-worker-2
status:
  pools:
    LocalDisk_PoolSSD:
      class: SSD
      disks:
        - capacityBytes: 214744170496
          devPath: /dev/sdb
          state: Available
          type: SSD
      freeCapacityBytes: 214744170496
      freeVolumeCount: 1
      totalCapacityBytes: 214744170496
      totalVolumeCount: 1
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 214744170496
      volumes:
  state: Ready
```
