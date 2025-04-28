---
sidebar_position: 11
sidebar_label: "Thin Provision Volumes"
---

# Hwameistor Thin Provision User Guide

## 1. Overview

Hwameistor now supports Thin Provision functionality, implemented based on LVM's thin provisioning feature. Compared to traditional thick provisioning, thin mode enables more efficient storage space utilization and supports rapid snapshot creation and cloning.

## 2. Use Cases

**Recommended scenarios for thin provisioning:**
- Frequent snapshot or volume clone creation required
- Limited storage space needing over-provisioning
- Applications without extreme storage performance requirements
- Single-replica scenarios (current version doesn't support thin multi-replica)

**Not recommended scenarios:**
- Performance-critical applications (thin provisioning has some overhead)
- High-availability scenarios requiring multiple replicas (current version limitation)

## 3. Quick Start

### 3.1 Create ThinPoolClaim

First create a ThinPoolClaim:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: ThinPoolClaim
metadata:
  name: example-thinpool
spec:
  nodeName: node1  # Specify node
  description:
    poolName: LocalStorage_PoolHDD  # Specify storage pool for thin pool creation. Options: LocalStorage_PoolHDD, LocalStorage_PoolSSD, LocalStorage_PoolNVMe
    capacity: 100  # ThinPool capacity in GiB
    overProvisionRatio: "1.0"  # Over-provisioning ratio, default and minimum is 1.0. Example: If ratio is "3.0" with 100GiB capacity, the pool can over-provision up to 300GiB
    poolMetadataSize: 1  # Metadata pool size in GiB. Default 1G sufficient for most scenarios
```

### 3.2 Create Thin StorageClass

```yaml
allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: hwameistor-storage-lvm-thin-hdd
parameters:
  convertible: "false"
  csi.storage.k8s.io/fstype: ext4
  poolClass: HDD
  poolType: REGULAR
  # Currently only "1" is supported
  replicaNumber: "1"
  striped: "true"
  # This is used to specify the SC to create thin PVC
  thin: "true"
  volumeKind: LVM
provisioner: lvm.hwameistor.io
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

## 4. Usage Examples

### 4.1 PVC Usage

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
  storageClassName: hwameistor-storage-lvm-thin-hdd
---
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: test-pvc
  terminationGracePeriodSeconds: 10
```

### 4.2 Snapshot

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-snapshot
spec:
  volumeSnapshotClassName: hwameistor-storage-lvm-snapshot
  source:
    persistentVolumeClaimName: test-pvc
```

### 4.3 Create PVC from Snapshot

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc2
spec:
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
  storageClassName: hwameistor-storage-lvm-thin-hdd
  dataSource:
    name: my-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
---
apiVersion: v1
kind: Pod
metadata:
  name: test-pod2
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: test-pvc2
```

### 4.4 Clone Operation

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloned-pvc
spec:
  storageClassName: hwameistor-storage-lvm-thin-hdd
  dataSource:
    name: test-pvc
    kind: PersistentVolumeClaim
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: cloned-pod
spec:
  containers:
  - name: busybox
    image: busybox:1.31.1
    command:
      - sleep
      - "360000000"
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: temp-pvc
      mountPath: /mnt/temp-fs
  volumes:
  - name: temp-pvc
    persistentVolumeClaim:
      claimName: cloned-pvc
  terminationGracePeriodSeconds: 10
```

## 5. Monitoring and Management

### 5.1 Check Thin Pool Status

```bash
kubectl get localstoragenodes node-name -o yaml
```

Key fields to monitor:
- `status.pools.<pool-name>.thinPool`: Contains thin pool details
- `status.pools.<pool-name>.thinPoolExtendRecords`: Records thin pool extension history

### 5.2 Expand Thin Pool

When thin pool usage approaches limit, create another ThinPoolClaim to expand. Both `spec.description.capacity` and `spec.description.poolMetadataSize` can be increased, while `spec.description.overProvisionRatio` can be adjusted as needed.

## 6. Important Notes

1. **Over-provisioning Risk**: While thin provisioning supports over-allocation, exceeding physical capacity causes serious issues. **Closely monitor thin pool usage (dataPercent, metadataPercent in `status.pools.<pool-name>.thinPool`) to prevent full capacity situations**
2. **Performance Impact**: Thin volumes have some performance overhead - evaluate carefully for performance-sensitive applications
3. **Version Compatibility**: Thin and thick volumes cannot be converted between each other
4. **Replica Limitation**: Current version only supports single-replica thin volumes

## 7. Troubleshooting

If thin pool nears full capacity:
1. Immediately stop creating new thin volumes
2. Delete unnecessary snapshots and clones
3. Expand thin pool capacity
4. If already full, refer to [LVM documentation](https://man7.org/linux/man-pages/man7/lvmthin.7.html) for recovery
