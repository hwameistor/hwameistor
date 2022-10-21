---
sidebar_position: 2
sidebar_label:  "卷的迁移"
---

# 卷的迁移

`Migrate` 迁移功能是 HwameiStor 中重要的运维管理功能，当应用绑定的数据卷所在节点副本损坏时，卷副本可以通过迁移到其他节点，并在成功迁移到新节点后，将应用重新调度到新节点，并进行数据卷的绑定挂载。

## 基本概念

`LocalVolumeGroup(LVG)`（数据卷组）管理是 HwameiStor 中重要的一个功能。当应用 Pod 申请多个数据卷 `PVC` 时，为了保证 Pod 能正确运行，这些数据卷必须具有某些相同属性，例如：数据卷的副本数量，副本所在的节点。通过数据卷组管理功能正确地管理这些相关联的数据卷，是 HwameiStor 中非常重要的能力。

## 前提条件

LocalVolumeMigrate 需要部署在 Kubernetes 系统中，需要部署应用满足下列条件：

* 支持 lvm 类型的卷
  * 基于 LocalVolume 粒度迁移时，默认所属相同 LocalVolumeGroup 的数据卷不会一并迁移（若一并迁移，需要配置开关 MigrateAllVols：true）

## 步骤 1: 创建 `StorageClass`

```console
$ cd ../../deploy/
$ kubectl apply -f storageclass-lvm.yaml
```

## 步骤 2: 创建 multiple `PVC`

```console
$ kubectl apply -f pvc-multiple-lvm.yaml
```

## 步骤 3: 部署多数据卷 Pod

```console
$ kubectl apply -f nginx-multiple-lvm.yaml
```

## 步骤 4: 解挂载多数据卷 Pod

```console
kubectl -n hwameistor scale --current-replicas=1 --replicas=0 deployment/nginx-local-storage-lvm
```

## 步骤 5: 创建迁移任务

```console
cat > ./migrate_lv.yaml <<- EOF
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

在迁移时，

1）如果指定了targetNodesSuggested，系统会从指定的节点中，选择一个适合的进行迁移。如果都不合适，迁移操作失败;

2）如果不指定 targetNodesSuggested，系统会根据容量平衡原则自动选择一个适合的节点进行迁移。

```console
$ kubectl apply -f ./migrate_lv.yaml
```

## 步骤 6: 查看迁移状态

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

## 步骤 7: 查看迁移成功状态

```console
[root@172-30-45-222 deploy]# kubectl  get lvr
NAME                                              CAPACITY     NODE            STATE   SYNCED   DEVICE                                                                  AGE
pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4-9cdkkn   1073741824   172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1a0913ac-32b9-46fe-8258-39b4e3b696a4   77s
pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a-7ppmrx   1073741824   172-30-45-223   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-d9d3ae9f-64af-44de-baad-4c69b9e0744a   77s
```

## 步骤 8: 迁移成功后，重新挂载数据卷 Pod

```console
$ kubectl -n hwameistor scale --current-replicas=0 --replicas=1 deployment/nginx-local-storage-lvm
```
