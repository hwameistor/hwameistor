---
sidebar_position: 2
sidebar_label: "Use Disk Volume"
---

# Use Disk Volume

HwameiStor provides another type of data volume known as raw disk data volume.
This volume is based on the raw disk present on the node and can be directly mounted for container use.
As a result, this type of data volume offers more efficient data read and write performance,
thereby fully unleashing the performance of the disk.

The following steps demonstrate how to create and use raw disk data volumes:

1. Prepare a raw disk storage node

   Make sure that the storage node has available disks. If not, refer to [disk storage node expansion](../nodes_and_disks/disk_nodes.md).

   Use the following command to check whether there are free disks:

   ```shell
   $ kubectl get localdisknodes
   NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
   k8s-worker-2   1073741824     1073741824      1           Ready    19d
   ```

2. Prepare StorageClass

   Use the following command to create a StorageClass named `hwameistor-storage-disk-ssd`:

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
3. Create a data volume PVC

   Use the following command to create a PVC named `hwameistor-disk-volume`:

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
