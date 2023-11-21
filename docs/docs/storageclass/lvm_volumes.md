---
sidebar_position: 1
sidebar_label: "LVM Volume"
---

# LVM Volume

HwameiStor provides LVM-based data volumes,
which offer read and write performance comparable to that of native disks.
These data volumes also provide advanced features such as data volume expansion, migration, high availability, and more.

The following steps demonstrate how to create a simple non-HA type data volume.

1. Prepare LVM Storage Node

   Ensure that the storage node has sufficient available capacity. If there is not enough capacity, 
   please refer to [Expanding LVM Storage Nodes](../nodes_and_disks/lvm_nodes.md).

   Check for available capacity using the following command:

   ```console
   $ kubectl get localstoragenodes k8s-worker-2 -oyaml | grep freeCapacityBytes
   freeCapacityBytes: 10523508736
   ```

2. Prepare StorageClass

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

3. Create Volume

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
