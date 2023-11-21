---
sidebar_position: 8
sidebar_label:  "Local Volumes"
---

# Local Volumes

Running a stateful application with HwameiStor is super easy.

As an example, we will deploy a MySQL application by creating a local volume.

:::note
The yaml file for MySQL is learnt from
[Kubernetes repo](https://github.com/kubernetes/website/blob/main/content/en/examples/application/mysql/mysql-statefulset.yaml)
:::

## Verify `StorageClass`

Make sure the StorageClasses have been created successfully by HwameiStor Operator. And then select one of them to provision the data volume for the application.

```console
$ kubectl get sc hwameistor-storage-lvm-hdd -o yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hwameistor-storage-lvm-hdd
parameters:
  convertible: "false"
  csi.storage.k8s.io/fstype: xfs
  poolClass: HDD
  poolType: REGULAR
  replicaNumber: "1"
  striped: "true"
  volumeKind: LVM
provisioner: lvm.hwameistor.io
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

## Create `StatefulSet`

With HwameiStor and its `StorageClass` ready, a MySQL StatefulSet and its volumes
can be deployed by a single command:

```Console
$ kubectl apply -f sts-mysql_local.yaml
```

Please note the `volumeClaimTemplates` uses `storageClassName: hwameistor-storage-lvm-hdd`:

```yaml
spec:
  volumeClaimTemplates:
  - metadata:
      name: data
      labels:
        app: sts-mysql-local
        app.kubernetes.io/name: sts-mysql-local
    spec:
      storageClassName: hwameistor-storage-lvm-hdd
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
```

Please note the minimum PVC size need to be over 4096 blocks, for example, 16MB with 4KB block.

## Verify MySQL Pod and `PVC/PV`

In this example, the pod is scheduled on node `k8s-worker-3`.

```console
$ kubectl get po -l  app=sts-mysql-local -o wide
NAME                READY   STATUS    RESTARTS   AGE     IP            NODE        
sts-mysql-local-0   2/2     Running   0          3m08s   10.1.15.154   k8s-worker-3

$ kubectl get pvc -l  app=sts-mysql-local
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE   VOLUMEMODE
data-sts-mysql-local-0   Bound    pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   1Gi        RWO            hwameistor-storage-lvm-hdd    3m   Filesystem
```

## Verify `LocalVolume` objects

By listing `LocalVolume(LV)` objects with the same name as that of the `PV`,
we can see that the local volume is also created on node `k8s-worker-3`

```console
$ kubectl get lv pvc-accf1ddd-6f47-4275-b520-dc317c90f80b

NAME                                       POOL                   REPLICAS   CAPACITY     ACCESSIBILITY   STATE   RESOURCE   PUBLISHED      AGE
pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-3    3m
```

## [Optional] Scale MySQL out into a 3-node Cluster

HwameiStor supports `StatefulSet` scaleout. Each `pod` of the `StatefulSet` will
attach and mount an independent HwameiStor volume.

```console
$ kubectl scale sts/sts-mysql-local --replicas=3

$ kubectl get po -l  app=sts-mysql-local -o wide
NAME                READY   STATUS     RESTARTS   AGE     IP            NODE        
sts-mysql-local-0   2/2     Running    0          4h38m   10.1.15.154   k8s-worker-3
sts-mysql-local-1   2/2     Running    0          19m     10.1.57.44    k8s-worker-2
sts-mysql-local-2   0/2     Init:0/2   0          14s     10.1.42.237   k8s-worker-1

$ kubectl get pvc -l  app=sts-mysql-local -o wide
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE     VOLUMEMODE
data-sts-mysql-local-0   Bound    pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   1Gi        RWO            hwameistor-storage-lvm-hdd   3m07s   Filesystem
data-sts-mysql-local-1   Bound    pvc-a4f8b067-9c1d-450f-aff4-5807d61f5d88   1Gi        RWO            hwameistor-storage-lvm-hdd   2m18s   Filesystem
data-sts-mysql-local-2   Bound    pvc-47ee308d-77da-40ec-b06e-4f51499520c1   1Gi        RWO            hwameistor-storage-lvm-hdd   2m18s   Filesystem

$ kubectl get lv
NAME                                       POOL                   REPLICAS   CAPACITY     ACCESSIBILITY   STATE   RESOURCE   PUBLISHED      AGE
pvc-47ee308d-77da-40ec-b06e-4f51499520c1   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-1   2m50s
pvc-a4f8b067-9c1d-450f-aff4-5807d61f5d88   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-2   2m50s
pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-3   3m40s
```
