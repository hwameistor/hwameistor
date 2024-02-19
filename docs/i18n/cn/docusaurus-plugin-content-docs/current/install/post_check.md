---
sidebar_position: 3
sidebar_label: "查看系统状态"
---

# 查看安装系统的状态

此处，以三个节点的 Kubernetes 集群为例，安装 HwameiStor 存储系统。

```console
$ kubectl get node
NAME           STATUS   ROLES   AGE   VERSION
10-6-234-40   Ready    control-plane,master   140d   v1.21.11
10-6-234-41   Ready    <none>                 140d   v1.21.11
10-6-234-42   Ready    <none>                 140d   v1.21.11
```

## 查看 HwameiStor 系统组件

以下 Pod 必须在系统中正常运行。

```console
$ kubectl -n hwameistor get pod
NAME                                                       READY   STATUS    RESTARTS      AGE
drbd-adapter-k8s-master-rhel7-gtk7t                        0/2     Completed 0             23m
drbd-adapter-k8s-node1-rhel7-gxfw5                         0/2     Completed 0             23m
drbd-adapter-k8s-node2-rhel7-lv768                         0/2     Completed 0             23m
hwameistor-admission-controller-dc766f976-mtlvw            1/1     Running   0             23m
hwameistor-apiserver-86d6c9b7c8-v67gg                      1/1     Running   0             23m
hwameistor-auditor-54f46fcbc6-jb4f4                        1/1     Running   0             23m
hwameistor-exporter-6498478c57-kr8r4                       1/1     Running   0             23m
hwameistor-failover-assistant-cdc6bd665-56wbw              1/1     Running   0             23m
hwameistor-local-disk-csi-controller-6587984795-fztcd      2/2     Running   0             23m
hwameistor-local-disk-manager-7gg9x                        2/2     Running   0             23m
hwameistor-local-disk-manager-kqkng                        2/2     Running   0             23m
hwameistor-local-disk-manager-s66kn                        2/2     Running   0             23m
hwameistor-local-storage-csi-controller-5cdff98f55-jj45w   6/6     Running   0             23m
hwameistor-local-storage-mfqks                             2/2     Running   0             23m
hwameistor-local-storage-pnfpx                             2/2     Running   0             23m
hwameistor-local-storage-whg9t                             2/2     Running   0             23m
hwameistor-pvc-autoresizer-86dc79d57-s2l68                 1/1     Running   0             23m
hwameistor-scheduler-6db69957f-r58j6                       1/1     Running   0             23m
hwameistor-ui-744cd78d84-vktjq                             1/1     Running   0             23m
hwameistor-volume-evictor-5db99cf979-4674n                 1/1     Running   0             23m
```

:::info
`local-disk-manager` 和 `local-storage` 组件是以 `DaemonSets` 方式进行部署的，必须在每个节点上运行。
:::

## 查看 HwameiStor CRD（即 API）

以下 HwameiStor CRD 必须安装在系统上。

```console
$ kubectl api-resources --api-group hwameistor.io
NAME                                 SHORTNAMES                   APIVERSION               NAMESPACED   KIND
clusters                             hmcluster                    hwameistor.io/v1alpha1   false        Cluster
events                               evt                          hwameistor.io/v1alpha1   false        Event
localdiskclaims                      ldc                          hwameistor.io/v1alpha1   false        LocalDiskClaim
localdisknodes                       ldn                          hwameistor.io/v1alpha1   false        LocalDiskNode
localdisks                           ld                           hwameistor.io/v1alpha1   false        LocalDisk
localdiskvolumes                     ldv                          hwameistor.io/v1alpha1   false        LocalDiskVolume
localstoragenodes                    lsn                          hwameistor.io/v1alpha1   false        LocalStorageNode
localvolumeconverts                  lvconvert                    hwameistor.io/v1alpha1   false        LocalVolumeConvert
localvolumeexpands                   lvexpand                     hwameistor.io/v1alpha1   false        LocalVolumeExpand
localvolumegroups                    lvg                          hwameistor.io/v1alpha1   false        LocalVolumeGroup
localvolumemigrates                  lvmigrate                    hwameistor.io/v1alpha1   false        LocalVolumeMigrate
localvolumereplicas                  lvr                          hwameistor.io/v1alpha1   false        LocalVolumeReplica
localvolumereplicasnapshotrestores   lvrsrestore,lvrsnaprestore   hwameistor.io/v1alpha1   false        LocalVolumeReplicaSnapshotRestore
localvolumereplicasnapshots          lvrs                         hwameistor.io/v1alpha1   false        LocalVolumeReplicaSnapshot
localvolumes                         lv                           hwameistor.io/v1alpha1   false        LocalVolume
localvolumesnapshotrestores          lvsrestore,lvsnaprestore     hwameistor.io/v1alpha1   false        LocalVolumeSnapshotRestore
localvolumesnapshots                 lvs                          hwameistor.io/v1alpha1   false        LocalVolumeSnapshot
resizepolicies                                                    hwameistor.io/v1alpha1   false        ResizePolicy
```

