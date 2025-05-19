---
sidebar_position: 11
sidebar_label: "数据卷Thin特性"
---

# Hwameistor Thin Provision 使用指南

## 1. 概述

Hwameistor 现在支持 Thin Provision（精简配置）功能，基于 LVM 的 thin 特性实现。与传统的 thick（厚置备）方式相比，thin 模式可以更高效地利用存储空间，并支持快速创建快照和克隆。

## 2. 适用场景

**推荐在以下情况下使用 thin 特性：**
- 需要频繁创建快照或克隆卷
- 存储空间有限，需要超分配置（over-provisioning）
- 应用场景对存储性能要求不是极端苛刻
- 单副本场景（当前版本暂不支持 thin 多副本）

**不推荐在以下情况使用 thin 特性：**
- 对性能要求极高的场景（thin 有一定性能开销）
- 需要多副本高可用的场景（当前版本限制）

## 3. 快速开始

### 3.1 创建 ThinPoolClaim

首先需要创建 ThinPoolClaim：

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: ThinPoolClaim
metadata:
  name: example-thinpool
spec:
  nodeName: node1  # 指定节点
  description:
    poolName: LocalStorage_PoolHDD  # 指定Thin Pool将在哪个存储池中创建精简池，可选：LocalStorage_PoolHDD, LocalStorage_PoolSSD, LocalStorage_PoolNVMe
    capacity: 100  # ThinPool 容量，单位GiB
    overProvisionRatio: "1.0"  # 超分比例, 默认且最小为1.0。例：如果OverProvisionRatio为"3.0"，capacity为100GiB，则该池可以超额配置300GiB
    poolMetadataSize: 1  # 元数据池大小，单位GiB。默认为1G。可以应对大多数场景
```

### 3.2 创建 Thin Storageclass

```yaml
allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hwameistor-storage-lvm-thin-hdd
parameters:
  convertible: "false"
  csi.storage.k8s.io/fstype: ext4
  poolClass: HDD
  poolType: REGULAR
  # 当前仅支持"1"
  replicaNumber: "1"
  striped: "true"
  # 用于指定该sc创建thin pvc
  thin: "true"
  volumeKind: LVM
provisioner: lvm.hwameistor.io
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

## 4. 使用示例

### 4.1 PVC使用

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
  storageClassName: hwameistor-storage-lvm-thin-hdd
---
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: test-pvc
  terminationGracePeriodSeconds: 10
```

### 4.2 快照

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-snapshot
spec:
  volumeSnapshotClassName: hwameistor-storage-lvm-snapshot
  source:
    persistentVolumeClaimName: test-pvc
```

### 4.3 以快照创建pvc

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc2
spec:
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
  storageClassName: hwameistor-storage-lvm-thin-hdd
  dataSource:
    name: my-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
---
apiVersion: v1
kind: Pod
metadata:
  name: test-pod2
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: test-pvc2
```

### 4.4 克隆操作

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloned-pvc
spec:
  storageClassName: hwameistor-storage-lvm-thin-hdd
  dataSource:
    name: test-pvc
    kind: PersistentVolumeClaim
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: cloned-pod
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: cloned-pvc
  terminationGracePeriodSeconds: 10
```

## 5. 监控与管理

### 5.1 查看 Thin Pool 状态

```bash
kubectl get localstoragenodes node-name -o yaml
```

关注以下字段：
- `status.pools.<pool-name>.thinPool`: 包含 thin pool 的详细信息
- `status.pools.<pool-name>.thinPoolExtendRecords`: 记录了 thin pool 的扩展记录

### 5.2 扩展 Thin Pool

当 thin pool 使用率接近上限时，可以再次创建 ThinPoolClaim 来扩展，其中 `spec.description.capacity`、`spec.description.poolMetadataSize` 字段均允许改为更大的值，而`spec.description.overProvisionRatio`允许根据需求自由调整

## 6. 注意事项

1. **超分风险**：虽然 thin 支持超分配置，但实际使用量超过物理容量会导致严重问题，因此**请务必密切监控 thin pool 的使用情况，关注`status.pools.<pool-name>.thinPool`中dataPercent、metadataPercent，避免写满**
2. **性能影响**：thin 卷有一定性能开销，对性能敏感的应用需谨慎评估
3. **版本兼容**：thin 和 thick 卷不能相互转换
4. **多副本限制**：当前版本仅支持单副本 thin 卷

## 7. 故障处理

如果 thin pool 接近写满：
1. 立即停止创建新的 thin 卷
2. 删除不必要的快照和克隆卷
3. 扩展 thin pool 容量
4. 如果已经写满，请参考[LVM使用文档](https://man7.org/linux/man-pages/man7/lvmthin.7.html)修复
