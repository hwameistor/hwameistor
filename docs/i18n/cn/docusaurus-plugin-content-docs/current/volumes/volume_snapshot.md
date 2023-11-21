---
sidebar_position: 1
sidebar_label: "数据卷快照"
---

# 数据卷快照

在 HwameiStor 中，它允许用户创建数据卷的快照（VolumeSnapshot），且可以基于数据卷快照进行还原、回滚操作。

:::note
目前仅支持对非高可用 LVM 类型数据卷创建快照。

为了避免数据不一致，请先暂停或者停止 I/O 然后再打快照。
:::

请按照以下步骤创建 VolumeSnapshotClass 和 VolumeSnapshot 来使用它。

## 创建新的 VolumeSnapshotClass

默认情况下，HwameiStor 在安装过程中不会自动创建这样的 VolumeSnapshotClass，因此您需要手动创建 VolumeSnapshotClass。

示例 VolumeSnapshotClass 如下：

```yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1
metadata:
  name: hwameistor-storage-lvm-snapshot
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
parameters:
  snapsize: "1073741824"
driver: lvm.hwameistor.io
deletionPolicy: Delete
```

- snapsize：指定创建卷快照的大小。

:::note
如果不指定 snapsize 参数，那么创建的快照大小和源卷的大小一致。
:::

创建 VolumeSnapshotClass 后，您可以使用它来创建 VolumeSnapshot。

## 使用 VolumeSnapshotClass 创建 VolumeSnapshot

示例 VolumeSnapshot 如下：

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: snapshot-local-storage-pvc-lvm
spec:
  volumeSnapshotClassName: hwameistor-storage-lvm-snapshot
  source:
    persistentVolumeClaimName: local-storage-pvc-lvm
```

- persistentVolumeClaimName：指定要创建快照的 PVC。

创建 VolumeSnapshot 后，您可以使用如下命令检查 VolumeSnapshot。

```console
$ kubectl get vs
NAME                             READYTOUSE   SOURCEPVC               SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS                     SNAPSHOTCONTENT                                    CREATIONTIME   AGE
snapshot-local-storage-pvc-lvm   true         local-storage-pvc-lvm                           1Gi           hwameistor-storage-lvm-snapshot   snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   53y            2m57s
```

创建 VolumeSnapshot 后，您可以使用如下命令检查 HwameiStor 本地卷快照。

```console
$ kubectl get lvs
NAME                                               CAPACITY     SOURCEVOLUME                               STATE   MERGING   INVALID   AGE
snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   1073741824   pvc-967baffd-ce10-4739-b996-87c9ed24e635   Ready                       5m31s
```

- CAPACITY：快照的容量大小
- SOURCEVOLUME：快照的源卷名称
- MERGING：快照是否处于合并状态（一般由**回滚操作**触发）
- INVALID：快照是否失效（一般在**快照容量写满**触发）
- AGE：快照真实创建的时间（不同于 CR 创建的时间，这个时间是底层快照数据卷的创建时间）

创建 VolumeSnapshot 后，您可以对 VolumeSnapshot 进行还原、回滚操作。

## 对卷快照进行还原操作

可以创建 pvc，对卷快照进行还原操作。具体如下：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: local-storage-pvc-lvm-restore
spec:
  storageClassName: local-storage-hdd-lvm
  dataSource:
    name: snapshot-local-storage-pvc-lvm
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

## 对卷快照进行回滚操作

:::note
对快照进行回滚必须先**停止源卷的 I/O**，比如先停止应用，等待回滚操作完全结束，
并**确认数据一致性**之后再使用回滚后的数据卷。
:::

可以通过创建资源 LocalVolumeSnapshotRestore，对卷快照进行回滚操作。具体如下：

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeSnapshotRestore
metadata:
  name: restore-test
spec:
  sourceVolumeSnapshot: snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110
  restoreType: "rollback"
```

- sourceVolumeSnapshot：指定要进行回滚操作的卷快照。

对创建的 LocalVolumeSnapshotRestore 进行观察，可以通过状态了解整个回滚的过程。回滚结束后，对应的 LocalVolumeSnapshotRestore 会被删除。

```console
NAME            TARGETVOLUME                               SOURCESNAPSHOT                                     STATE        AGE
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   Submitted    0s
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-81a1f605-c28a-4e60-8c78-a3d504cbf6d9   InProgress   0s
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-81a1f605-c28a-4e60-8c78-a3d504cbf6d9   Completed    2s
```
