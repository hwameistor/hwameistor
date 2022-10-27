---
sidebar_position: 4
sidebar_label: "API"
---

# API

## CRD 对象类

Hwameistor 在 Kubernetes 已有的 PV 和 PVC 对象类基础上，Hwameistor 定义了更丰富的对象类，把 PV/PVC 和本地数据盘关联起来。

|Kind|缩写|功能|
|--|--|--|
|LocalDiskNode|ldn|注册节点|
|LocalDisk|ld|注册节点上数据盘，自动识别空闲可用的数据盘|
|LocalDiskClaim|ldc|筛选并注册本地数据盘|
|LocalStorageNode|lsn|自动创建存储池，也就是 LVM 逻辑卷组|
|LocalVolume|lv|创建 LVM 逻辑卷，分配给 PersistentVolume|
|LocalDiskExpand|lvexpand|存储池扩容|
