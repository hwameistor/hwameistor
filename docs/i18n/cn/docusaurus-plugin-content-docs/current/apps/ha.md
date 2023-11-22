---
sidebar_position: 3
sidebar_label:  "使用高可用卷"
---

# 使用高可用卷

当 HwameiStor 的 HA 模块被开启后，HwameiStor Operator 会自动创建一个 HA 的 StorageClass 用于创建 HA 数据卷。

我们通过创建高可用 HA 卷来部署一个 MySQL 应用作为例子。

:::note
下面的 MySQL Yaml 文件来自于 [Kubernetes 的官方 Repo](https://github.com/kubernetes/website/blob/main/content/en/examples/application/mysql/mysql-statefulset.yaml)
:::

## 查看 `StorageClass`

`StorageClass` "hwameistor-storage-lvm-hdd-ha" 使用参数 `replicaNumber: "2"` 开启高可用功能：

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

## 创建 `StatefulSet`

在 HwameiStor 和 `StorageClass` 就绪后, 一条命令就能创建 MySQL 容器和它的数据卷：

```Console
$ kubectl apply -f exapmles/sts-mysql_ha.yaml
```

请注意 `volumeClaimTemplates` 使用 `storageClassName: hwameistor-storage-lvm-hdd-ha`：

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

## 查看 MySQL Pod 和 `PVC/PV`

在这个例子里，MySQL 容器被调度到了节点 `k8s-worker-3`。

```console
$ kubectl get po -l  app=sts-mysql-ha -o wide
NAME                READY   STATUS    RESTARTS   AGE     IP            NODE        
sts-mysql-ha-0   2/2     Running   0          3m08s   10.1.15.151   k8s-worker-1

$ kubectl get pvc -l  app=sts-mysql-ha
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE   VOLUMEMODE
data-sts-mysql-ha-0   Bound    pvc-5236ee6f-8212-4628-9876-1b620a4c4c36   1Gi        RWO            hwameistor-storage-lvm-hdd    3m   Filesystem
```

## 查看 `LocalVolume` 和 `LocalVolumeReplica` 对象

通过查看和 `PV` 同名的 `LocalVolume(LV)`, 可以看到本地卷创建在了节点 `k8s-worker-1` 和节点 `k8s-worker-2`。

```console
$ kubectl get lv pvc-5236ee6f-8212-4628-9876-1b620a4c4c36

NAME                                       POOL                   REPLICAS   CAPACITY     ACCESSIBILITY   STATE   RESOURCE   PUBLISHED                    AGE
pvc-5236ee6f-8212-4628-9876-1b620a4c4c36   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-1,k8s-worker-2    3m
```

`LocalVolumeReplica (LVR)` 进一步显示每个节点上的后端逻辑卷设备：

```console
$ kubectl get lvr
NAME                                          CAPACITY     NODE           STATE   SYNCED   DEVICE                                                              AGE
5236ee6f-8212-4628-9876-1b620a4c4c36-d2kn55   1073741824   k8s-worker-1   Ready   true     /dev/LocalStorage_PoolHDD-HA/5236ee6f-8212-4628-9876-1b620a4c4c36   4m
5236ee6f-8212-4628-9876-1b620a4c4c36-glm7rf   1073741824   k8s-worker-3   Ready   true     /dev/LocalStorage_PoolHDD-HA/5236ee6f-8212-4628-9876-1b620a4c4c36   4m
```
