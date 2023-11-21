---
sidebar_position: 8
sidebar_label:  "本地卷"
---

# 本地卷

使用 HwameiStor 能非常轻松的运行有状态的应用。

我们通过创建本地卷来部署一个 MySQL 应用作为例子。

:::note
下面的 MySQL Yaml 文件来自于 [Kubernetes 的官方 Repo](https://github.com/kubernetes/website/blob/main/content/en/examples/application/mysql/mysql-statefulset.yaml)
:::

## 查看 `StorageClass`

首先确认 HwameiStor Operator 创建了 StorageClass。然后从中选一个合适的用于创建单副本数据卷。

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

如果这个 `storageClass` 没有在安装时生成，可以运行以下的 yaml 文件重新生成它：

```console
$ kubectl apply -f examples/sc-local.yaml
```

## 创建 `StatefulSet`

在 HwameiStor 和 `StorageClass` 就绪后, 一条命令就能创建 MySQL 容器和它的数据卷:

```Console
$ kubectl apply -f sts-mysql_local.yaml
```

请注意 `volumeClaimTemplates` 使用 `storageClassName: hwameistor-storage-lvm-hdd`:

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

请注意，PVC 容量的最小值需要超过 4096 个块，例如使用 4KB 块时为 16MB。

## 查看 MySQL 容器和 `PVC/PV`

在这个例子里，MySQL 容器被调度到了节点 `k8s-worker-3`。

```console
$ kubectl get po -l  app=sts-mysql-local -o wide
NAME                READY   STATUS    RESTARTS   AGE     IP            NODE        
sts-mysql-local-0   2/2     Running   0          3m08s   10.1.15.154   k8s-worker-3

$ kubectl get pvc -l  app=sts-mysql-local
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                 AGE   VOLUMEMODE
data-sts-mysql-local-0   Bound    pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   1Gi        RWO            hwameistor-storage-lvm-hdd    3m   Filesystem
```

## 查看 `LocalVolume` 对象

通过查看和 `PV` 同名的 `LocalVolume(LV)`, 可以看到本地卷创建在了节点 `k8s-worker-3`上：

```console
$ kubectl get lv pvc-accf1ddd-6f47-4275-b520-dc317c90f80b

NAME                                       POOL                   REPLICAS   CAPACITY     ACCESSIBILITY   STATE   RESOURCE   PUBLISHED      AGE
pvc-accf1ddd-6f47-4275-b520-dc317c90f80b   LocalStorage_PoolHDD   1          1073741824                   Ready   -1         k8s-worker-3    3m
```

## [可选] 扩展 MySQL 应用成一个三节点的集群

HwameiStor 支持 `StatefulSet` 的横向扩展. `StatefulSet`容器都会挂载一个独立的本地卷：

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
