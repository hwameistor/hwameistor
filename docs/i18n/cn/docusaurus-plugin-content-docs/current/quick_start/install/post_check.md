---
sidebar_position: 3
sidebar_label: "安装后检查"
---

# 安装后检查

下文展示了一个 kubernetes 集群的示例。

运行以下命令查看各节点状态：

```bash
kubectl get no
```

发现集群由 4 个节点组成：

```console
NAME           STATUS   ROLES   AGE   VERSION
k8s-master-1   Ready    master  82d   v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker  36d   v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker  59d   v1.24.3-2+63243a96d1c393
k8s-worker-3   Ready    worker  36d   v1.24.3-2+63243a96d1c393
```

## 步骤 1：检查 Pod

运行以下命令检查 hwameistor 相关 Pod。

```bash
kubectl -n hwameistor get pod
```

屏幕显示正在运行的 Pod：

```console
NAME                                                       READY   STATUS                  RESTARTS   AGE
hwameistor-local-disk-csi-controller-665bb7f47d-6227f      2/2     Running                 0          30s
hwameistor-local-disk-manager-5ph2d                        2/2     Running                 0          30s
hwameistor-local-disk-manager-jhj59                        2/2     Running                 0          30s
hwameistor-local-disk-manager-k9cvj                        2/2     Running                 0          30s
hwameistor-local-disk-manager-kxwww                        2/2     Running                 0          30s
hwameistor-local-storage-csi-controller-667d949fbb-k488w   3/3     Running                 0          30s
hwameistor-local-storage-csqqv                             2/2     Running                 0          30s
hwameistor-local-storage-gcrzm                             2/2     Running                 0          30s
hwameistor-local-storage-v8g7t                             2/2     Running                 0          30s
hwameistor-local-storage-zkwmn                             2/2     Running                 0          30s
hwameistor-scheduler-58dfcf79f5-lswkt                      1/1     Running                 0          30s
hwameistor-webhook-986479678-278cr                         1/1     Running                 0          30s
```

:::info
`local-disk-manager` 和 `local-storage` 是 `DaemonSet`。在每个 Kubernetes 节点上都应该有一个 DaemonSet Pod。
:::

## 步骤 2：检查 StorageClass

创建 `storageClass` 后，运行以下命令：

```bash
kubectl get storageclass hwameistor-storage-disk-hdd
```

屏幕显示当前已创建的 StorageClass：

```console
NAME                          PROVISIONER          RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-disk-hdd   disk.hwameistor.io   Delete          WaitForFirstConsumer   true                   4m29s
```

## 步骤 3：检查 API

运行以下命令通过 HwameiStor CRD 创建 API。

```bash
kubectl api-resources --api-group hwameistor.io
```

创建的 API 如下：

```console
NAME                       SHORTNAMES   APIVERSION               NAMESPACED   KIND
localdiskclaims            ldc          hwameistor.io/v1alpha1   false        LocalDiskClaim
localdisknodes             ldn          hwameistor.io/v1alpha1   false        LocalDiskNode
localdisks                 ld           hwameistor.io/v1alpha1   false        LocalDisk
localdiskvolumes           ldv          hwameistor.io/v1alpha1   false        LocalDiskVolume
localstoragenodes          lsn          hwameistor.io/v1alpha1   false        LocalStorageNode
localvolumeconverts        lvconvert    hwameistor.io/v1alpha1   true         LocalVolumeConvert
localvolumeexpands         lvexpand     hwameistor.io/v1alpha1   false        LocalVolumeExpand
localvolumegroupconverts   lvgconvert   hwameistor.io/v1alpha1   true         LocalVolumeGroupConvert
localvolumegroupmigrates   lvgmigrate   hwameistor.io/v1alpha1   true         LocalVolumeGroupMigrate
localvolumegroups          lvg          hwameistor.io/v1alpha1   true         LocalVolumeGroup
localvolumemigrates        lvmigrate    hwameistor.io/v1alpha1   true         LocalVolumeMigrate
localvolumereplicas        lvr          hwameistor.io/v1alpha1   false        LocalVolumeReplica
localvolumes               lv           hwameistor.io/v1alpha1   false        LocalVolume
```

有关 CRD 的详细信息，另请参见 [CRD](../../architecture/apis.md) 一节。

## 步骤 4：检查 localDiskNode 和 localDisks

HwameiStor 自动扫描每个节点并将每个磁盘注册为 CRD `localDisk(ld)`。未使用的磁盘显示为 `PHASE: Unclaimed`。

运行以下命令查看带有 localDisk 的节点：

```bash
kubectl get localdisknodes
```

屏幕显示当前 4 个节点上的磁盘状况：

```console
NAME           TOTALDISK   FREEDISK
k8s-master-1   5           3
k8s-worker-1   5           2
k8s-worker-2   5           2
k8s-worker-3   5           2
```

运行以下命令查看 localdisk：

```bash
kubectl get localdisks
```

屏幕显示 sda、sdb、sdc、sdd、sde 等磁盘的信息：

```console
NAME               NODEMATCH      CLAIM   PHASE
k8s-master-1-sda   k8s-master-1           Inuse
k8s-worker-1-sda   k8s-worker-1           Inuse
k8s-worker-1-sdb   k8s-worker-1           Unclaimed
k8s-worker-1-sdc   k8s-worker-1           Unclaimed
k8s-worker-2-sda   k8s-worker-2           Inuse
k8s-worker-2-sdb   k8s-worker-2           Unclaimed
k8s-worker-2-sdc   k8s-worker-2           Unclaimed
k8s-worker-3-sda   k8s-worker-3           Inuse
k8s-worker-3-sdb   k8s-worker-3           Unclaimed
k8s-worker-3-sdc   k8s-worker-3           Unclaimed
```
