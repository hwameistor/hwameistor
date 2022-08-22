
# HwameiStor-MinIO Cloud Native Solution

## Solution Introduction

HwameiStor is a high availability local storage system for cloud-native stateful workloads. It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes and provides data persistence capabilities for stateful cloud-native workloads or components.

MinIO is a High Performance Object Storage released under GNU Affero General Public License v3.0. It is API-compatible with Amazon S3 cloud storage service， and can be used to build high performance infrastructure for machine learning, analytics and application data workloads.

MinIO is usually used as the main storage for cloud-native applications which require higher throughput and lower latency. MinIO's read/write performance can achieve up to the speed of 183 GB/sec and 171 GB/sec.

The ultimate high performance of MinIO is inseparable from the underlying storage platform.  Local storage has been recognized as the highest r/w performance protocol among many other storage protocols, which can undoubtedly provide performance guarantee for MinIO.  HwameiStor is such a local storage system that waves with the cloud-native era. High performance, high availability, automation, low cost, rapid deployment, and more benefits attract you to replace those expensive traditional storage schemes such as SAN (Storage Area Network).

## Solution Validation

Testing Environment

This test uses three virtual machine nodes to deploy a Kubernetes cluster: 1 Master + 3 Worker nodes, and the kubelet version is 1.22.0.

```console
$ kubectl get node
NAME           STATUS   ROLES            AGE     VERSION
k8s-master-1   Ready    master           96d     v1.24.3-2+63243a96d1c393
k8s-worker-1   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-2   Ready    worker           96h     v1.24.3-2+63243a96d1c393
k8s-worker-3   Ready    worker           96d     v1.24.3-2+63243a96d1c393
```

Deploy HwameiStor local storage on Kubernetes.

```console
# kubectl get all -nhwameistor
NAME                                                           READY   STATUS    RESTARTS   AGE
pod/hwameistor-admission-controller-56bbc5c9fc-5bptb           1/1     Running   1          45h
pod/hwameistor-local-disk-csi-controller-c7bdffcff-tnmmh       2/2     Running   272        45h
pod/hwameistor-local-disk-manager-4w4m2                        2/2     Running   49         38h
pod/hwameistor-local-disk-manager-cmzdk                        2/2     Running   49         40h
pod/hwameistor-local-disk-manager-mfb4z                        2/2     Running   15         45h
pod/hwameistor-local-disk-manager-mmq4h                        2/2     Running   40         38h
pod/hwameistor-local-storage-b6wmd                             2/2     Running   24         45h
pod/hwameistor-local-storage-c52ft                             2/2     Running   21         45h
pod/hwameistor-local-storage-csi-controller-86d55d6bdc-64wmc   3/3     Running   378        45h
pod/hwameistor-local-storage-gwx9b                             2/2     Running   24         45h
pod/hwameistor-local-storage-p2q7r                             2/2     Running   28         45h
pod/hwameistor-scheduler-68dc49bd69-hh4b8                      1/1     Running   124        45h


NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/hwameistor-admission-controller   ClusterIP   10.108.62.244   <none>        443/TCP             45h
service/local-disk-manager-metrics        ClusterIP   10.109.190.29   <none>        8383/TCP,8686/TCP   45h

NAME                                           DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/hwameistor-local-disk-manager   4         4         4       4            4           <none>          45h
daemonset.apps/hwameistor-local-storage        4         4         4       4            4           <none>          45h

NAME                                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/hwameistor-admission-controller           1/1     1            1           45h
deployment.apps/hwameistor-local-disk-csi-controller      1/1     1            1           45h
deployment.apps/hwameistor-local-storage-csi-controller   1/1     1            1           45h
deployment.apps/hwameistor-scheduler                      1/1     1            1           45h

NAME                                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/hwameistor-admission-controller-56bbc5c9fc           1         1         1       45h
replicaset.apps/hwameistor-local-disk-csi-controller-c7bdffcff       1         1         1       45h
replicaset.apps/hwameistor-local-storage-csi-controller-86d55d6bdc   1         1         1       45h
replicaset.apps/hwameistor-scheduler-68dc49bd69                      1         1         1       45h
```

View local storage disks status

