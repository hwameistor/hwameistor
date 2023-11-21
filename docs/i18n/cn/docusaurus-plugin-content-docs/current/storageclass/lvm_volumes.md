---
sidebar_position: 1
sidebar_label: "创建 LVM 存储池"
---

# LVM 数据卷

HwameiStor 提供了基于 LVM 的数据卷。
这种类型的数据卷提供了接近原生磁盘的读写性能，并且在此之上提供了数据卷扩容、迁移、HA 等等高级特性。

下文将示例创建一个最简单的非 HA 类型数据卷。

1. 准备 LVM 存储节点

   需要保证该存储节点有可用容量，如果没有，可以参考[LVM 存储节点扩容](../nodes_and_disks/lvm_nodes.md)。

   通过以下命令查看是否有可用容量：

   ```shell
   $ kubectl get localstoragenodes k8s-worker-2 -oyaml | grep freeCapacityBytes
   freeCapacityBytes: 10523508736
   ```

2. 准备 StorageClass

   使用以下命令创建一个名称为 `hwameistor-storage-lvm-ssd` 的 StorageClass：

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

3. 创建数据卷 PVC

   使用以下命令创建一个名称为 `hwameistor-lvm-volume` 的 PVC：

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
