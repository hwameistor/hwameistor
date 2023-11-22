---
sidebar_position: 2
sidebar_label:  "Use HA Volumes"
---

# Use HA Volumes

When the HA module is enabled, HwameiStor Operator will generate a StorageClass of HA automatically.

As an example, we will deploy a MySQL application by creating a highly available (HA) volume.

:::note
The yaml file for MySQL is learnt from
[Kubernetes repo](https://github.com/kubernetes/website/blob/main/content/en/examples/application/mysql/mysql-statefulset.yaml)
:::

## Verify `StorageClass`

`StorageClass` "hwameistor-storage-lvm-hdd-ha" has parameter `replicaNumber: "2"`,
which indicates a DRBD replication pair.

```console
$ kubectl apply -f examples/sc_ha.yaml

$ kubectl get sc hwameistor-storage-lvm-hdd-ha -o yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hwameistor-storage-lvm-hdd-ha
parameters:
  replicaNumber: "2"
  convertible: "false"
  csi.storage.k8s.io/fstype: xfs
  poolClass: HDD
  poolType: REGULAR
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
$ kubectl apply -f exapmles/sts-mysql_ha.yaml
```

Please note the `volumeClaimTemplates` uses `storageClassName: hwameistor-storage-lvm-hdd-ha`:

```yaml
spec:
  volumeClaimTemplates:
  - metadata:
      name: data
      labels:
        app: sts-mysql-ha
        app.kubernetes.io/name: sts-mysql-ha
    spec:
      storageClassName: hwameistor-storage-lvm-hdd-ha
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
```

## Verify MySQL Pod and `PVC/PV`

In this example, the pod is scheduled on node `k8s-worker-3`.

```console
$ kubectl get po -l  app=sts-mysql-ha -o wide
NAME                READY   STATUS    RESTARTS   AGE     IP            NODE        
sts-mysql-ha-0   2/2     Running   0          3m08s   10.1.15.151   k8s-worker-1

$ kubectl get pvc -l  app=sts-mysql-ha
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE   VOLUMEMODE
data-sts-mysql-ha-0   Bound    pvc-5236ee6f-8212-4628-9876-1b620a4c4c36   1Gi        RWO            hwameistor-storage-lvm-hdd    3m   Filesystem
```

## Verify `LocalVolume` and `LocalVolumeReplica` objects

By listing `LocalVolume(LV)` objects with the same name as that of the `PV`,
we can see that the `LV` object is created on two nodes: `k8s-worker-1` and `k8s-worker-2`.

```console
$ kubectl get lv pvc-5236ee6f-8212-4628-9876-1b620a4c4c36

NAME                                       POOL                   REPLICAS   CAPACITY     ACCESSIBILITY   STATE   RESOURCE   PUBLISHED                    AGE
pvc-5236ee6f-8212-4628-9876-1b620a4c4c36   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-1,k8s-worker-2    3m
```

`LocalVolumeReplica (LVR)` further shows the backend logical volume devices on each node.

```concole
$ kubectl get lvr
NAME                                          CAPACITY     NODE           STATE   SYNCED   DEVICE                                                              AGE
5236ee6f-8212-4628-9876-1b620a4c4c36-d2kn55   1073741824   k8s-worker-1   Ready   true     /dev/LocalStorage_PoolHDD-HA/5236ee6f-8212-4628-9876-1b620a4c4c36   4m
5236ee6f-8212-4628-9876-1b620a4c4c36-glm7rf   1073741824   k8s-worker-3   Ready   true     /dev/LocalStorage_PoolHDD-HA/5236ee6f-8212-4628-9876-1b620a4c4c36   4m
```
