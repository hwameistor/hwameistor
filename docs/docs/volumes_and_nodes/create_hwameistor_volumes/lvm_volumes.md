---
sidebar_position: 1
sidebar_label: "LVM Volume"
---

# LVM Volume

HwameiStor provides LVM-based data volumes,
which offer read and write performance comparable to that of native disks.
These data volumes also provide advanced features such as data volume expansion, migration, high availability, and more.

The following example will demonstrate how to create a simple non-HA type data volume.

## Steps

### 1. Prepare LVM Storage Node

It is necessary to ensure that the storage node has available capacity. If there is not enough capacity, 
please refer to [Expanding LVM Storage Nodes](../node_expansion/lvm_nodes.md).

Check for available capacity using the following command:

```console
$ kubectl get loalstoragenodes k8s-worker-2 -oyaml | grep freeCapacityBytes
freeCapacityBytes: 10523508736
```

### 2. Prepare StorageClass

Create a StorageClass named `hwameistor-storage-lvm-ssd` using the following command:

```console
$ cat << EOF | kubectl apply -f - 
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:  
  name: hwameistor-storage-lvm-ssd 
parameters:
  convertible: "false"
  csi.storage.k8s.io/fstype: xfs
  poolClass: SSD
  poolType: REGULAR
  replicaNumber: "1"
  striped: "true"
  volumeKind: LVM
provisioner: lvm.hwameistor.io
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
EOF 
```

### 3. Create Volume

Create a PVC named `hwameistor-lvm-volume` using the following command:

```console
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hwameistor-lvm-volume
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: hwameistor-storage-lvm-ssd
EOF
```

### 4. Use Volume

Create an nginx application and use `hwameistor-lvm-volume` volume using the following command:

```console
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
containers:
- name: volume-test
  image: docker.io/library/nginx:latest
  imagePullPolicy: IfNotPresent
  volumeMounts:
  - name: data
    mountPath: /data
  ports:
  - containerPort: 80
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: hwameistor-disk-volume
```
