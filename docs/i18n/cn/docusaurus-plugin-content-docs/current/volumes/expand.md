---
sidebar_position: 2
sidebar_label:  "卷的扩容"
---

# 卷的扩容

HwameiStor 支持 `CSI 卷扩容` 。这个功能实现了通过修改 `PVC` 的大小在线扩容卷。

下面的例子里，我们把 PVC `data-sts-mysql-local-0` 从 1GiB 扩容到 2GiB。

当前 `PVC/PV` 大小：

```console
$ kubectl get pvc data-sts-mysql-local-0
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
data-sts-mysql-local-0   Bound    pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   1Gi        RWO            hwameistor-storage-lvm-hdd   85m

$ kubectl get pv pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS                 REASON   AGE
pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   1Gi        RWO            Delete           Bound    default/data-sts-mysql-local-0   hwameistor-storage-lvm-hdd            85m
```

## 查看 `StorageClass` 是否使用了参数 `allowVolumeExpansion: true`

```console
$ kubectl get pvc data-sts-mysql-local-0 -o jsonpath='{.spec.storageClassName}'
hwameistor-storage-lvm-hdd

$ kubectl get sc hwameistor-storage-lvm-hdd -o jsonpath='{.allowVolumeExpansion}'
true
```

## 修改 `PVC` 的大小

```console
$ kubectl edit pvc data-sts-mysql-local-0

...
spec:
  resources:
    requests:
      storage: 2Gi
...
```

## 观察扩容过程

增加的容量越多，扩容所需时间越长。可以在 `PVC` 的事件日志中观察整个扩容的过程.

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

## 观察扩容完成后的 `PVC/PV`

```console
$ kubectl get pvc data-sts-mysql-local-0
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
data-sts-mysql-local-0   Bound    pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   2Gi        RWO            hwameistor-storage-lvm-hdd   96m

$ kubectl get pv pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS                 REASON   AGE
pvc-b9fc8651-97b8-414c-8bcf-c8d2708c4ee8   2Gi        RWO            Delete           Bound    default/data-sts-mysql-local-0   hwameistor-storage-lvm-hdd            96m
```
