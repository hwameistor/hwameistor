---
sidebar_position: 2
sidebar_label: "Expand Volumes"
---

# Expand Volumes

HwameiStor supports `CSI Volume Expansion`, by which altering the size of `PVC`
can dynamically expand the volume online.

The below example will expand PVC `data-sts-mysql-local-0` from 1GiB to 2GiB.

Check the current size of the `PVC/PV`.

```console
$ kubectl get pvc data-sts-mysql-local-0
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
data-sts-mysql-local-0   Bound    pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   1Gi        RWO            hwameistor-storage-lvm-hdd   85m

$ kubectl get pv pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS                 REASON   AGE
pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   1Gi        RWO            Delete           Bound    default/data-sts-mysql-local-0   hwameistor-storage-lvm-hdd            85m
```

## Verify `StorageClass`

Verify if the `StorageClass` has the parameter `allowVolumeExpansion: true`.

```console
$ kubectl get pvc data-sts-mysql-local-0 -o jsonpath='{.spec.storageClassName}'
hwameistor-storage-lvm-hdd

$ kubectl get sc hwameistor-storage-lvm-hdd -o jsonpath='{.allowVolumeExpansion}'
true
```

## Edit `PVC` size

```console
$ kubectl edit pvc data-sts-mysql-local-0

...
spec:
  resources:
    requests:
      storage: 2Gi
...
```

## Observe the process

The larger the volume, the longer it takes to expand the volume. You may observe the process from `PVC` events.

```console
$ kubectl describe pvc data-sts-mysql-local-0

Events:
  Type     Reason                      Age                From                                Message
  ----     ------                      ----               ----                                -------
  Warning  ExternalExpanding           34s                volume_expand                       Ignoring the PVC: didn't find a plugin capable of expanding the volume; waiting for an external controller to process this PVC.
  Warning  VolumeResizeFailed          33s                external-resizer lvm.hwameistor.io  resize volume "pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8" by resizer "lvm.hwameistor.io" failed: rpc error: code = Unknown desc = volume expansion not completed yet
  Normal   Resizing                    32s (x2 over 33s)  external-resizer lvm.hwameistor.io  External resizer is resizing volume pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8
  Normal   FileSystemResizeRequired    32s                external-resizer lvm.hwameistor.io  Require file system resize of volume on node
  Normal   FileSystemResizeSuccessful  11s                kubelet                             MountVolume.NodeExpandVolume succeeded for volume "pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8" k8s-worker-3
```

## Verify the size of `PVC/PV` after expansion

```console
$ kubectl get pvc data-sts-mysql-local-0
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
data-sts-mysql-local-0   Bound    pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   2Gi        RWO            hwameistor-storage-lvm-hdd   96m

$ kubectl get pv pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS                 REASON   AGE
pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   2Gi        RWO            Delete           Bound    default/data-sts-mysql-local-0   hwameistor-storage-lvm-hdd            96m
```
