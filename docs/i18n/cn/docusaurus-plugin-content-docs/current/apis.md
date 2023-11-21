---
sidebar_position: 11
sidebar_label: "API 管理"
---

# API 管理

## CRD 对象类

HwameiStor 在 Kubernetes 已有的 PV 和 PVC 对象类基础上，HwameiStor 定义了更丰富的对象类，把 PV/PVC 和本地数据盘关联起来。

| 名称                                 | 缩写                         | Kind                              | 功能                        |
|------------------------------------|----------------------------|-----------------------------------|---------------------------|
| clusters                           | hmcluster                  | Cluster                           | HwameiStor 集群             |
| events                             | evt                        | Event                             | HwameiStor 集群的审计日志        |
| localdiskclaims                    | ldc                        | LocalDiskClaim                    | 筛选并分配本地数据盘                |
| localdisknodes                     | ldn                        | LocalDiskNode                     | 裸磁盘类型数据卷的存储节点             |
| localdisks                         | ld                         | LocalDisk                         | 节点上数据盘，自动识别空闲可用的数据盘       |
| localdiskvolumes                   | ldv                        | LocalDiskVolume                   | 裸磁盘类型数据卷                  |
| localstoragenodes                  | lsn                        | LocalStorageNode                  | LVM 类型数据卷的存储节点            |
| localvolumeconverts                | lvconvert                  | LocalVolumeConvert                | 将普通LVM类型数据卷转化为高可用LVM类型数据卷 |
| localvolumeexpands                 | lvexpand                   | LocalVolumeExpand                 | 扩容LVM类型数据卷的容量             |
| localvolumegroups                  | lvg                        | LocalVolumeGroup                  | LVM 类型数据卷组                |
| localvolumemigrates                | lvmigrate                  | LocalVolumeMigrate                | 迁移LVM类型数据卷                |
| localvolumereplicas                | lvr                        | LocalVolumeReplica                | LVM 类型数据卷的副本              |
| localvolumereplicasnapshotrestores | lvrsrestore,lvrsnaprestore | LocalVolumeReplicaSnapshotRestore | 恢复 LVM 类型数据卷副本的快照         |
| localvolumereplicasnapshots        | lvrs                       | LocalVolumeReplicaSnapshot        | LVM 类型数据卷副本的快照            |
| localvolumes                       | lv                         | LocalVolume                       | LVM 类型数据卷                 |
| localvolumesnapshotrestores        | lvsrestore,lvsnaprestore   | LocalVolumeSnapshotRestore        | 恢复 LVM 类型数据卷快照            |
| localvolumesnapshots               | lvs                        | LocalVolumeSnapshot               | LVM 类型数据卷快照               |
| resizepolicies                     |                            | ResizePolicy                      | PVC自动扩容策略                 |
