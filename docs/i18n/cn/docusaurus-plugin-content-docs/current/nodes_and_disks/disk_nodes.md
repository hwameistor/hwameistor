---
sidebar_position: 3
sidebar_label: "磁盘节点扩展"
---

# 磁盘节点扩展

裸磁盘存储节点为应用提供裸磁盘类型的数据卷，并且维护了该存储节点上面的裸磁盘和裸磁盘数据卷的对应关系，
本页说明如何扩展这种磁盘节点。

## 步骤

### 1. 准备新的存储节点

在 Kubernetes 集群中新增一个节点，或者，选择一个已有的集群节点（非 HwameiStor 节点）。
本例中，所用的新增存储节点和磁盘信息如下所示：

- name: k8s-worker-2
- devPath: /dev/sdb
- diskType: SSD disk

新增节点已经成功加入 Kubernetes 集群之后，检查并确保下列 Pod 正常运行在该节点上，以及相关资源存在于集群中：

```console
$ kubectl get node
NAME           STATUS   ROLES            AGE     VERSION
k8s-master-1   Ready    master           96d     v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker           96h     v1.24.3-2+63243a96d1c393

$ kubectl -n hwameistor get pod -o wide | grep k8s-worker-2
hwameistor-local-disk-manager-sfsf1     2/2     Running   0     19h   10.6.128.150      k8s-worker-2   <none>  <none>

# 检查 LocalDiskNode 资源
$ kubectl get localdisknode k8s-worker-2
NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
k8s-worker-2                                              Ready    21d
```

### 2. 添加新增存储节点到 HwameiStor 系统

首先，需要将磁盘 sdb 的 `owner` 信息修改成 local-disk-manager，具体如下：

```console
$ kubectl edit ld localdisk-2307de2b1c5b5d051058bc1d54b41d5c
apiVersion: hwameistor.io/v1alpha1
kind: LocalDisk
metadata:
  name: localdisk-2307de2b1c5b5d051058bc1d54b41d5c
spec:
  devicePath: /dev/sdb
  nodeName: k8s-worker-2
+ owner: local-disk-manager
...
```

为增加存储节点创建资源 LocalStorageClaim，以此为新增存储节点构建存储池。这样，节点就已经成功加入 HwameiStor 系统。具体如下：

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-2
spec:
  nodeName: k8s-worker-2
  owner: local-disk-manager
  description:
    diskType: SSD
EOF
```

### 3. 后续检查

完成上述步骤后，检查新增存储节点及其存储池的状态，确保节点和 HwameiStor 系统的正常运行。具体如下：

```console
$ kubectl get localdisknode k8s-worker-2 -o yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskNode
metadata:
  name: k8s-worker-2
spec:
  nodeName: k8s-worker-2  
status:  
  pools:
    LocalDisk_PoolSSD:
      class: SSD
      disks:
      - capacityBytes: 214744170496
        devPath: /dev/sdb
        state: Available
        type: SSD
      freeCapacityBytes: 214744170496
      freeVolumeCount: 1     
      totalCapacityBytes: 214744170496
      totalVolumeCount: 1
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 214744170496
      volumes:
  state: Ready
```
