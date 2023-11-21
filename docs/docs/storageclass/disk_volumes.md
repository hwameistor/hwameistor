---
sidebar_position: 2
sidebar_label: "Disk Volume"
---

# Disk Volume

HwameiStor provides another type of data volume known as raw disk data volume.
This volume is based on the raw disk present on the node and can be directly mounted for container use.
As a result, this type of data volume offers more efficient data read and write performance,
thereby fully unleashing the performance of the disk.

The following steps demonstrate how to create a disk voluem.

1. Prepare Disk Storage Node

   Ensure that the storage node has sufficient available capacity. If there is not enough capacity,
   please refer to [Expanding LVM Storage Nodes](../nodes_and_disks/disk_nodes.md).

   Check for available capacity using the following command:

   ```shell
   $ kubectl get localdisknodes
   NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
   k8s-worker-2   1073741824     1073741824      1           Ready    19d
   ```

2. Prepare StorageClass

   Create a StorageClass named `hwameistor-storage-disk-ssd` using the following command:

   ```console
   $ cat << EOF | kubectl apply -f - 
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:  
     name: hwameistor-storage-disk-ssd
   parameters:
     diskType: SSD
   provisioner: disk.hwameistor.io
   allowVolumeExpansion: false
   reclaimPolicy: Delete
   volumeBindingMode: WaitForFirstConsumer
   EOF 
   ```

3. Create Volume

   Create a PVC named `hwameistor-disk-volume` using the following command:

   ```console
   $ cat << EOF | kubectl apply -f -
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: hwameistor-disk-volume
   spec:
     accessModes:
     - ReadWriteOnce
     resources:
       requests:
         storage: 1Gi
     storageClassName: hwameistor-storage-disk-ssd
   EOF
   ```
