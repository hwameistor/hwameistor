---
sidebar_position: 4
sidebar_label: "Volume Provisioned IO"
---

# Volume Provisioned IO

In HwameiStor, it allows users to specify the maximum IOPS and throughput of a volume on a Kuberentes cluster.

Please follow the steps below to create a volume with the maximum IOPS and throughput and create a workload to use it.

## Requirements (if you want to limit non-direct io)

cgroup v2 has the following requirements:

- OS distribution enables cgroup v2
- Linux Kernel version is 5.8 or later

More info, please refer to the [Kubernetes website](https://kubernetes.io/docs/concepts/architecture/cgroups)

## Create a new StorageClass with the maximum IOPS and throughput parameters

By default, HwameiStor won't auto-create such a StorageClass during the installation, so you need to create it manually.

A sample StorageClass is as follows:

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

Compare to the regular StorageClass created by HwameiStor installer, the following parameters are added:

- provision-iops-on-creation: It specifies the maximum IOPS of the volume on creation.
- provision-throughput-on-creation: It specifies the maximum throughput of the volume on creation.

After the StorageClass is created, you can use it to create a PVC.

## Create a PVC with the StorageClass

A sample PVC is as follows:

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

After the PVC is created, you can create a deployment to use it.

## Create a Deployment with the PVC

A sample Deployment is as follows:

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

After the Deployment is created, you can test the volume's IOPS and throughput by using the following command:

shell 1:

```bash
kubectl exec -it pod-sample-5f5f8f6f6f-5q4q5 -- /bin/sh
dd if=/dev/zero of=/data/test bs=4k count=1000000 oflag=direct
```

shell 2:

`/dev/LocalStorage_PoolHDD/pvc-c623054b-e7e9-41d7-a987-77acd8727e66` is the path of the volume on the node. you can find it by using the `kubectl get lvr` command.

```bash
iostat -d /dev/LocalStorage_PoolHDD/pvc-c623054b-e7e9-41d7-a987-77acd8727e66  -x -k 2
```

:::note
Due to the cgroupv1 limitation, the settings of the maximum IOPS and throughput may not take effect
on non-direct IO. However, it will take effect on non-direct IO in cgroupv2.
:::

## How to change the maximum IOPS and throughput of a volume

The maximum IOPS and throughput are specified on the parameters of the StorageClass,
you can not change it directly because it is immutable today.

Different from the other storage vendors, HwameiStor is a Native Kubernetes storage solution
and it defines a set of operation primitives based on the Kubernetes CRDs. It means that you
can modify the related CRD to change the actual maximum IOPS and throughput of a volume.

The following steps show how to change the maximum IOPS and throughput of a volume.

### Find the corresponding LocalVolume CR for the given PVC

```console
$ kubectl get pvc pvc-sample

NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                        AGE
demo             Bound     pvc-c354a56a-5cf4-4ff6-9472-4e24c7371e10   10Gi       RWO            hwameistor-storage-lvm-hdd          5d23h
pvc-sample       Bound     pvc-cac82087-6f6c-493a-afcd-09480de712ed   10Gi       RWO            hwameistor-storage-lvm-hdd-sample   5d23h


$ kubectl get localvolume

NAME                                       POOL                   REPLICAS   CAPACITY      USED       STATE   RESOURCE   PUBLISHED   FSTYPE   AGE
pvc-c354a56a-5cf4-4ff6-9472-4e24c7371e10   LocalStorage_PoolHDD   1          10737418240   33783808   Ready   -1         master      xfs      5d23h
pvc-cac82087-6f6c-493a-afcd-09480de712ed   LocalStorage_PoolHDD   1          10737418240   33783808   Ready   -1         master      xfs      5d23h
```

According to the print out, the LocalVolume CR for the PVC is `pvc-cac82087-6f6c-493a-afcd-09480de712ed`.

### Modify the LocalVolume CR

```bash
kubectl edit localvolume pvc-cac82087-6f6c-493a-afcd-09480de712ed
```

In the editor, find the `spec.volumeQoS` section and modify the `iops` and `throughput` fields. By the way, an empty value means no limit.

At last, save the changes and exit the editor. The settings will take effect in a few seconds.

:::note
In the future, we will allow users to modify the maximum IOPS and throughput of a volume directly
once the Kubernetes supports [it](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/3751-volume-attributes-class#motivation).
:::

## How to check the actual IOPS and throughput of a volume

HwameiStor uses the [cgroupv1](https://www.kernel.org/doc/Documentation/cgroup-v1/blkio-controller.txt)
or [cgroupv2](https://www.kernel.org/doc/Documentation/cgroup-v2.txt) to limit the IOPS and throughput
of a volume, so you can use the following command to check the actual IOPS and throughput of a volume.

```console
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

# if cgroup version is v1.

$ cat /sys/fs/cgroup/blkio/blkio.throttle.read_iops_device
253:3 100

$ cat /sys/fs/cgroup/blkio/blkio.throttle.write_iops_device
253:3 100

$ cat /sys/fs/cgroup/blkio/blkio.throttle.read_bps_device
253:3 1048576

$ cat /sys/fs/cgroup/blkio/blkio.throttle.write_bps_device
253:3 1048576

# if cgroup version is v2.

# cat /sys/fs/cgroup/kubepods.slice/io.max
253:0 rbps=1048576 wbps=1048576 riops=100 wiops=100
```
