---
sidebar_position: 2
sidebar_label: "数据盘扩展"
---

# 数据盘扩展

当存储系统中的某个节点存储容量不足时，可以通过为该节点增加磁盘扩充容量。
在 HwameiStor 中，可以通过下列步骤完成为节点增加磁盘（数据盘）。

## 步骤

### 1. 准备新的存储磁盘

从 HwameiStor 中选择需要扩容的节点，将新增磁盘插入该节点的磁盘槽位。
本例中，所用的新增存储节点和磁盘信息如下所示：

- name: k8s-worker-4
- devPath: /dev/sdc
- diskType: SSD

在新增磁盘被插入到 HwameiStor 存储节点 `k8s-worker-4` 后，检查该节点上的新磁盘状态，如下：

```console
# 1. 检查新增磁盘是否成功插入节点，并被正确识别
$ ssh root@k8s-worker-4
$ lsblk | grep sdc
sdc        8:32     0     20G  1 disk

# 2. 检查 HwameiStor 是否为新增磁盘正确创建资源 LocalDisk，并且状态为 `Unclaimed`
$ kubectl get localdisk | grep k8s-worker-4 | grep sdc
k8s-worker-4-sdc   k8s-worker-4       Unclaimed 
```

### 2. 将新增磁盘加入到节点的存储池

通过创建资源 LocalDiskClaim，将新增磁盘加入节点的存储池。如下所示。完成下列操作后，新磁盘应该被自动加入节点的 SSD 存储池中。如果该节点上没有 SSD 存储池，HwameiStor 会为其自动创建，并将新磁盘加入其中。

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-4-expand
spec:
  nodeName: k8s-worker-4
  description:
    diskType: SSD
EOF
```

### 3. 后续检查

完成上述步骤后，检查新增磁盘及其存储池的状态，确保节点和 HwameiStor 系统的正常运行。具体如下：

```console
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  name: k8s-worker-4
spec:
  hostname: k8s-worker-4
  storageIP: 10.6.182.103
  topogoly:
    region: default
    zone: default
status:
  pools:
    LocalStorage_PoolSSD:
      class: SSD
      disks:
      - capacityBytes: 214744170496
        devPath: /dev/sdb
        state: InUse
        type: SSD
      - capacityBytes: 214744170496
        devPath: /dev/sdc
        state: InUse
        type: SSD
      freeCapacityBytes: 429488340992
      freeVolumeCount: 1000
      name: LocalStorage_PoolSSD
      totalCapacityBytes: 429488340992
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 429488340992
      volumes:
  state: Ready
```
