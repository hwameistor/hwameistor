---
sidebar_position: 8
sidebar_label: "常见问题"
---

# 常见问题

## 问题 1：HwameiStor 本地存储调度器 scheduler 在 kubernetes 平台中是如何工作的？ 

HwameiStor 的调度器是以 Pod 的形式部署在 HwameiStor 的命名空间。

![img](img/clip_image002.png)

当应用（Deployment 或 StatefulSet ）被创建后，应用的 Pod 会被自动部署到已配置好具备 HwameiStor 本地存储能力的 Worker 节点上。

## 问题 2: HwameiStor 如何应对应用多副本工作负载的调度？与传统通用型共享存储有什么不同？

HwameiStor 建议使用有状态的 StatefulSet 用于多副本的工作负载。

有状态应用 StatefulSet 会将复制的副本部署到同一 Worker 节点，但会为每一个 Pod 副本创建一个对应的 PV 数据卷。如果需要部署到不同节点分散 workload，需要通过 pod affinity 手动配置。

![img](img/clip_image004.png)

由于无状态应用 deployment 不能共享 block 数据卷，所以建议使用单副本。


## Q3: 如何运维一个 Kubernetes 节点?

HwameiStor 提供了数据卷驱逐和迁移功能。在移除或者重启一个 Kubernetes 节点的时候，可以将该节点上的 Pods 和数据卷自动迁移到其他可用节点上，并确保 Pods 持续运行并提供服务。

### 移除节点

为了确保 Pods 的持续运行，以及保证 HwameiStor 本地数据持续可用，在移除 Kubernetes 节点之前，需要将该节点上的 Pods 和本地数据卷迁移至其他可用节点。可以通过下列步骤进行操作

```
## Step 1:

$ kubectl drain NODE --ignore-daemonsets=true. --ignore-daemonsets=true
```

该命令可以将节点上的 Pods 驱逐，并重新调度。同时，也会自动触发 HwameiStor 的数据卷驱逐行为。HwameiStor 会自动将该节点上的所有数据卷副本迁移到其他节点，并确保数据仍然可用。可以通过下列命令检查迁移是否完成

```
## Step 2:

$ kubectl get localstoragenode NODE
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  name: NODE
spec:
  hostname: NODE
  storageIP: 10.6.113.22
  topogoly:
    region: default
    zone: default
status:
  ...
  pools:
    LocalStorage_PoolHDD:
      class: HDD
      disks:
      - capacityBytes: 17175674880
        devPath: /dev/sdb
        state: InUse
        type: HDD
      freeCapacityBytes: 16101933056
      freeVolumeCount: 999
      name: LocalStorage_PoolHDD
      totalCapacityBytes: 17175674880
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 1073741824
      usedVolumeCount: 1
      volumeCapacityBytesLimit: 17175674880
    ## **** 确保 volumes 字段为空 **** ##
      volumes:  
  state: Ready
```

同时，HwameiStor 会自动重新调度被驱逐的 Pods，将它们调度到有效数据卷所在的节点上，并确保 Pods 正常运行。

下列命令将节点移除 Kubernetes 集群

```
## Step 3:
$ kubectl delete nodes NODE
```

### 重启节点

重启节点通常需要很长的时间才能将节点恢复正常。在这期间，该节点上的所有 Pods 和本地数据都无法正常运行。这种情况对于一些应用（例如，数据库）来说，会产生巨大的代价，甚至不可接受。

HwameiStor 可以立即将 Pods 调度到其他数据卷所在的可用节点，并持续运行。对于使用 HwameiStor 多副本数据卷的 Pod，这一过程会非常迅速，大概需要～10秒（受 Kubernetes 的原生调度机制影响）；对于使用单副本数据卷的 Pod，这个过程所需时间依赖于数据卷迁移所需时间，受用户数据量大小影响。

如果用户希望将数据卷保留在该节点上，在节点重启后仍然可以访问，可以在节点上添加下列标签，阻止系统迁移该节点上的数据卷。系统仍然会立即将 Pod 调度到其他有数据卷副本的节点上。

```
## Step 1 (可选):
$ kubectl label node NODE hwameistor.io/eviction=disable
```

```
## Step 2:

$ kubectl drain NODE --ignore-daemonsets=true. --ignore-daemonsets=true
```

如果执行了 Step 1，待 Step 2 成功后，用户即可重启节点。
如果没有执行 Step 1，待 Step 2 成功后，用户察看数据迁移是否完成（方法如同“移除节点”的 Step 2）。待数据迁移完成后，即可重启节点。

在前两步成功之后，用户可以重启节点，并等待节点系统恢复正常。

之后，通过下列步骤将节点恢复至 Kubernetes 的正常状态。
```
## Step 3:
$ kubectl uncordon NODE
```


### 对于传统通用型共享存储

有状态应用 statefulSet 会将复制的副本优先部署到其他节点以分散 workload，但会为每一个 Pod 副本创建一个对应的 PV 数据卷。只有当副本数超过 Worker 节点数的时候会出现多个副本在同一个节点。

无状态应用 deployment 会将复制的副本优先部署到其他节点以分散 workload，并且所有的 Pod 共享一个 PV 数据卷（目前仅支持 NFS）。只有当副本数超过 Worker 节点数的时候会出现多个副本在同一个节点。对于 block 存储，由于数据卷不能共享，所以建议使用单副本。