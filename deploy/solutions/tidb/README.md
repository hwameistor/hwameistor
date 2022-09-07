
# HwameiStor-TiDB Cloud Native Solution

## Solution Introduction

HwameiStor is a high availability local storage system for cloud-native stateful workloads. It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes and provides data persistence capabilities for stateful cloud-native workloads or components.

TiDB is a cloud-native distributed database product with the abilities of expansion / contraction, financial-grade high availability, real-time HTAP (Hybrid Transactional and Analytical Processing) etc., and the ultimate goal is to provide users with one-stop OLTP (Online Transactional Processing), OLAP (Online Analytical Processing), and HTAP solutions. TiDB is suitable for various application scenarios such as high availability, high demand of consistency, and large data scale.

Local storage has the highest read and write performance among many storage protocols, which can undoubtedly provide performance guarantee for TiDB. HwameiStor is a storage system that meets the requirements of the cloud-native era. It has the advantages of high performance, high availability, automation, low cost, and rapid deployment, and can replace expensive traditional SAN storage.

## TiDB Components

The TiDB distributed database splits the overall architecture into multiple modules, which communicate with each other to form a complete TiDB system.

TiDB Server: The SQL layer, which exposes the connection endpoint of the MySQL protocol, is responsible for accepting connections from clients, performing SQL parsing and optimization, and finally generating a distributed execution plan.  The TiDB layer itself is stateless, and multiple TiDB instances can be started.  A unified access address is provided externally through load balancing components (such as LVS, HAProxy or F5), and client connections can be evenly distributed among multiple TiDB instances in order to achieve the effect of load balancing. TiDB Server itself does not store data, but only parses SQL and forwards the actual data read request to the underlying storage node TiKV (or TiFlash).

PD (Placement Driver) Server: The meta-data management module of the entire TiDB cluster, responsible for storing the real-time data distribution of each TiKV node and the overall topology of the cluster, also providing the TiDB Dashboard control interface, and assigning transaction IDs to distributed transactions. PD not only stores meta data, but also issues data scheduling commands to specific TiKV nodes according to the real-time data distribution status reported by TiKV nodes.  PD Server is so called the "brain" of the entire cluster. 

TiKV Server: Responsible for storing data with the basic data unit of Region. Each Region is responsible for storing the data of a Key Range (the left-closed and right-open interval from StartKey to EndKey). Each TiKV node is responsible for multiple Regions. TiKV's API provides native support for distributed transactions at the KV key-value pair level, and provides the SI (Snapshot Isolation) isolation level by default, which is also the core of TiDB's support for distributed transactions at the SQL level. After the SQL layer of TiDB completes the SQL parsing, it will convert the SQL execution plan into the actual call to the TiKV API. Therefore, the data is stored in TiKV. In addition, data in TiKV will automatically maintain multiple copies (the default is three copies), which naturally supports high availability and automatic failover.

## Solution Validation

### Testing Environment

This test uses three virtual machine nodes to deploy a Kubernetes cluster: 1 Master + 3 Worker nodes, and the kubelet version is 1.21.0.

```console
$ kubectl get no
NAME              STATUS   ROLES                  AGE   VERSION
k8s-10-6-163-51   Ready    <none>                 86d   v1.21.0
k8s-10-6-163-52   Ready    control-plane,master   87d   v1.21.0
k8s-10-6-163-53   Ready    <none>                 87d   v1.21.0
k8s-10-6-163-54   Ready    <none>                 29d   v1.21.0
```

Deploy HwameiStor local storage on Kubernetes.

