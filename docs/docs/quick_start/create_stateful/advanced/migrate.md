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
* convertible type volume (need to add the configuration item convertible: true in sc)
  * When applying pod to apply for multiple data volume `PVCs`, the corresponding data volume needs to use the same configuration sc
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
$ kubectl patch deployment nginx-local-storage-lvm --patch '{"spec": {"replicas": 0}}' -n hwameistor
```

## Step 5: Create migration tasks

```console
cat > ./migrate_lv.yaml <<- EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeMigrate
metadata:
  namespace: hwameistor
  name: <localVolumeMigrateName>
spec:
  targetNodesNames: 
  - <targetNodesName1>
  - <targetNodesName2>
  sourceNodesNames:
  - <sourceNodesName1>
  - <sourceNodesName2>
  volumeName: <volName>
  migrateAllVols: <true/false>
EOF
```

```console
$ kubectl apply -f ./migrate_lv.yaml
```

## Step 6: Check migration Status

```console
$ kubectl  get LocalVolumeMigrate  -o yaml
apiVersion: v1
items:
- apiVersion: hwameistor.io/v1alpha1
  kind: LocalVolumeMigrate
  metadata:
  annotations:
  kubectl.kubernetes.io/last-applied-configuration: |
  {"apiVersion":"hwameistor.io/v1alpha1","kind":"LocalVolumeMigrate","metadata":{"annotations":{},"name":"localvolumemigrate-1","namespace":"hwameistor"},"spec":{"migrateAllVols":true,"sourceNodesNames":["dce-172-30-40-61"],"targetNodesNames":["172-30-45-223"],"volumeName":"pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4"}}
  creationTimestamp: "2022-07-07T12:34:31Z"
  generation: 1
  name: localvolumemigrate-1
  namespace: hwameistor
  resourceVersion: "12828637"
  uid: 78af7f1b-d701-4b03-84de-27fafca58764
  spec:
  abort: false
  migrateAllVols: true
  sourceNodesNames:
  - dce-172-30-40-61
    targetNodesNames:
  - 172-30-45-223
    volumeName: pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4
    status:
    replicaNumber: 1
    state: InProgress
    kind: List
    metadata:
    resourceVersion: ""
    selfLink: ""
```

## Step 7: Verify migration results

```console
[root@172-30-45-222 deploy]# kubectl  get lvr
NAME                                              CAPACITY     NODE            STATE   SYNCED   DEVICE                                                                  AGE
pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4-9cdkkn   1073741824   172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4   77s
pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a-7ppmrx   1073741824   172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a   77s
```

## Step 8: Reattach/Remount volume

```console
$ kubectl patch deployment nginx-local-storage-lvm --patch '{"spec": {"replicas": 1}}' -n hwameistor
```