---
sidebar_position: 3
sidebar_label: "独立安装部署"
---

# 独立安装部署

本页说明如何在 Kubernetes 节点独立安装部署 HwaweiStor 本地存储。

## 步骤 1：选择和配置节点

部署本地磁盘之前，需要先选择 Kubernetes 节点并进行配置。这些节点将加入本地磁盘系统。因此，这些节点要有空闲的磁盘。此外，还需要确定每个节点上的持久化数据卷类型：LVM、DISK 或 RAM。配置为 LVM/DISK 的节点，还可以额外配置 RAM。这样，在该节点上，既可以创建 LVM/DISK 数据卷，也可以创建 RAM 数据卷。

```bash
# 1. 列出所有 kubernetes 节点
$ kubectl get nodes
NAME					STATUS   ROLES             AGE   VERSION
localstorage-10-6-161-21   Ready    master,registry   10d   v1.18.6
localstorage-10-6-161-25   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-26   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-27   Ready    <none>            10d   v1.18.6

# 2. 为选择的每个节点添加本地磁盘标签，key 为 lvm.hwameistor.io/enable
$ kubectl label node localstorage-10-6-161-27 lvm.hwameistor.io/enable=true
node/localstorage-10-6-161-27 labeled

# ***  重要说明   ***
# 不要更改第 2 步和第 3 步的顺序
```

## 步骤 2：部署 local-storage、CSI Sidecar、调度器

```bash
# 0. 下载源代码
$ git clone https://github.com/hwameistor/local-storage.git
$ cd local-storage
# 1. 创建独立的命名空间，例如 local-storage-system
$ kubectl apply -f deploy/01_namespace.yaml
# 2. 在命名空间中创建 RBAC 和 limitrange
$ kubectl apply -f deploy/02_rbac.yaml
$ kubectl apply -f deploy/03_limitsrange.yaml
# 3. 部署 local-storage 自定义资源
$ kubectl apply -f deploy/crds
# 4. 部署 local-storage 集群
$ kubectl apply -f deploy/05_cluster.yaml
# 5. 部署 CSI sidecar
$ kubectl apply -f deploy/06_csi_controller.yaml
# 6. 部署调度器
$ kubectl apply -f deploy/07_scheduler.yaml
# 7. 检查无状态应用的状态
$ kubectl -n local-storage-system get pod -o wide
NAME							READY	STATUS	RESTARTS   AGE   IP               NODE              NOMINATED NODE   READINESS GATES
hwameistor-csi-controller-0			3/3	Running		15		13h   172.29.54.20     localstorage-10-6-161-27   <none>	<none>
hwameistor-local-storage-4b6n8		3/3	Running		0		18m   10.6.161.27      localstorage-10-6-161-27   <none>	<none>
hwameistor-local-storage-dv7nd		3/3	Running		0		18m   10.6.161.26      localstorage-10-6-161-26   <none>	<none>
hwameistor-local-storage-vzdqh		3/3	Running		0		18m   10.6.161.25      localstorage-10-6-161-25   <none>	<none>
hwameistor-scheduler-6585bb5897-9xj85 1/1 Running	0		 15h   172.29.164.160   localstorage-10-6-161-25   <none>	 <none>
```

## 步骤 3：创建 StorageClass

```bash
# 需要为每种存储卷（LVM、DISK、RAM）创建 storageclass
# LVM volume storageclass (waitforfistconsumer mode) 带扩容能力
$ kubectl apply -f deploy/storageclass-lvm.yaml
# Disk volume storageclass (waitforfistconsumer mode) 不带扩容能力
$ kubectl apply -f deploy/storageclass-disk.yaml
# RAMdisk volume storageclass (waitforfistconsumer mode) 不带扩容能力
$ kubectl apply -f deploy/storageclass-ram.yaml
# 检查 storageclass
$ kubectl get sc
NAME                     PROVISIONER                 RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-storage-hdd-disk   localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  21d
local-storage-hdd-lvm    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-lvm-ha localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-ram    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  15d
```

## 步骤 4：创建 PVC

```bash
# 用 LVM 本地卷创建测试 PVC
$ kubectl apply -f deploy/pvc-lvm.yaml

# 检查 PVC 状态是否为 Pending
$ kubectl get pvc
NAME                     STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-lvm    Pending                                      local-storage-hdd-lvm    3s
```

## 步骤 5：用 PVC 部署 Nginx

```bash
# 部署适用 LVM 本地卷 PVC 的 nginx 应用程序
$ kubectl apply -f deploy/nginx-lvm.yaml
$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   0/1     Pending   0          63s
$ kubectl get pvc
NAME                    STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Pending                                      local-storage-hdd-lvm   102s
```

## 步骤 6：检查 local-storage 的状态

```bash
# 检查 local-storage 节点的状态
$ kubectl get lsn # localstoragenode
NAME                       VOLUMEKIND   RAMDISKQUOTA   ZONE      REGION    STATUS   AGE
localstorage-10-6-161-26   LVM          0              default   default   Ready    14d

# 检查本地卷和卷副本的状态
$ kubectl get lv # localvolume
NAME	POOL	KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM	1	1073741824   k8s-node1  Creating	2m50s
```

此时因为未分配磁盘, 由于节点存储容量不存在，pvc 处于 Pending，lv 处于 Creating 状态。此时需要部署 local disk manager 服务申请存储资源。

## 步骤 7：部署 local disk manager 服务

如何部署 local-disk-manager，请参考 [local-disk-manager](../01features/01local-disk-manager.md)。

查看 local-storage 节点上的每块物理磁盘的具体信息及磁盘申请信息。

```bash
$ kubectl get ldc -A # localdiskclaim
NAMESPACE    NAME                      NODEMATCH   PHASE
hwameistor   localdiskclaim-sample-1   k8s-node1   Bound
$ kubectl get ld # localdisk
NAME             NODEMATCH    CLAIM                     PHASE
k8s-node1-sdb    k8s-node1    localdiskclaim-sample-1   Claimed
```

此时磁盘已经申请及成功分配给 localstoragenode k8s-node1，此时检查服务状态，均正常。

```bash
$ kubectl get pvc # pvc
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Bound    pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   1Gi        RWO            local-storage-hdd-lvm   37m
#  检查本地存储卷和存储卷副本的状态
$ kubectl get lv # localvolume
NAME				POOL                   KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM	1	1073741824   k8s-node1			Ready	-1	  22m

$ kubectl get lvr # localvolumereplica
NAME										KIND   CAPACITY     NODE        STATE   SYNCED   DEVICE			AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2-v5scm9   LVM    1073741824   k8s-node1   Ready   true     /dev/LocalStorage_PoolHDD/pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   80s

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   1/1     Running   0          36m
```