```console
$ kubectl get all -n hwameistor
NAME                                                           READY   STATUS    RESTARTS   AGE
pod/hwameistor-admission-controller-56bbc5c9fc-5bptb           1/1     Running   1          8d
pod/hwameistor-local-disk-csi-controller-c7bdffcff-tnmmh       2/2     Running   670        8d
pod/hwameistor-local-disk-manager-4w4m2                        2/2     Running   161        8d
pod/hwameistor-local-disk-manager-cmzdk                        2/2     Running   156        8d
pod/hwameistor-local-disk-manager-mfb4z                        2/2     Running   15         8d
pod/hwameistor-local-disk-manager-mmq4h                        2/2     Running   141        8d
pod/hwameistor-local-storage-b6wmd                             2/2     Running   92         8d
pod/hwameistor-local-storage-c52ft                             2/2     Running   81         8d
pod/hwameistor-local-storage-csi-controller-86d55d6bdc-64wmc   3/3     Running   959        8d
pod/hwameistor-local-storage-gwx9b                             2/2     Running   87         8d
pod/hwameistor-local-storage-p2q7r                             2/2     Running   89         8d
pod/hwameistor-scheduler-68dc49bd69-hh4b8                      1/1     Running   318        8d

NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/hwameistor-admission-controller   ClusterIP   10.108.62.244   <none>        443/TCP             8d
service/local-disk-manager-metrics        ClusterIP   10.109.190.29   <none>        8383/TCP,8686/TCP   8d

NAME                                           DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/hwameistor-local-disk-manager   4         4         4       4            4           <none>          8d
daemonset.apps/hwameistor-local-storage        4         4         4       4            4           <none>          8d

NAME                                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/hwameistor-admission-controller           1/1     1            1           8d
deployment.apps/hwameistor-local-disk-csi-controller      1/1     1            1           8d
deployment.apps/hwameistor-local-storage-csi-controller   1/1     1            1           8d
deployment.apps/hwameistor-scheduler                      1/1     1            1           8d

NAME                                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/hwameistor-admission-controller-56bbc5c9fc           1         1         1       8d
replicaset.apps/hwameistor-local-disk-csi-controller-c7bdffcff       1         1         1       8d
replicaset.apps/hwameistor-local-storage-csi-controller-86d55d6bdc   1         1         1       8d
replicaset.apps/hwameistor-scheduler-68dc49bd69                      1         1         1       8d
```

View local storage disks status.

```console
$ kubectl get ld
NAME                  NODEMATCH         CLAIM   PHASE
k8s-10-6-163-51-sda   k8s-10-6-163-51           Inuse
k8s-10-6-163-51-sdb   k8s-10-6-163-51           Claimed
k8s-10-6-163-51-sdc   k8s-10-6-163-51           Claimed
k8s-10-6-163-51-sdd   k8s-10-6-163-51           Claimed
k8s-10-6-163-51-sde   k8s-10-6-163-51           Claimed
k8s-10-6-163-51-sdf   k8s-10-6-163-51           Claimed
k8s-10-6-163-52-sda   k8s-10-6-163-52           Inuse
k8s-10-6-163-52-sdb   k8s-10-6-163-52           Unclaimed
k8s-10-6-163-53-sda   k8s-10-6-163-53           Inuse
k8s-10-6-163-53-sdb   k8s-10-6-163-53           Claimed
k8s-10-6-163-53-sdc   k8s-10-6-163-53           Claimed
k8s-10-6-163-53-sdd   k8s-10-6-163-53           Claimed
k8s-10-6-163-53-sde   k8s-10-6-163-53           Claimed
k8s-10-6-163-53-sdf   k8s-10-6-163-53           Claimed
k8s-10-6-163-54-sda   k8s-10-6-163-54           Inuse
k8s-10-6-163-54-sdb   k8s-10-6-163-54           Unclaimed
k8s-10-6-163-54-sdc   k8s-10-6-163-54           Unclaimed
k8s-10-6-163-54-sdd   k8s-10-6-163-54           Unclaimed
k8s-10-6-163-54-sde   k8s-10-6-163-54           Unclaimed
k8s-10-6-163-54-sdf   k8s-10-6-163-54           Unclaimed
```

View StorageClass status

```console
$ kubectl get sc
NAME                         PROVISIONER         RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-lvm-hdd   lvm.hwameistor.io   Delete          WaitForFirstConsumer   true                   8d
```

### Deploy TiDB Operator

#### Install TiDB Operator CRDs

