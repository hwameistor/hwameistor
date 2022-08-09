---
sidebar_position: 4
sidebar_label: "创建存储池"
---

# 创建存储池

## 步骤 1：创建 LocalDiskClaim 对象

HwameiStor 根据存储介质类型创建 `LocalDiskClaim` 对象来创建存储池。
要在所有 Kubernetes Worker 节点上创建一个 HDD 存储池，运行以下命令：

```bash
helm template helm/hwameistor \
      -s templates/post-install-claim-disks.yaml \
      --set storageNodes='{k8s-worker-1,k8s-worker-2,k8s-worker-3}' \
      | kubectl apply -f -
```

## 步骤 2：验证 LocalDiskClaim 对象

运行以下命令：

```bash
kubectl get ldc
```

输出类似于：

```console
NAME           NODEMATCH      PHASE
k8s-worker-1   k8s-worker-1   Bound
k8s-worker-2   k8s-worker-2   Bound
k8s-worker-3   k8s-worker-3   Bound
```

## 步骤 3：验证 StorageClass

运行以下命令：

```bash
kubectl get sc hwameistor-storage-lvm-hdd
```

输出类似于：

```console
NAME                         PROVISIONER         RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-lvm-hdd   lvm.hwameistor.io   Delete          WaitForFirstConsumer   true                   114s
```

## 步骤 4：验证 LocalDisk 对象

运行以下命令：

```bash
kubectl get ld
```

输出类似于：

```console
NAME               NODEMATCH      CLAIM          PHASE
k8s-worker-1-sda   k8s-worker-1                  Inuse
k8s-worker-1-sdb   k8s-worker-1   k8s-worker-1   Claimed
k8s-worker-1-sdc   k8s-worker-1   k8s-worker-1   Claimed
k8s-worker-2-sda   k8s-worker-2                  Inuse
k8s-worker-2-sdb   k8s-worker-2   k8s-worker-2   Claimed
k8s-worker-2-sdc   k8s-worker-2   k8s-worker-2   Claimed
k8s-worker-3-sda   k8s-worker-3                  Inuse
k8s-worker-3-sdb   k8s-worker-3   k8s-worker-3   Claimed
k8s-worker-3-sdc   k8s-worker-3   k8s-worker-3   Claimed
```

## 步骤 5（可选）：观察 VG

在一个 Kubernetes Worker 节点上，观察为 `LocalDiskClaim` 对象创建 `VG`。

运行以下命令：

```bash
vgdisplay LocalStorage_PoolHDD
```

输出类似于：

```console
  --- Volume group ---
  VG Name               LocalStorage_PoolHDD
  System ID
  Format                lvm2
  Metadata Areas        2
  Metadata Sequence No  1
  VG Access             read/write
  VG Status             resizable
  MAX LV                0
  Cur LV                0
  Open LV               0
  Max PV                0
  Cur PV                2
  Act PV                2
  VG Size               199.99 GiB
  PE Size               4.00 MiB
  Total PE              51198
  Alloc PE / Size       0 / 0
  Free  PE / Size       51198 / 199.99 GiB
  VG UUID               jJ3s7g-iyoJ-c4zr-3Avc-3K4K-BrJb-A5A5Oe
```

## 安装期间配置存储池

通过 Helm 命令安装 HwameiStor 期间可以配置一个存储池：

```bash
helm install \
  --namespace hwameistor \
  --create-namespace \
  hwameistor \
  helm/hwameistor \
  --set storageNodes='{k8s-worker-1,k8s-worker-2,k8s-worker-3}'
```
