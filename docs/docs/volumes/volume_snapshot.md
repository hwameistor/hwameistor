---
sidebar_position: 1
sidebar_label: "Volume Snapshot"
---

# Volume Snapshot

In HwameiStor, it allows users to create snapshots of data volumes and perform restore and rollback operations based on data volume snapshots.

:::note
Currently, only snapshots are supported for non highly available LVM type data volumes.

To avoid data inconsistency, please pause or stop I/O before taking a snapshot.
:::

Please follow the steps below to create a VolumeSnapshotClass and a VolumeSnapshot to use it.

## Create VolumeSnapshotClass

By default, HwameiStor does not automatically create a VolumeSnapshotClass during installation, so you need to create a VolumeSnapshotClass manually.

A sample VolumeSnapshotClass is as follows:

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

- snapsize：It specifies the size of VolumeSnapshot

:::note
If the snapsize parameter is not specified, the size of the created snapshot is consistent with the size of the source volume.
:::

After you create a VolumeSnapshotClass, you can use it to create VolumeSnapshot.

## Create a VolumeSnapshot using the VolumeSnapshotClass

A sample VolumeSnapshot is as follows:

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

- persistentVolumeClaimName：It specifies the PVC to create the VolumeSnapshot

After creating a VolumeSnapshot, you can check the VolumeSnapshot using the following command.

```console
$ kubectl get vs
NAME                             READYTOUSE   SOURCEPVC               SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS                     SNAPSHOTCONTENT                                    CREATIONTIME   AGE
snapshot-local-storage-pvc-lvm   true         local-storage-pvc-lvm                           1Gi           hwameistor-storage-lvm-snapshot   snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   53y            2m57s
```

After creating a VolumeSnapshot, you can check the Hwameistor LocalvolumeSnapshot using the following command.

```console
$ kubectl get lvs
NAME                                               CAPACITY     SOURCEVOLUME                               STATE   MERGING   INVALID   AGE
snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   1073741824   pvc-967baffd-ce10-4739-b996-87c9ed24e635   Ready                       5m31s
```

- CAPACITY: The capacity size of the snapshot
- SourceVOLUME: The source volume name of the snapshot
- MERGING: Whether the snapshot is in a merged state (usually triggered by *rollback operation*)
- INVALID: Whether the snapshot is invalidated (usually triggered when *the snapshot capacity is full*)
- AGE: The actual creation time of the snapshot (different from the CR creation time, this time is the creation time of the underlying snapshot data volume)

After creating a VolumeSnapshot, you can restore and rollback the VolumeSnapshot.

## Restore VolumeSnapshot

You can create pvc to restore VolumeSnapshot, as follows:

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

## Rollback VolumeSnapshot

:::note
To roll back a snapshot, you must first stop the I/O of the source volume, such as stopping the application and waiting for the rollback operation to complete,
*confirm data consistency* before using the rolled back data volume.
:::

VolumeSnapshot can be rolled back by creating the resource LocalVolumeSnapshotRestore, as follows:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeSnapshotRestore
metadata:
  name: restore-test
spec:
  sourceVolumeSnapshot: snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110
  restoreType: "rollback"
```

- sourceVolumeSnapshot：It specifies the VolumeSnapshot to be rollback.

Observing the created LocalVolumeSnapshotRestore, you can understand the entire rollback process through the state. After the rollback is complete, the corresponding LocalVolumeSnapshotRestore will be deleted.

```console
NAME            TARGETVOLUME                               SOURCESNAPSHOT                                     STATE        AGE
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   Submitted    0s
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-81a1f605-c28a-4e60-8c78-a3d504cbf6d9   InProgress   0s
restore-test2   pvc-967baffd-ce10-4739-b996-87c9ed24e635   snapcontent-81a1f605-c28a-4e60-8c78-a3d504cbf6d9   Completed    2s
```