```console
# kubectl get ld
NAME                NODEMATCH         CLAIM   PHASE
k8s-master-1-sda   k8s-master-1               Inuse
k8s-master-1-sdb   k8s-master-1               Claimed
k8s-master-1-sdc   k8s-master-1               Claimed
k8s-master-1-sdd   k8s-master-1               Claimed
k8s-master-1-sde   k8s-master-1               Claimed
k8s-master-1-sdf   k8s-master-1               Claimed
k8s-worker-1-sda   k8s-worker-1               Inuse
k8s-worker-1-sdb   k8s-worker-1               Unclaimed
k8s-worker-2-sda   k8s-worker-2               Inuse
k8s-worker-2-sdb   k8s-worker-2               Claimed
k8s-worker-2-sdc   k8s-worker-2               Claimed
k8s-worker-2-sdd   k8s-worker-2               Claimed
k8s-worker-2-sde   k8s-worker-2               Claimed
k8s-worker-2-sdf   k8s-worker-2               Claimed
k8s-worker-3-sda   k8s-worker-3               Inuse
k8s-worker-3-sdb   k8s-worker-3               Unclaimed
k8s-worker-3-sdc   k8s-worker-3               Unclaimed
k8s-worker-3-sdd   k8s-worker-3               Unclaimed
k8s-worker-3-sde   k8s-worker-3               Unclaimed
k8s-worker-3-sdf   k8s-worker-3               Unclaimed
```

View StorageClass status

```console
# kubectl get sc
NAME                         PROVISIONER         RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-lvm-hdd   lvm.hwameistor.io   Delete          WaitForFirstConsumer   true                   45h
```

## Deploy MinIO on Kubernetes

Deploy based on official helm chart and install MinIO chart.

```console
$ helm repo add minio <https://helm.min.io>

$ helm repo list | grep minio
minio          <https://helm.min.io/>
```

### Standalone mode deployment

```console
$ helm install minio-2 \
  --namespace minio-2 --create-namespace \
  --set accessKey=admin,secretKey=admin123 \
  --set mode=standalone \
  --set service.type=NodePort \
  --set persistence.enabled=true \
  --set persistence.size=2Gi \
  --set persistence.storageClass=hwameistor-storage-lvm-hdd \
  minio/minio

$ kubectl get all -nminio-2
NAME                           READY   STATUS    RESTARTS   AGE
pod/minio-2-785f5c9985-7f5pf   1/1     Running   0          97m

NAME              TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
service/minio-2   NodePort   10.104.40.2   <none>        9000:32000/TCP   97m

NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/minio-2   1/1     1            1           97m

NAME                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/minio-2-785f5c9985   1         1         1       97m
```

View PVCs on HwameiStor

```console
$ kubectl get pvc -nminio-2
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
minio-2   Bound    pvc-3d4c1846-fc64-4af0-8104-64fc66c1c1bb   2Gi        RWO            hwameistor-storage-lvm-hdd   103m
```

### Distributed mode deployment

```console
$ helm install minio-1 \
  --namespace minio --create-namespace \
  --set accessKey=admin,secretKey=admin123 \
  --set mode=distributed \
  --set replicas=4 \
  --set service.type=NodePort \
  --set persistence.size=2Gi \
  --set persistence.storageClass=hwameistor-storage-lvm-hdd \
  minio/minio

$ kubectl get all -nminio
NAME            READY   STATUS    RESTARTS   AGE
pod/minio-1-0   1/1     Running   0          13h
pod/minio-1-1   1/1     Running   0          13h
pod/minio-1-2   1/1     Running   0          13h
pod/minio-1-3   1/1     Running   0          13h

NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/minio-1       NodePort    10.108.88.252   <none>        9000:32001/TCP   13h
service/minio-1-svc   ClusterIP   None            <none>        9000/TCP         13h

NAME                       READY   AGE
statefulset.apps/minio-1   4/4     13h
```

View PVCs on HwameiStor

```console
$ kubectl get pvc -nminio
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE
export-minio-1-0   Bound    pvc-bfbff95b-1afc-4484-8039-6ae402fd9116   2Gi        RWO            hwameistor-storage-lvm-hdd   13h
export-minio-1-1   Bound    pvc-fc178030-1dde-4d14-90db-796078041ae2   2Gi        RWO            hwameistor-storage-lvm-hdd   13h
export-minio-1-2   Bound    pvc-527cc3af-7fa4-4496-b4fc-08d69166d582   2Gi        RWO            hwameistor-storage-lvm-hdd   13h
export-minio-1-3   Bound    pvc-29ff7a1c-5097-4e84-ac04-961eb735ddec   2Gi        RWO            hwameistor-storage-lvm-hdd   13h
```
