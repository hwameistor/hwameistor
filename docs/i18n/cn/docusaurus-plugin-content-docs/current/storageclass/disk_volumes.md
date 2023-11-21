---
sidebar_position: 2
sidebar_label: "创建裸磁盘存储池"
---

# 裸磁盘数据卷

HwameiStor 提供的另一种类型数据卷是裸磁盘数据卷。
这种数据卷是基于节点上面的裸磁盘并将其直接挂载给容器使用。
因此这种类型的数据卷提供了更高效的数据读写性能，将磁盘的性能
完全释放。

以下步骤演示了如何创建裸磁盘存储池：

1. 准备裸磁盘存储节点

   需要保证该存储节点有可用磁盘，如果没有，可以参考[裸磁盘存储节点扩容](../nodes_and_disks/disk_nodes.md)。

   通过以下命令查看是否有空闲磁盘：

   ```shell
   $ kubectl get localdisknodes
   NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
   k8s-worker-2   1073741824     1073741824      1           Ready    19d
   ```

2. 准备 StorageClass

   使用以下命令创建一个名称为 `hwameistor-storage-disk-ssd` 的 StorageClass：

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

3. 创建数据卷 PVC

   使用以下命令创建一个名称为 `hwameistor-disk-volume` 的 PVC：

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
