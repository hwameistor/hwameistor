---
sidebar_position: 3
sidebar_label: "Install Independently"
---

# Install Independently

This page explains how you can independently install the HwaweiStor local storage on a Kubernetes node.

## Step 1: Select and configure nodes

Before installing a local-storage, you should determine which Kubernetes nodes the storage will run on. These nodes will be added to the local-storage. The nodes should have free disks and also be able to host applications using local volumes. In addition, you should determine which kind of volume (LVM, DISK, or RAM) should be able to provision on each node. Besides, for the node which is already configured with LVM or DISK, you can still configure RAM volume for it by adding "ramdiskTotalCapacity" into the configuration as below.

```bash
# 1. List all the kubernetes nodes
$ kubectl get nodes
NAME                       STATUS   ROLES             AGE   VERSION
localstorage-10-6-161-21   Ready    master,registry   10d   v1.18.6
localstorage-10-6-161-25   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-26   Ready    <none>            10d   v1.18.6
localstorage-10-6-161-27   Ready    <none>            10d   v1.18.6

# 2. Add a local-storage label for each selected node and the key is "lvm.hwameistor.io/enable"
$ kubectl label node localstorage-10-6-161-27 lvm.hwameistor.io/enable=true
node/localstorage-10-6-161-27 labeled

# *** Important notes ***
# Do NOT change the sequence of step 2 and 3
```

## Step 2: Deploy local-storage, CSI sidecars, and scheduler

```bash
# 0. Check out the code
$ git clone https://github.com/hwameistor/local-storage.git
$ cd local-storage

# 1. Create a separate namespace for local-storage, such as local-storage-system
$ kubectl apply -f deploy/01_namespace.yaml

# 2. Create a RBAC and limitrange in the namespace
$ kubectl apply -f deploy/02_rbac.yaml
$ kubectl apply -f deploy/03_limitsrange.yaml

# 3. Deploy local-storage CRDs
$ kubectl apply -f deploy/crds

# 4. Deploy a local-storage cluster
$ kubectl apply -f deploy/05_cluster.yaml

# 5. Deploy CSI sidecars
$ kubectl apply -f deploy/06_csi_controller.yaml

# 6. Deploy scheduler
$ kubectl apply -f deploy/07_scheduler.yaml

# 7. Check status of the deployment
$ kubectl -n local-storage-system get pod -o wide
NAME							    READY STATUS  RESTARTS   AGE   IP               NODE              NOMINATED NODE   READINESS GATES
hwameistor-csi-controller-0			3/3	Running		15		13h   172.29.54.20     localstorage-10-6-161-27   <none>	<none>
hwameistor-local-storage-4b6n8		3/3	Running		0		18m   10.6.161.27      localstorage-10-6-161-27   <none>	<none>
hwameistor-local-storage-dv7nd		3/3	Running		0		18m   10.6.161.26      localstorage-10-6-161-26   <none>	<none>
hwameistor-local-storage-vzdqh		3/3	Running		0		18m   10.6.161.25      localstorage-10-6-161-25   <none>	<none>
hwameistor-scheduler-6585bb5897-9xj85 1/1 Running	0		15h   172.29.164.160   localstorage-10-6-161-25   <none>	 <none>
```

## Step 3: Create StorageClass

```bash
# You should create a storageclass for each volume kind, such as LVM, DISK, and RAM

# LVM volume storageclass (waitforfistconsumer mode) with the expansion capability
$ kubectl apply -f deploy/storageclass-lvm.yaml
# Disk volume storageclass (waitforfistconsumer mode) without the expansion capability
$ kubectl apply -f deploy/storageclass-disk.yaml
# RAMdisk volume storageclass (waitforfistconsumer mode) without the expansion capability
$ kubectl apply -f deploy/storageclass-ram.yaml

# check for storageclass
$ kubectl get sc
NAME                     PROVISIONER                 RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-storage-hdd-disk   localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  21d
local-storage-hdd-lvm    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-lvm-ha localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
local-storage-hdd-ram    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   false                  15d
```

## Step 4: Create PVC

```bash
# Create a test PVC with LVM local volume
$ kubectl apply -f deploy/pvc-lvm.yaml

# Check the PVC status, which should be in Pending
$ kubectl get pvc
NAME                     STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-lvm    Pending                                      local-storage-hdd-lvm    3s
```

## Step 5: Deploy Nginx with PVC

```bash
# Deploy a nginx application which uses the LVM local volume PVC
$ kubectl apply -f deploy/nginx-lvm.yaml

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   0/1     Pending   0          63s

$ kubectl get pvc
NAME                    STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Pending                                      local-storage-hdd-lvm   102s
```

## Step 6: Check status of local-storage

```bash
# Check status of local-storage nodes
$ kubectl get lsn # localstoragenode
NAME                       VOLUMEKIND   RAMDISKQUOTA   ZONE      REGION    STATUS   AGE
localstorage-10-6-161-26   LVM          0              default   default   Ready    14d

# Check status of local volume and volume replica
$ kubectl get lv # localvolume
NAME					POOL	KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM	1	1073741824   k8s-node1  Creating	2m50s
```

For the disk that is not allocated, the status of PVC is in pending and that of LV is in creating. Because of no storage capacity, you should deploy the local disk manager service to claim more storage resources.

## Step 7: Deploy local disk manager service

For installing local-disk-manager, refer to [local-disk-manager](../01features/01local-disk-manager.md).

Check specific information for each physical disk on the local-storage node, and check the disk claim information.

```bash
$ kubectl get ldc -A # localdiskclaim
NAMESPACE    NAME                      NODEMATCH   PHASE
hwameistor   localdiskclaim-sample-1   k8s-node1   Bound
$ kubectl get ld # localdisk
NAME             NODEMATCH    CLAIM                     PHASE
k8s-node1-sdb    k8s-node1    localdiskclaim-sample-1   Claimed
```

The disk has been applied and successfully assigned to localstoragenode k8s-node1, and the service status is normal.

```bash
$ kubectl get pvc # pvc
NAME				STATUS   VOLUME	CAPACITY   ACCESS MODES   STORAGECLASS            AGE
local-storage-pvc-lvm   Bound    pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   1Gi        RWO            local-storage-hdd-lvm   37m
#  Check status of local volume and volume replica
$ kubectl get lv # localvolume
NAME				POOL                   KIND   REPLICAS   CAPACITY     ACCESSIBILITY   STATE      RESOURCE   PUBLISHED   AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   LocalStorage_PoolHDD   LVM	1	1073741824   k8s-node1			Ready	-1	  22m

$ kubectl get lvr # localvolumereplica
NAME							KIND   CAPACITY     NODE        STATE   SYNCED   DEVICE			AGE
pvc-996b05e8-80f2-4240-ace4-5f5f250310e2-v5scm9   LVM    1073741824   k8s-node1   Ready   true     /dev/LocalStorage_PoolHDD/pvc-996b05e8-80f2-4240-ace4-5f5f250310e2   80s

$ kubectl get pod
NAME                                       READY   STATUS    RESTARTS   AGE
nginx-local-storage-lvm-86d8c884c9-q58kq   1/1     Running   0          36m
```