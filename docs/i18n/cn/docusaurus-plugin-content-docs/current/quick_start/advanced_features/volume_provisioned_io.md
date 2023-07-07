---
sidebar_position: 4
sidebar_label: "数据卷IO限速"
---

# 数据卷IO限速

在 Hwameistor 中，它允许用户指定 Kuberentes 集群上卷的最大 IOPS 和吞吐量。

请按照以下步骤创建具有最大 IOPS 和吞吐量的卷并创建工作负载来使用它。

## 使用最大 IOPS 和吞吐量参数创建新的 StorageClass

默认情况下，Hwameistor在安装过程中不会自动创建这样的StorageClass，因此您需要手动创建它。

示例 StorageClass 如下：

```yaml
allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hwameistor-storage-lvm-hdd-sample
parameters:
  convertible: "false"
  csi.storage.k8s.io/fstype: xfs
  poolClass: HDD
  poolType: REGULAR
  provision-iops-on-creation: "100"
  provision-throughput-on-creation: 1Mi
  replicaNumber: "1"
  striped: "true"
  volumeKind: LVM
provisioner: lvm.hwameistor.io
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

与 Hwameistor 安装程序创建的常规 StorageClass 相比，添加了以下参数：

- Provision-iops-on-creation：指定创建时卷的最大 IOPS。
- Provision-throughput-on-creation：它指定创建时卷的最大吞吐量。

创建StorageClass后，您可以使用它来创建PVC。

## 使用 StorageClass 创建 PVC

示例 PVC 如下：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-sample
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: hwameistor-storage-lvm-hdd-sample
```

创建 PVC 后，您可以创建 Deployment 来使用它。

## 创建带有 PVC 的 Deployment

示例 Deployment 如下：:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: pod-sample
  name: pod-sample
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pod-sample
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: pod-sample
    spec:
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: pvc-sample
      containers:
      - command:
        - sleep
        - "100000"
        image: busybox
        name: busybox
        resources: {}
        volumeMounts:
        - name: data
          mountPath: /data
status: {}
```

创建 Deployment 后，您可以使用以下命令测试卷的 IOPS 和吞吐量：

```bash
$ kubectl exec -it pod-sample-5f5f8f6f6f-5q4q5 -- /bin/sh
$ dd if=/dev/zero of=/data/test bs=4k count=1000000 oflag=direct
```

**注意**：由于cgroupv1限制，最大IOPS和吞吐量的设置可能对非直接IO不生效。

## 如何更改数据卷的最大 IOPS 和吞吐量

最大 IOPS 和吞吐量在 StorageClass 的参数上指定，您不能直接更改它，因为它现在是不可变的。

与其他存储厂商不同的是，Hwameistor是一个基于 Kubernetes 的存储解决方案，它定义了一组操作原语
基于 Kubernetes CRD。 这意味着您可以修改相关的CRD来更改卷的实际最大IOPS和吞吐量。

以下步骤显示如何更改数据卷的最大 IOPS 和吞吐量。

### 查找给定 PVC 对应的 LocalVolume CR

```
$ kubectl get pvc pvc-sample

NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                        AGE
demo             Bound     pvc-c354a56a-5cf4-4ff6-9472-4e24c7371e10   10Gi       RWO            hwameistor-storage-lvm-hdd          5d23h
pvc-sample       Bound     pvc-cac82087-6f6c-493a-afcd-09480de712ed   10Gi       RWO            hwameistor-storage-lvm-hdd-sample   5d23h


$ kubectl get localvolume

NAME                                       POOL                   REPLICAS   CAPACITY      USED       STATE   RESOURCE   PUBLISHED   FSTYPE   AGE
pvc-c354a56a-5cf4-4ff6-9472-4e24c7371e10   LocalStorage_PoolHDD   1          10737418240   33783808   Ready   -1         master      xfs      5d23h
pvc-cac82087-6f6c-493a-afcd-09480de712ed   LocalStorage_PoolHDD   1          10737418240   33783808   Ready   -1         master      xfs      5d23h
```

根据打印输出，PVC 的 LocalVolume CR 为 `pvc-cac82087-6f6c-493a-afcd-09480de712ed`。

### 修改LocalVolume CR

```
$ kubectl edit localvolume pvc-cac82087-6f6c-493a-afcd-09480de712ed
```

在编辑器中，找到 `spec.volumeQoS` 部分并修改 `iops` 和 `throughput` 字段。 顺便说一下，空值意味着没有限制。

最后，保存更改并退出编辑器。 设置将在几秒钟后生效。

**注意**：将来，一旦 Kubernetes 支持[它](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/3751-volume-attributes-class#motivation)，我们将允许用户直接修改卷的最大 IOPS 和吞吐量。

## 如何检查数据卷的实际 IOPS 和吞吐量

Hwameistor 使用 [cgroupv1](https://www.kernel.org/doc/Documentation/cgroup-v1/blkio-controller.txt) 来限制数据卷的 IOPS 和吞吐量，因此您可以使用以下命令来检查数据卷的实际 IOPS 和吞吐量。

```
$ lsblk
NAME            MAJ:MIN RM   SIZE RO TYPE MOUNTPOINT
sda               8:0    0   160G  0 disk
├─sda1            8:1    0     1G  0 part /boot
└─sda2            8:2    0   159G  0 part
  ├─centos-root 253:0    0   300G  0 lvm  /
  ├─centos-swap 253:1    0   7.9G  0 lvm
  └─centos-home 253:2    0 101.1G  0 lvm  /home
sdb               8:16   0   100G  0 disk
├─LocalStorage_PoolHDD-pvc--cac82087--6f6c--493a--afcd--09480de712ed
                253:3    0    10G  0 lvm  /var/lib/kubelet/pods/3d6bc980-68ae-4a65-a1c8-8b410b7d240f/v
└─LocalStorage_PoolHDD-pvc--c354a56a--5cf4--4ff6--9472--4e24c7371e10
                253:4    0    10G  0 lvm  /var/lib/kubelet/pods/521fd7b4-3bef-415b-8720-09225f93f231/v
sdc               8:32   0   300G  0 disk
└─sdc1            8:33   0   300G  0 part
  └─centos-root 253:0    0   300G  0 lvm  /
sr0              11:0    1   973M  0 rom

$ cat /sys/fs/cgroup/blkio/blkio.throttle.read_iops_device
253:3 100

$ cat /sys/fs/cgroup/blkio/blkio.throttle.write_iops_device
253:3 100

$ cat /sys/fs/cgroup/blkio/blkio.throttle.read_bps_device
253:3 1048576

$ cat /sys/fs/cgroup/blkio/blkio.throttle.write_bps_device
253:3 1048576
```