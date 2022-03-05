### 前提条件

local-storage需要部署在Kuberntes系统中，需要集群满足下列条件：

* LocalStorage Version: `4.0+`
* Kubernetes Version: `1.18+`
* Node
  * 空闲磁盘
  * LVM (`可选`)

### 步骤 1: 选择和配置节点

部署local-storage之前，需要选择Kubernetes节点并且进行配置。这些节点会被加入local-storage系统。因此，这些节点要有空闲的磁盘。此外，还需要确定每个节点上的持久化数据卷类型，LVM, DISK 或者 RAM。配置为LVM/DISK的节点，还可以额外的配置RAM。这样，在该节点上，既可以创建LVM/DISK数据卷，也可以创建RAM数据卷。

``` bash
# 1. List all the kubernetes nodes
$ kubectl get nodes
NAME              STATUS   ROLES             AGE   VERSION
localstorage-10-6-161-21   Ready    master,registry   10d   v1.18.6
localstorage-10-6-161-25   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-26   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-27   Ready    <none>            10d   v1.18.6

# 2. Add local-storage config for each selected node as an annotation, key is "localstorage.hwameistor.io/local-storage-conf"
$ kubectl annotate node localstorage-10-6-161-27 localstorage.hwameistor.io/local-storage-conf='{"storage":{"volumeKind": "LVM", "ramdiskTotalCapacity": "1GB"}}'
node/localstorage-10-6-161-27 annotated

# 3. Add local-storage label for each selected node, key is "csi.driver.hwameistor.io/local-storage"
$ kubectl label node localstorage-10-6-161-27 csi.driver.hwameistor.io/local-storage=true
node/localstorage-10-6-161-27 labeled

# *** Important notes ***
# can NOT change the order of step 2 and 3
```

### 步骤 2: 部署local-storage、CSI Sidecars、scheduler

``` bash
# 0. checkout the code
$ git clone https://github.com/hwameistor/local-storage.git
$ cd local-storage

# 1. create a separate namespace for local-storage, e.g. local-storage-system
$ kubectl apply -f deploy/01_namespace.yaml

# 2. create a RBAC, limitrange in the namespace
$ kubectl apply -f deploy/02_rbac.yaml
$ kubectl apply -f deploy/03_limitsrange.yaml

# 3. deploy local-storage CRDs
$ kubectl apply -f deploy/crds

# 4. deploy local-storage cluster
$ kubectl apply -f deploy/05_cluster.yaml

# 5. deploy CSI sidecars
$ kubectl apply -f deploy/06_csi_controller.yaml

# 6. deploy scheduler
$ kubectl apply -f deploy/07_scheduler.yaml

# 7. check status of the deployment
$ kubectl -n local-storage-system get pod -o wide
NAME                                               READY   STATUS    RESTARTS   AGE   IP               NODE              NOMINATED NODE   READINESS GATES
hwameistor-csi-controller-0             3/3     Running   15         13h   172.29.54.20     localstorage-10-6-161-27   <none>           <none>
hwameistor-4b6n8                        3/3     Running   0          18m   10.6.161.27      localstorage-10-6-161-27   <none>           <none>
hwameistor-dv7nd                        3/3     Running   0          18m   10.6.161.26      localstorage-10-6-161-26   <none>           <none>
hwameistor-vzdqh                        3/3     Running   0          18m   10.6.161.25      localstorage-10-6-161-25   <none>           <none>
hwameistor-scheduler-6585bb5897-9xj85   1/1     Running   0          15h   172.29.164.160   localstorage-10-6-161-25   <none>           <none>


```

### 步骤 3: 创建 StorageClass

``` bash
# You should create a storageclass for each volume kind, i.e. LVM, DISK, RAM

# LVM volume storageclass (waitforfirstconsumer mode) with expansion capability
$ kubectl apply -f deploy/storageclass-lvm.yaml
# LVM volume support HA storageclass (waitforfirstconsumer mode) with expansion capability
$ kubectl apply -f deploy/storageclass-lvm-ha.yaml
# Disk volume storageclass (waitforfirstconsumer mode) without expansion capability
$ kubectl apply -f deploy/storageclass-disk.yaml
# RAMdisk volume storageclass (waitforfirstconsumer mode) without expansion capability
$ kubectl apply -f deploy/storageclass-ram.yaml

# check for storageclass
$ kubectl get sc
NAME                     PROVISIONER                 RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-storage-hdd-disk   localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  21d
local-storage-hdd-lvm    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-lvm-ha localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-ram    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  15d
```

### 步骤 4: 创建 non HA PVC

``` bash
# create a test PVC with LVM local volume
$ kubectl apply -f deploy/pvc-lvm.yaml

# check PVC status. It should be in Pending
$ kubectl get pvc
NAME                     STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-lvm    Pending                                      local-storage-hdd-lvm    3s
```

### 步骤 5: 部署 Pod

``` bash
# deploy a nginx application which uses the LVM local volume PVC
$ kubectl apply -f deploy/nginx-lvm.yaml

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   0/1     Pending   0          63s

$ kubectl get pvc
NAME                    STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Pending                                      local-storage-hdd-lvm   102s
```

### 步骤 6: 查看local-storage状态

``` bash
# check status of local-storage nodes
$ kubectl get lsn # localstoragenode
NAME                       VOLUMEKIND   RAMDISKQUOTA   ZONE      REGION    STATUS   AGE
localstorage-10-6-161-26   LVM          0              default   default   Ready    14d

# check status of local volume and volume replica
$ kubectl get lv # localvolume
NAME                                       POOL                   KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM    1          1073741824   k8s-node1       Creating                          2m50s
```

此时因为未分配磁盘, 由于节点存储容量不存在, pvc 处于Pending， lv 处于Creating状态
此时需要部署local disk manager 服务申请存储资源

### 步骤 7: 部署local disk manager 服务

 如何部署local-disk-manager, 请参考 [local-disk-manager](https://github.com/hwameistor/local-disk-manager/blob/main/README-zh.md)


查看local-storage节点上的每块物理磁盘的具体信息,及磁盘申请信息

``` bash
$ kubectl get ldc -A # localdiskclaim
NAMESPACE    NAME                      NODEMATCH   PHASE
hwameistor   localdiskclaim-sample-1   k8s-node1   Bound

$ kubectl get ld # localdisk
NAME             NODEMATCH    CLAIM                     PHASE
k8s-node1-sdb    k8s-node1    localdiskclaim-sample-1   Claimed
```

此时磁盘已经申请及成功分配给localstoragenode k8s-node1， 此时检查服务状态，均正常

```
$ kubectl get pvc # pvc
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Bound    pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   1Gi        RWO            local-storage-hdd-lvm   37m

#  check status of local volume and volume replica
$ kubectl get lv # localvolume
NAME                                       POOL                   KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM    1          1073741824   k8s-node1       Ready   -1                     22m

$ kubectl get lvr # localvolumereplica
NAME                                              KIND   CAPACITY     NODE        STATE   SYNCED   DEVICE                                                               AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2-v5scm9   LVM    1073741824   k8s-node1   Ready   true     /dev/LocalStorage_PoolHDD/pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   80s

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   1/1     Running   0          36m

```