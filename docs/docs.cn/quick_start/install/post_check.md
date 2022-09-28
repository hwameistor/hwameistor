---
sidebar_position: 3
sidebar_label: "安装后检查"
---

# 安装后检查

下文的示例来自一个 4 节点的 Kubernetes 集群

```console
$ kubectl get no
NAME           STATUS   ROLES   AGE   VERSION
k8s-master-1   Ready    master  82d   v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker  36d   v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker  59d   v1.24.3-2+63243a96d1c393
k8s-worker-3   Ready    worker  36d   v1.24.3-2+63243a96d1c393
```

## 检查 Pod

运行以下命令检查 hwameistor 相关 Pod。

```console
$ kubectl -n hwameistor get pod
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

## 检查 `StorageClass`

创建 `StorageClass` 后，运行以下命令：

```console
$ kubectl get storageclass hwameistor-storage-disk-hdd

NAME                          PROVISIONER          RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-disk-hdd   disk.hwameistor.io   Delete          WaitForFirstConsumer   true                   4m29s
```

## 检查 API

运行以下命令通过 HwameiStor CRD 创建 API。

```console
$ kubectl api-resources --api-group hwameistor.io

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

## 检查 `LocalDiskNode` 和 `LocalDisks`

HwameiStor 自动扫描每个节点并将每块数据盘注册为 CRD `LocalDisk(ld)`。未使用的数据盘显示为 `PHASE: Unclaimed`。

运行以下命令查看带有 `LocalDisk` 的节点：

```console
$ kubectl get localdisknodes
NAME           TOTALDISK   FREEDISK
k8s-master-1   5           3
k8s-worker-1   5           2
k8s-worker-2   5           2
k8s-worker-3   5           2

$ kubectl get localdisks

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

## [可选] 检查 DRBD 的安装

`drbd-adapter` Pod 需要在每个 Worker 节点上运行：

```console
$ kubectl -n hwameistor get po -l k8s-app=drbd-adapter -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP            NODE        
drbd-adapter-4rndg   1/1     Running   0          9h    10.6.254.22   k8s-worker-2   
drbd-adapter-bpprj   1/1     Running   0          9h    10.6.254.21   k8s-worker-1
drbd-adapter-n52w4   1/1     Running   0          9h    10.6.254.24   k8s-worker-4
drbd-adapter-rs9zk   1/1     Running   0          9h    10.6.254.25   k8s-worker-5
drbd-adapter-zc882   1/1     Running   0          9h    10.6.254.23   k8s-worker-3
```

在每个 Worker 节点上，DRBD 的内核模块必须被加载，例如在节点 `k8s-worker-1` 上：

```console
[root@k8s-worker-1 ~]$ lsmod | grep ^drbd
drbd_transport_tcp     22227  0
drbd                  606840  1 drbd_transport_tcp

# cat /proc/drbd
version: 9.0.32-1 (api:2/proto:86-121)
GIT-hash: 7d2933d5a3764fcc5e0bf54b71fd9cfb0363be1a build by @4904565a901d, 2022-09-07 08:53:17
Transports (api:17): tcp (9.0.32-1)
```
