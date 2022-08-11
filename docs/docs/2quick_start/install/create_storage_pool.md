---
sidebar_position: 4
sidebar_label: "Set up a Storage Pool"
---

# Step up a Storage Pool

## Step 1: Create LocalDiskClaim objects

HwameiStor sets up storage pools by creating `LocalDiskClaim` objects according to the storage media types. To create an HDD pool on all kubernetes worker nodes:

```bash
$ helm template helm/hwameistor \
        -s templates/post-install-claim-disks.yaml \
        --set storageNodes='{k8s-worker-1,k8s-worker-2,k8s-worker-3}' \
        | kubectl apply -f -
```

## Step 2: Verify LocalDiskClaim objects

```bash
$ kubectl get ldc
NAME           NODEMATCH      PHASE
k8s-worker-1   k8s-worker-1   Bound
k8s-worker-2   k8s-worker-2   Bound
k8s-worker-3   k8s-worker-3   Bound
```

## Step 3: Verify StorageClass

```bash
$  kubectl get sc hwameistor-storage-lvm-hdd
NAME                         PROVISIONER         RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
hwameistor-storage-lvm-hdd   lvm.hwameistor.io   Delete          WaitForFirstConsumer   true                   114s
```

## Step 4: Verify LocalDisk objects

```bash
$ kubectl get ld
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

## Step 5 (Optional): Observe VG

On a kubernetes worker node, observe a `VG` is created for an `LocalDiskClaim` object

```bash
root@k8s-worker-1:~$ vgdisplay LocalStorage_PoolHDD
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

## Set up storage pool during deployment

A storage pool can be configured during HwameiStor deployment by helm command:

```bash
$ helm install \
    --namespace hwameistor \
    --create-namespace \
    hwameistor \
    helm/hwameistor \
    --set storageNodes='{k8s-worker-1,k8s-worker-2,k8s-worker-3}'
```
