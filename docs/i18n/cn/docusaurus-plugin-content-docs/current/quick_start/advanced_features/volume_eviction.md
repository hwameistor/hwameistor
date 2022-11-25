---
sidebar_position: 3
sidebar_label: "数据卷驱逐"
---

# 数据卷驱逐

数据卷迁移和驱逐是 HwameiStor 系统的重要功能，保障 HwameiStor 在生产环境中持续正常运行。
HwameiStor 会将数据卷从一个节点迁移到另一个节点，同时保证数据仍然可用。
当 Kubernetes 节点或者应用 Pod 由于某种原因被驱逐时，系统会自动发现节点或者 Pod 所关联的 HwameiStor 数据卷，并自动将其迁移到其他节点，从而保证被驱逐的 Pod 可以调度到其他节点并正常运行。
此外，运维人员也可以主动迁移数据卷，从而平衡系统资源，保证系统平稳运行。

**驱逐节点**

在 Kubernetes 系统中，可以使用下列命令驱逐节点，将正在该节点上运行的 Pod 移除并迁移到其他节点上。
同时，也将 Pod 使用的 HwameiStor 数据卷从该节点迁移到其他节点，保证 Pod 可以在其他节点上正常运行。

```console
$ kubectl drain k8s-node-1 --ignore-daemonsets=true
```

可以使用下列命令查看所关联的 HwameiStor 数据卷是否迁移成功。

```console
$ kubectl get LocalStorageNode k8s-node-1 -o yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalStorageNode
metadata:
  creationTimestamp: "2022-10-11T07:41:58Z"
  generation: 1
  name: k8s-node-1
  resourceVersion: "6402198"
  uid: c71cc6ac-566a-4e0b-8687-69679b07471f
spec:
  hostname: k8s-node-1
  storageIP: 10.6.113.22
  topogoly:
    region: default
    zone: default
status:
  ...
  pools:
    LocalStorage_PoolHDD:
      class: HDD
      disks:
      - capacityBytes: 17175674880
        devPath: /dev/sdb
        state: InUse
        type: HDD
      freeCapacityBytes: 16101933056
      freeVolumeCount: 999
      name: LocalStorage_PoolHDD
      totalCapacityBytes: 17175674880
      totalVolumeCount: 1000
      type: REGULAR
      usedCapacityBytes: 1073741824
      usedVolumeCount: 1
      volumeCapacityBytesLimit: 17175674880
      # ** 确保 volumes 为空 ** #
      volumes:  
  state: Ready  
```

同时，可以使用下列命令查看被驱逐节点上是否还有 HwameiStor 的数据卷。

```console
$ kubectl get localvolumereplica
NAME                                              CAPACITY     NODE         STATE   SYNCED   DEVICE                                                                  AGE
pvc-1427f36b-adc4-4aef-8d83-93c59064d113-957f7g   1073741824   k8s-node-3   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1427f36b-adc4-4aef-8d83-93c59064d113   20h
pvc-1427f36b-adc4-4aef-8d83-93c59064d113-qlpbmq   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1427f36b-adc4-4aef-8d83-93c59064d113   30m
pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4-scrxjb   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD/pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4      30m
pvc-f8f017f9-eb09-4fbe-9795-a6e2d6873148-5t782b   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-f8f017f9-eb09-4fbe-9795-a6e2d6873148   30m

```

在一些情况下，重启节点时，用户希望仍然保留数据卷在该节点上。可以通过在该节点上添加下列标签实现：

```
$ kubectl label node k8s-node-1 hwameistor.io/eviction=disable
```

**驱逐 Pod**

当 Kubernetes 节点负载过重时，系统会选择性地驱逐一些 Pod，从而释放一些系统资源，保证其他 Pod 正常运行。
如果被驱逐的 Pod 使用了 HwameiStor 数据卷，系统会捕捉到这个被驱逐的 Pod，自动将相关的 HwameiStor 数据卷迁移到其他节点，从而保证该 Pod 能正常运行。

**迁移 Pod**

运维人员可以主动迁移应用 Pod 和其使用的 HwameiStor 数据卷，从而平衡系统资源，保证系统平稳运行。
可以通过下列两种方式进行主动迁移：

```console
kubectl label pod mysql-pod hwameistor.io/eviction=start
kubectl delete pod mysql-pod
```

```console
$ cat << EOF | kubectl apply -f -
apiVersion: hwameistor.io/v1alpha1
kind: LocalVolumeMigrate
metadata:
  name: migrate-pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4
spec:
  sourceNode: k8s-node-1
  targetNodesSuggested: 
  - k8s-node-2
  - k8s-node-3
  volumeName: pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4
  migrateAllVols: true
EOF

$ kubectl delete pod mysql-pod
```
