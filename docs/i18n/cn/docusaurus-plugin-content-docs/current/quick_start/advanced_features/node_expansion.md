---
sidebar_position: 1
sidebar_label: "节点扩展"
---

# 节点扩展

存储系统可以通过增加存储节点实现扩容。在 HwameiStor 里，通过下列步骤可以添加新的存储节点。

## 步骤

### 1. 准备新的存储节点

在 Kubernetes 集群中新增一个节点，或者，选择一个已有的集群节点（非 HwameiStor 节点）。
该节点必须满足 [Prerequisites](../install/prereq.md) 要求的所有条件。
本例中，所用的新增存储节点和磁盘信息如下所示：

- name: k8s-worker-4
- devPath: /dev/sdb
- diskType: SSD disk

新增节点已经成功加入 Kubernetes 集群之后，检查并确保下列 Pods 正常运行在该节点上，以及相关资源存在于集群中：

```console
$ kubectl get node
NAME           STATUS   ROLES            AGE     VERSION
k8s-master-1   Ready    master           96d     v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-3   Ready    worker           96d     v1.24.3-2+63243a96d1c393
k8s-worker-4   Ready    worker           1h      v1.24.3-2+63243a96d1c393

$ kubectl -n hwameistor get pod -o wide | grep k8s-worker-4
hwameistor-local-disk-manager-c86g5     2/2     Running   0     19h   10.6.182.105      k8s-worker-4   <none>  <none>
hwameistor-local-storage-s4zbw          2/2     Running   0     19h   192.168.140.82    k8s-worker-4   <none>  <none>

# 检查 LocalStorageNode 资源
$ kubectl get localstoragenode k8s-worker-4
NAME                 IP           ZONE      REGION    STATUS   AGE
k8s-worker-4   10.6.182.103       default   default   Ready    8d
```

### 2. 添加新增存储节点到 HwameiStor 系统

为增加存储节点创建资源 LocalStorageClaim，以此为新增存储节点构建存储池。这样，节点就已经成功加入 HwameiStor 系统。具体如下：

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-4
spec:
  nodeName: k8s-worker-4
  description:
    diskType: SSD
EOF
```

### 3. 后续检查

完成上述步骤后，检查新增存储节点及其存储池的状态，确保节点和 HwameiStor 系统的正常运行。具体如下：

```console
$ kubectl get localstoragenode k8s-worker-4
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
      freeCapacityBytes: 214744170496
      freeVolumeCount: 1000
      name: LocalStorage_PoolSSD
      totalCapacityBytes: 214744170496
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 214744170496
      volumes:
  state: Ready
```