```console
$ kubectl create -f https://raw.githubusercontent.com/pingcap/tidb-operator/v1.3.7/manifests/crd.yaml

$ kubectl get crd | grep pingcap
backups.pingcap.com                                   2022-08-22T22:38:57Z
backupschedules.pingcap.com                           2022-08-22T22:38:56Z
dmclusters.pingcap.com                                2022-08-22T22:38:57Z
restores.pingcap.com                                  2022-08-22T22:38:57Z
tidbclusterautoscalers.pingcap.com                    2022-08-22T22:38:57Z
tidbclusters.pingcap.com                              2022-08-23T09:12:15Z
tidbinitializers.pingcap.com                          2022-08-22T22:38:58Z
tidbmonitors.pingcap.com                              2022-08-22T22:38:58Z
tidbngmonitorings.pingcap.com                         2022-08-22T22:38:58Z
```

#### Install TiDB Operator

```console
$ helm repo add pingcap https://charts.pingcap.org/

$ kubectl create namespace tidb-admin

$ helm install --namespace tidb-admin tidb-operator pingcap/tidb-operator --version v1.3.7 \
    --set operatorImage=registry.cn-beijing.aliyuncs.com/tidb/tidb-operator:v1.3.7 \
    --set tidbBackupManagerImage=registry.cn-beijing.aliyuncs.com/tidb/tidb-backup-manager:v1.3.7 \
    --set scheduler.kubeSchedulerImageName=registry.cn-hangzhou.aliyuncs.com/google_containers/kube-scheduler

NAME: tidb-operator
LAST DEPLOYED: Mon Jun  1 12:31:43 2020
NAMESPACE: tidb-admin
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Make sure tidb-operator components are running:


$ kubectl get po -n tidb-admin
NAME                                       READY   STATUS    RESTARTS   AGE
tidb-controller-manager-74c4f7758c-8nfhf   1/1     Running   13         23h
tidb-scheduler-85cb548c55-2q9kt            2/2     Running   16         23h
```

### Deploy TiDB Cluster & Monitor

#### Install TiDB Cluster

```console
$ kubectl create namespace tidb-cluster && \
    kubectl -n tidb-cluster apply -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/examples/basic-cn/tidb-cluster.yaml

namespace/tidb-cluster created
tidbcluster.pingcap.com/basic created
```

Install TiDB Monitor

```console
$ kubectl -n tidb-cluster apply -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/examples/basic-cn/tidb-monitor.yaml

tidbmonitor.pingcap.com/basic created

$ kubectl get po -n tidb-cluster
NAME                               READY   STATUS    RESTARTS   AGE
basic-discovery-54cc4bf9fb-ncvfc   1/1     Running   0          14h
basic-pd-0                         1/1     Running   0          14h
basic-tidb-0                       2/2     Running   0          14h
basic-tidb-1                       2/2     Running   0          14h
basic-tikv-0                       1/1     Running   0          14h
```

View PVCs on HwameiStor

```console
$ kubectl get pvc -n tidb-cluster
NAME                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
pd-basic-pd-0       Bound    pvc-e0bad105-eddf-45c6-9003-c64d04cfad3a   2Gi        RWO            hwameistor-storage-lvm-hdd   38h
tikv-basic-tikv-0   Bound    pvc-4bdb0e53-662c-4706-a992-59ec68f8b295   2Gi        RWO            hwameistor-storage-lvm-hdd   38h
```

#### Connect TiDB Cluster

```console
$ kubectl get svc -n tidb-cluster
NAME              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)               AGE
basic-discovery   ClusterIP   10.111.145.100   <none>        10261/TCP,10262/TCP   38h
basic-pd          ClusterIP   10.108.191.26    <none>        2379/TCP              38h
basic-pd-peer     ClusterIP   None             <none>        2380/TCP,2379/TCP     38h
basic-tidb        ClusterIP   10.100.183.144   <none>        4000/TCP,10080/TCP    38h
basic-tidb-peer   ClusterIP   None             <none>        10080/TCP             38h
basic-tikv-peer   ClusterIP   None             <none>        20160/TCP             38h

$ kubectl port-forward -n tidb-cluster svc/basic-tidb 14000:4000 > pf14000.out &

$ mysql -h 127.0.0.1 -P 4000 -u root
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 10691
Server version: 5.7.25-TiDB-v6.1.0 TiDB Server (Apache License 2.0) Community Edition, MySQL 5.7 compatible

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.
```
