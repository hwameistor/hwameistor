# Local Storage System (local-storage)

English | [Simplified_Chinese](https://github.com/hwameistor/local-storage/blob/main/README_zh.md)

## Introduction

Local Storage System is a cloud native storage system. It manages the free disks of each node and provision high performance persistent volume with local access to application.

Support local volume kind: `LVM`, `Disk`, `RAMDisk`.

Support disk type: `HDD`, `SSD`, `NVMe`, `RAMDisk`.

## 软件架构图

![image](https://github.com/hwameistor/local-storage/blob/main/HwameiStor-arch.png)

## Features and Roadmap

|        Feature       |   Status      |  Release  |   TP Date  |    GA Date   |                Description               |
| :---: | :---: | :---: | :---: | :---: | :---: |
|    non-HA LVM volume      |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  volume by LVM |
|    non-HA Disk volume   |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  volume by a physical disk |
|    non-HA RAM Disk volume   |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  volume by a ram disk |
|      CSI Driver      |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  basic CSI driver for dynamic provision |
| pod/volume schedule  |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  schedule pod to the node where the volume locates |
| disk health monitor  |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  monitor disk and predict failure |
| non-HA LVM Volume expansion |   Completed   |   v1.0    |  2020.Q3   |   2020.Q4    |  expand LVM volume capacity online |
| non-HA Disk volume expansion  |   NotSupport   |   |    |    |  expand Disk volume capacity |
| non-HA Disk volume snapshot   |   NotSupoort   |   |    |    |  snapshot of Disk volume |
| non-HA Disk volume snapshot restore |   NotSupport   |    |    |    |  restore Disk volume from snapshot |
| non-HA Disk volume clone     |   NotPlaned   |      |    |      |  clone Disk volume |
| HA Volume     |   Planed   |      |  2020.Q4  |   2021.Q1   |  Volume with HA |
| Replace Disk    |   Planed   |      |  2022.Q1  |   2022.Q1   |  Replace Disk |
| Observability    |   Planed   |      |  2022.Q2  |   2022.Q2   |  Observability |
| non-HA LVM volume stripe writing support    |   Planed   |      |  2022.Q2  |   2022.Q2   |  non-HA LVM volume stripe writing support |
| non-HA LVM volume thin provision support    |   Planed   |      |  2022.Q2  |   2022.Q2   |  non-HA LVM volume thin provision support |
| non-HA LVM volume snapshot  |   Planed   |       |  2022.Q2   |   2022.Q2    |  snapshot of LVM volume |
| non-HA LVM volume snapshot restore |   Planed   |      |  2022.Q2   |   2022.Q2 |  restore LVM volume from snapshot |
| non-HA LVM volume clone     |   Planed   |   |  2022.Q3  |  2022.Q3    |  clone LVM volume |
| non-HA LVM volume mirror     |   Planed   |   |  2022.Q3  |  2022.Q3    |  mirror LVM volume |
| volume backup     |   Planed   |      |  2022.Q3  |  2022.Q3    |  backup volume to external S3 |
| performance optimization     |   Planed   |      |  2022.Q3  |  2022.Q3    |  performance optimization |

## Use Cases

Currently, local-storage offers high performance local volume without HA. It's one of best data persistent solution for the following use cases:

* ***Database*** with HA capability, such as MySQL, OceanBase, MongoDB, etc..
* ***Messaging system*** with HA capability, such as Kafka, RabbitMQ, etc..
* ***Key-value store*** with HA capability, such as Redis, etc..
* ***Distributed storage system***, such as MinIO, Ozone, etc..
* Others with HA capability

## Usage

This is for developing or test, and will deploy local-storage from github repo.

### Prerequisite

local-storage is a cloud native local storage system, which should be deployed in a Kubernetes cluster with the following requirements:

* LocalStorage Version: `4.0+`
* Kubernetes Version: `1.18+`
* Node
  * Free disks
  * LVM (`Optional`)

### Step 1: Select and Configure Nodes

Before deploying local-storage, you must decide which Kubernetes nodes for it. The node must have the free disks and also be able to host the applications using the local volume. In addition, you must decide which kind of volume (`LVM`, `DISK` or `RAM`) should be able to provision on each node. Besides, For the node which is already configured with `LVM` or `DISK`, you can still configure `RAM` volume for it by adding "ramdiskTotalCapacity" into the configuration as below.

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

# 3. Add local-storage label for each selected node, key is "localstorage.hwameistor.io/local-storage"
$ kubectl label node localstorage-10-6-161-27 localstorage.hwameistor.io/local-storage=true
node/localstorage-10-6-161-27 labeled

# *** Important notes ***
# can NOT change the order of step 2 and 3
```

### Step 2: Deploy local-storage, CSI Sidecars, and scheduler

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
localstorage-local-storage-csi-controller-0             3/3     Running   15         13h   172.29.54.20     localstorage-10-6-161-27   <none>           <none>
localstorage-local-storage-4b6n8                        3/3     Running   0          18m   10.6.161.27      localstorage-10-6-161-27   <none>           <none>
localstorage-local-storage-dv7nd                        3/3     Running   0          18m   10.6.161.26      localstorage-10-6-161-26   <none>           <none>
localstorage-local-storage-vzdqh                        3/3     Running   0          18m   10.6.161.25      localstorage-10-6-161-25   <none>           <none>
localstorage-local-storage-scheduler-6585bb5897-9xj85   1/1     Running   0          15h   172.29.164.160   localstorage-10-6-161-25   <none>           <none>
```

### Step 3: Create StorageClass

``` bash
# You should create a storageclass for each volume kind, i.e. LVM, DISK, RAM

# LVM volume storageclass (waitforfistconsumer mode) with expansion capability
$ kubectl apply -f deploy/storageclass-lvm.yaml
# Disk volume storageclass (waitforfistconsumer mode) without expansion capability
$ kubectl apply -f deploy/storageclass-disk.yaml
# RAMdisk volume storageclass (waitforfistconsumer mode) without expansion capability
$ kubectl apply -f deploy/storageclass-ram.yaml

# check for storageclass
$ kubectl get sc
NAME                     PROVISIONER                 RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-storage-hdd-disk   localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  21d
local-storage-hdd-lvm    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-ram    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  15d
```

### Step 4: Create PVC

``` bash
# create a test PVC with LVM local volume
$ kubectl apply -f deploy/pvc-lvm.yaml
# create a test PVC with Disk local volume
$ kubectl apply -f deploy/pvc-disk.yaml
# create a test PVC with RAM local volume
$ kubectl apply -f deploy/pvc-ram.yaml

# check PVC status. It should be in Pending
$ kubectl get pvc
NAME                     STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-lvm    Pending                                      local-storage-hdd-lvm    3s
local-storage-pvc-disk   Pending                                      local-storage-hdd-disk   3s
local-storage-pvc-ram    Pending                                      local-storage-hdd-ram    3s
```

### Step 5: Deploy Nginx with PVC

``` bash
# deploy a nginx application which uses the LVM local volume PVC
$ kubectl apply -f deploy/nginx-lvm.yaml
# deploy a nginx application which uses the Disk local volume PVC
$ kubectl apply -f deploy/nginx-disk.yaml
# deploy a nginx application which uses the RAMdisk local volume PVC
$ kubectl apply -f deploy/nginx-ram.yaml

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-disk-fcc89fd9-5jrqb    1/1     Running   0          15d
nginx-local-storage-lvm-759d7d7489-9268f   1/1     Running   0          17d
nginx-local-storage-ram-64d67c975d-4bmt5   1/1     Running   0          15d

$ kubectl get pvc
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-disk   Bound    pvc-33f86c60-a80a-45aa-bbec-7a234fc9f5bb   100Gi      RWO            local-storage-hdd-disk   15d
local-storage-pvc-lvm    Bound    pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85   9Gi        RWO            local-storage-hdd-lvm    17d
local-storage-pvc-ram    Bound    pvc-490ce626-869b-491f-870a-da373704bed5   50Mi       RWO            local-storage-hdd-ram    15d
```

### Step 6: Check status of local-storage

``` bash
# check status of local-storage nodes
$ kubectl get lsn # localstoragenode
NAME              VOLUMEKIND   RAMDISKQUOTA   ZONE      REGION    STATUS   AGE
localstorage-10-6-161-25   DISK         1073741824     default   default   Ready    14d
localstorage-10-6-161-26   LVM          0              default   default   Ready    14d
localstorage-10-6-161-27   LVM          0              default   default   Ready    14d

# check status of local volume and volume replica
$ kubectl get lv # localvolume
NAME                                       POOL                   KIND   REPLICANUMBER   REQUIRED     ACCESSIBILITY     DELETE   STATE   SYNCED   ALLOCATED      REPLICAS                                           PUBLISHED         AGE
pvc-33f86c60-a80a-45aa-bbec-7a234fc9f5bb   LocalStorage_PoolHDD   DISK   1               5368709120   localstorage-10-6-161-25   false    Ready   true     107374182400   [pvc-33f86c60-a80a-45aa-bbec-7a234fc9f5bb-kzz6v]   localstorage-10-6-161-25   15d
pvc-490ce626-869b-491f-870a-da373704bed5   LocalStorage_PoolRAM   RAM    1               52428800     localstorage-10-6-161-25   false    Ready   true     52428800       [pvc-490ce626-869b-491f-870a-da373704bed5-cwqrm]   localstorage-10-6-161-25   15d
pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85   LocalStorage_PoolHDD   LVM    1               8589934592   localstorage-10-6-161-26   false    Ready   true     8594128896     [pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85-h6qrq]   localstorage-10-6-161-26   17d

$ kubectl get lvr # localvolumereplica
NAME                                             KIND   REQUIRED     NODE              DELETE   STATE   SYNCED   ALLOCATED      STORAGE    DEVICE                                                                   AGE
pvc-33f86c60-a80a-45aa-bbec-7a234fc9f5bb-kzz6v   DISK   5368709120   localstorage-10-6-161-25   false    Ready   true     107374182400   /dev/sdf   /dev/LocalStorage_DiskPoolHDD/pvc-33f86c60-a80a-45aa-bbec-7a234fc9f5bb   15d
pvc-490ce626-869b-491f-870a-da373704bed5-cwqrm   RAM    52428800     localstorage-10-6-161-25   false    Ready   true     52428800       ramdisk    /dev/LocalStorage_PoolRAM/pvc-490ce626-869b-491f-870a-da373704bed5       15d
pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85-h6qrq   LVM    8589934592   localstorage-10-6-161-26   false    Ready   true     8594128896                /dev/LocalStorage_PoolHDD/pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85       17d
```

### Step 7: Check detail info including health of each physical disk

``` bash
$ k get ld # localdisk
NAME              NODEMATCH    CLAIM   PHASE
k8s-master-dm-0   k8s-master           Unclaimed
k8s-master-dm-1   k8s-master           Unclaimed
k8s-master-sda    k8s-master           Unclaimed
k8s-master-sdb    k8s-master           Unclaimed
k8s-master-sr0    k8s-master           Unclaimed
k8s-node1-dm-0    k8s-node1            Unclaimed
k8s-node1-dm-1    k8s-node1            Unclaimed
k8s-node1-sda     k8s-node1            Unclaimed
k8s-node1-sdb     k8s-node1            Unclaimed
k8s-node1-sdc     k8s-node1            Unclaimed
k8s-node1-sdd     k8s-node1            Unclaimed
k8s-node1-sr0     k8s-node1            Unclaimed
```

## Feedbacks

Please submit any feedback and issue at: [Issues](https://github.com/hwameistor/local-storage/issues)
