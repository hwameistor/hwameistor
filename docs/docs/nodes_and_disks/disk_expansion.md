---
sidebar_position: 4
sidebar_label: "Disk Expansion"
---

# Disk Expansion

A storage system is usually expected to expand its capacity by adding a new disk
into a storage node. In HwameiStor, it can be done with the following steps.

## Prepare a new storage disk

Select a storage node from the HwameiStor system, and add a new disk into it.

For example, the storage node and new disk information are as follows:

- name: k8s-worker-4
- devPath: /dev/sdc
- diskType: SSD

After the new disk is added into the storage node `k8s-worker-4`, you can check the disk status as below.

1. Check if the new disk is added into the node successfully

    ```bash
    ssh root@k8s-worker-4
    lsblk | grep sdc
    ```

    The output may look like:

    ```none
    sdc        8:32     0     20G  1 disk
    ```

2. Check if the LocalDisk CR already exists for the new disk and the status is "Unclaimed"

    ```bash
    kubectl get localdisk | grep k8s-worker-4 | grep sdc
    ```

    The output may look like:

    ```none
    k8s-worker-4-sdc   k8s-worker-4       Available 
    ```

## Add the new disk into the node's storage pool

The new disk should be added into the existing SSD storage pool of the node.
If the storage pool doesn't exist, it will be constructed automatically and the new disk should be added into it.

```console
$ kubectl apply -f - <<EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: k8s-worker-4-expand
spec:
  nodeName: k8s-worker-4
  owner: local-storage
  description:
    diskType: SSD
EOF
```

## Post check

Finally, check if the new disk has been added into the node's storage pool successfully by checking the LocalStorageNode CR:

```bash
kubectl get localstoragenode k8s-worker-4
```

The output may look like:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  name: k8s-worker-4
spec:
  hostname: k8s-worker-4
  storageIP: 10.6.182.103
  topogoly:
    region: default
    zone: default
status:
  pools:
    LocalStorage_PoolSSD:
      class: SSD
      disks:
      - capacityBytes: 214744170496
        devPath: /dev/sdb
        state: InUse
        type: SSD
      - capacityBytes: 214744170496
        devPath: /dev/sdc
        state: InUse
        type: SSD
      freeCapacityBytes: 429488340992
      freeVolumeCount: 1000
      name: LocalStorage_PoolSSD
      totalCapacityBytes: 429488340992
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 0
      usedVolumeCount: 0
      volumeCapacityBytesLimit: 429488340992
      volumes:
  state: Ready
```
