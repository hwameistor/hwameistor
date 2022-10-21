---
sidebar_position: 2
sidebar_label:  "Migrate Volumes"
---

# Migrate Volumes

The `Migrate` function is an important operation and maintenance management function in HwameiStor. When the copy of the node where the data volume bound to the application is located is damaged, the copy of the volume can be migrated to other nodes, and after successfully migrated to the new node, the application can be rescheduled to the new node and bind mount the data volume.

## Basic Concepts

`LocalVolumeGroup(LVG)` management is an important function in HwameiStor. When an application Pod applies for multiple data volume PVCs, in order to ensure the correct operation of the Pod, these data volumes must have certain attributes, such as the number of copies of the data volume and the node where the copies are located. Properly managing these associated data volumes through the data volume group management function is a very important capability in HwameiStor.

## Preconditions

`LocalVolumeMigrate` needs to be deployed in the Kubernetes system, and the deployed application needs to meet the following conditions:

* Support `lvm` type volumes
  * When migrating based on `LocalVolume` granularity, the data volumes belonging to the same `LocalVolumeGroup` by default will not be migrated together (if they are migrated together, you need to configure the switch `MigrateAllVols: true`)

## Step 1: Create convertible `StorageClass`

```console
$ cd ../../deploy/
$ kubectl apply -f storageclass-convertible-lvm.yaml
```

## Step 2: Create multiple `PVCs`

```console
$ kubectl apply -f pvc-multiple-lvm.yaml
```

## Step 3: Deploy multi-volume pod

```console
$ kubectl apply -f nginx-multiple-lvm.yaml
```

## Step 4: Detach multi-volume pod

```console
$ kubectl -n hwameistor scale --current-replicas=1 --replicas=0 deployment/nginx-local-storage-lvm
```

## Step 5: Create migration tasks

```console
cat << EOF | kubectl apply -f -
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeMigrate
metadata:
  namespace: hwameistor
  name: <localVolumeMigrateName>
spec:
  sourceNode: <sourceNodeName>
  targetNodesSuggested: 
  - <targetNodesName1>
  - <targetNodesName2>
  volumeName: <volName>
  migrateAllVols: <true/false>
EOF
```

Attentions:

1) HwameiStor will select a target node from targetNodesSuggested to migrate. If all the candidates don't have enough storage space, the migrate will fail.

2) If targetNodesSuggested is emtpy or not set, HwameiStore will automatically select a propriate node for the migrate. If there is no valid candidate, the migrate will fail.

## Step 6: Check migration Status

```console
$ kubectl  get LocalVolumeMigrate localvolumemigrate-1 -o yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeMigrate
metadata:
  generation: 1
  name: localvolumemigrate-1
  namespace: hwameistor
  resourceVersion: "12828637"
  uid: 78af7f1b-d701-4b03-84de-27fafca58764
spec:
  abort: false
  migrateAllVols: true
  sourceNode: k8s-172-30-40-61
  targetNodesSuggested:
  - k8s-172-30-45-223
  volumeName: pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4
status:
  originalReplicaNumber: 1
  targetNode: k8s-172-30-45-223
  state: Completed
  message: 
```

## Step 7: Verify migration results

```console
[root@172-30-45-222 deploy]# kubectl  get lvr
NAME                                              CAPACITY     NODE                STATE   SYNCED   DEVICE                                                                  AGE
pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4-9cdkkn   1073741824   k8s-172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4   77s
pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a-7ppmrx   1073741824   k8s-172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a   77s
```

## Step 8: Reattach/Remount volume

```console
$ kubectl -n hwameistor scale --current-replicas=0 --replicas=1 deployment/nginx-local-storage-lvm
```
