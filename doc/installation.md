### Prerequisite

local-storage is a cloud native local storage system, which should be deployed in a Kubernetes cluster with the following requirements:

* Kubernetes Version: `1.22+`
* Node
  * Free disks
  * LVM (`Optional`)

### Step 1: Select and Configure Nodes

Before deploying local-storage, you must decide which Kubernetes nodes for it. The node must have the free disks and also be able to host the applications using the local volume. In addition, you must decide which kind of volume (`LVM`, or `RAM`) should be able to provision on each node. Besides, For the node which is already configured with `LVM`, you can still configure `RAM` volume for it by adding "ramdiskTotalCapacity" into the configuration as below.

``` bash
# 1. List all the kubernetes nodes
$ kubectl get nodes
NAME              STATUS   ROLES             AGE   VERSION
localstorage-10-6-161-21   Ready    master,registry   10d   v1.23.7
localstorage-10-6-161-25   Ready    <none>            10d   v1.23.7
localstorage-10-6-161-26   Ready    <none>            10d   v1.23.7
localstorage-10-6-161-27   Ready    <none>            10d   v1.23.7

# 2. Add local-storage label for each selected node, key is "lvm.hwameistor.io/enable"
$ kubectl label node localstorage-10-6-161-27 lvm.hwameistor.io/enable=true
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
hwameistor-local-storage-4b6n8                        3/3     Running   0          18m   10.6.161.27      localstorage-10-6-161-27   <none>           <none>
hwameistor-local-storage-dv7nd                        3/3     Running   0          18m   10.6.161.26      localstorage-10-6-161-26   <none>           <none>
hwameistor-local-storage-vzdqh                        3/3     Running   0          18m   10.6.161.25      localstorage-10-6-161-25   <none>           <none>
localstorage-local-storage-scheduler-6585bb5897-9xj85   1/1     Running   0          15h   172.29.164.160   localstorage-10-6-161-25   <none>           <none>
```

### Step 3: Create StorageClass

``` bash
# You should create a storageclass for each volume kind, i.e. LVM, RAM

# LVM volume storageclass (waitforfistconsumer mode) with expansion capability
$ kubectl apply -f deploy/storageclass-lvm.yaml

# check for storageclass
$ kubectl get sc
NAME                     PROVISIONER                 RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-storage-hdd-lvm    localstorage.hwameistor.io   Delete          WaitForFirstConsumer   true                   21d
```

### Step 4: Create PVC

``` bash
# create a test PVC with LVM local volume
$ kubectl apply -f deploy/pvc-lvm.yaml

# check PVC status. It should be in Pending
$ kubectl get pvc
NAME                     STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS             AGE
local-storage-pvc-lvm    Pending                                      local-storage-hdd-lvm    3s
```

### Step 5: Deploy Nginx with PVC

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

### Step 6: Check status of local-storage

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
 For the disk is not allocated, the pvc is Pending and the lv is Creating becauseof no storage capacity
local disk manager Services needs to deploy to claim storage resources

### Step 7: Deploy local disk manager service

To start using [local-disk-manager](https://github.com/hwameistor/local-disk-manager/blob/main/README.md)

check specific information for each physical disk on the local-storage node，and the disk claim information

``` bash
$ kubectl get ldc -A # localdiskclaim
NAMESPACE    NAME                      NODEMATCH   PHASE
hwameistor   localdiskclaim-sample-1   k8s-node1   Bound

$ kubectl get ld # localdisk
NAME             NODEMATCH    CLAIM                     PHASE
k8s-node1-sdb    k8s-node1    localdiskclaim-sample-1   Claimed
```

The disk has been applied and successfully assigned to localstoragenode k8s-node1, and checking the service status is normal

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
