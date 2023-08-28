---
sidebar_position: 5
sidebar_label: "Volume Snapshot"
---

# Volume Snapshot

In HwameiStor, it allows users to create snapshots of non-highly available volumes. And restore and rollback operations can be performed based on volume snapshots

Please follow the steps below to create a VolumeSnapshotClass and a VolumeSnapshot to use it.

## Create a new VolumeSnapshotClass

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
```

- snapsize：It specifies the size of VolumeSnapshot


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
```yaml
$ kubectl get vs
NAME                             READYTOUSE   SOURCEPVC               SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS                     SNAPSHOTCONTENT                                    CREATIONTIME   AGE
snapshot-local-storage-pvc-lvm   true         local-storage-pvc-lvm                           1Gi           hwameistor-storage-lvm-snapshot   snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   53y            2m57s

```

After creating a VolumeSnapshot, you can check the Hwameistor LocalvolumeSnapshot using the following command.

```yaml
$ kubectl get lvs
NAME                                               CAPACITY     SOURCEVOLUME                               STATE   MERGING   INVALID   AGE
snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110   1073741824   pvc-967baffd-ce10-4739-b996-87c9ed24e635   Ready                       5m31s

```

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

VolumeSnapshot can be rolled back by creating the resource LocalVolumeSnapshotRecover, as follows:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeSnapshotRecover
metadata:
  name: recover-test
spec:
  sourceVolumeSnapshot: snapcontent-0fc17697-68ea-49ce-8e4c-7a791e315110
  recoverType: "rollback"
  targetPoolName: LocalStorage_PoolHDD
  targetVolume: pvc-967baffd-ce10-4739-b996-87c9ed24e635
```
- sourceVolumeSnapshot：It specifies the VolumeSnapshot to be rollback.
- targetPoolName: It specifies the storage pool where the rollback target volume is located.
- targetVolume:  It specifies the data volume of the target of the rollback.