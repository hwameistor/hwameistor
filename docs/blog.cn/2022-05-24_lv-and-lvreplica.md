---
slug: 3
title: LV 和 LVReplica
authors: Niulechuan
tags: [hello, LVs]
---

在 Kubernetes 中，当用户创建一个 PVC，并指定使用 HwameiStor 作为底层存储时，HwameiStor 会创建两类 CR，即本文的主角`LocalVolume` 和 `LocalVolumeReplica`。HwameiStor 为什么为一个 PV 创建这两类资源呢？本文将为您揭开谜团。

![LV 副本](img/lv_replicas_cn.png)

## LocalVolume

`LocalVolume` 是 HwameiStor 定义的 CRD，代表 HwameiStor 为用户提供的数据卷。`LocalVolume` 和 Kubernetes 的 `PersistentVolume` 是一一对应的，含义也是类似的，均代表一个数据卷。不同之处在于，`LocalVolume` 记录 HwameiStor 相关的信息，而 `PersistentVolume` 记录 Kubernetes 平台本身的信息，并关联到 `LocalVolume`。

可以通过以下命令查看系统中 `LocalVolume` 的详细信息：

```
#  check status of local volume and volume replica
$ kubectl get lv # or localvolume
NAME                                       POOL                   KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM    1          1073741824   k8s-node1       Ready      -1                     22m
```

既然 HwameiStor 可以通过 `LocalVolume` 表示一个数据卷，为什么还需要 `LocalVolumeReplica` 呢？

## LocalVolumeReplica

`LocalVolumeReplica` 也是 HwameiStor 定义的 CRD。但是与 `LocalVolume` 不同，`LocalVolumeReplica` 代表数据卷的副本。

在 HwameiStor 中，`LocalVolume` 会指定某个属于它的 `LocalVolumeReplica` 作为当前激活的副本。可以看出`LocalVolume` 可以拥有多个 `LocalVolumeReplica`，即一个数据卷可以有多个副本。目前 HwameiStor 会在众多副本中激活其中一个，被应用程序挂载，其他副本作为热备副本。

可以通过以下命令查看系统中 `LocalVolumeReplica` 的详细信息：

```
$ kubectl get lvr # or localvolumereplica
NAME                                              KIND   CAPACITY     NODE        STATE   SYNCED   DEVICE                                                               AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2-v5scm9   LVM    1073741824   k8s-node1   Ready   true     /dev/LocalStorage_PoolHDD/pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   80s
```

有了卷副本（LocalVolumeReplica）的概念后，HwameiStor 作为一款本地存储系统，具备了一些很有竞争力的特性，例如数据卷的HA，迁移，热备，Kubernetes 应用快速恢复等等。

## 总结

其实 `LocalVolume` 和 `LocalVolumeReplica` 在很多存储系统中都有引入，是个很通用的概念。只是通过这一概念，实现了各具特色的产品，在解决某个技术难点的时候也可能采取不同的解决方案，因此而适合于不同的生产场景。

随着 HwameiStor 的迭代和演进，我们将会提供更多的能力，从而适配越来越多的使用场景。无论您是用户还是开发者，欢迎您加入 HwameiStor 的大家庭！