想了解具体的 CRD 信息，请查阅 [CRD](../apis.md)。

## 查看 `LocalDiskNode` 和 `localDisks`

HwameiStor 自动扫描每个节点上的磁盘，并为每一块磁盘生成一个 CRD 资源 `LocalDisk (LD)`。
没有被使用的磁盘，其状态被标记为 `PHASE: Available`。

```console
$ kubectl get localdisknodes
NAME         FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
k8s-master                                              Ready    28h
k8s-node1                                               Ready    28h
k8s-node2                                               Ready    28h

$ kubectl get localdisks
NAME                                         NODEMATCH    DEVICEPATH   PHASE       AGE
localdisk-2307de2b1c5b5d051058bc1d54b41d5c   k8s-node1    /dev/sdb     Available   28h
localdisk-311191645ea00c62277fe709badc244e   k8s-node2    /dev/sdb     Available   28h
localdisk-37a20db051af3a53a1c4e27f7616369a   k8s-master   /dev/sdb     Available   28h
localdisk-b57b108ad2ccc47f4b4fab6f0b9eaeb5   k8s-node2    /dev/sda     Bound       28h
localdisk-b682686c65667763bda58e391fbb5d20   k8s-master   /dev/sda     Bound       28h
localdisk-da121e8f0dabac9ee1bcb6ed69840d7b   k8s-node1    /dev/sda     Bound       28h
```

## 查看 `LocalStorageNodes` 及存储池

HwameiStor 为每个存储节点创建一个 CRD 资源 `LocalStorageNode (LSN)`。
每个 LSN 将会记录该存储节点的状态，及节点上的所有存储资源，包括存储池、数据卷、及相关配置信息。

```console
$ kubectl get lsn
NAME          IP            STATUS   AGE
10-6-234-40   10.6.234.40   Ready    3m52s
10-6-234-41   10.6.234.41   Ready    3m54s
10-6-234-42   10.6.234.42   Ready    3m55s

$ kubectl get lsn 10-6-234-41 -o yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  creationTimestamp: "2023-04-11T06:46:52Z"
  generation: 1
  name: 10-6-234-41
  resourceVersion: "13575433"
  uid: 4986f7b8-6fe1-43f1-bdca-e68b6fa53f92
spec:
  hostname: 10-6-234-41
  storageIP: 10.6.234.41
  topogoly:
    region: default
    zone: default
status:
  pools:
    LocalStorage_PoolHDD:
      class: HDD
      disks:
      - capacityBytes: 10733223936
        devPath: /dev/sdb
        state: InUse
        type: HDD
      - capacityBytes: 1069547520
        devPath: /dev/sdc
        state: InUse
        type: HDD
      - capacityBytes: 1069547520
        devPath: /dev/sdd
        state: InUse
        type: HDD
      - capacityBytes: 1069547520
        devPath: /dev/sde
        state: InUse
        type: HDD
      - capacityBytes: 1069547520
        devPath: /dev/sdf
        state: InUse
        type: HDD
      - capacityBytes: 1069547520
        devPath: /dev/sdg
        state: InUse
        type: HDD
      freeCapacityBytes: 16080961536
      freeVolumeCount: 1000
      name: LocalStorage_PoolHDD
      totalCapacityBytes: 16080961536
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 16080961536
  state: Ready
```

## 查看 `StorageClass`

HwameiStor Operator 在完成 HwameiStor 系统组件安装和系统初始化之后，会根据系统配置
（例如：是否开启 HA 模块、磁盘类型等）自动创建相应的 `StorageClass` 用于创建数据卷。

```console
$ kubectl get sc
NAME                                     PROVISIONER         RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-lvm-hdd               lvm.hwameistor.io   Delete          WaitForFirstConsumer   false                  23h
hwameistor-storage-lvm-hdd-convertible   lvm.hwameistor.io   Delete          WaitForFirstConsumer   false                  23h
hwameistor-storage-lvm-hdd-ha            lvm.hwameistor.io   Delete          WaitForFirstConsumer   false                  23h
```
