---
sidebar_position: 3
sidebar_label: "Eviction"
---

# Volume Eviction

Volume eviction/migration is one of most important funcions in the HwameiStor, especially in the production environment.
HwameiStor will keep the data in the volume when migrate it.

Once a Kubernetes node or pod is evicted by the system for any reason, HwameiStor will detect all the volume replicas located on this node, and automatically migrate them to other avaliable nodes. So that, the evicted pods can be rescheduled to another node and continue to run.

## Node Eviction

In a Kubernetes cluster, a node can be drained by using the following procedure. So that, all the pods and volume replicas on this node will be evicted, and then continue the services on other avaliable nodes.

```console
$ kubectl label k8s-node-1 hwameistor.io/eviction=start
$ kubectl drain k8s-node-1 --ignore-daemonsets=true
```

Check if all the volumes' migration complete or not by:

```console
$ kubectl get node k8s-node-1 -o yaml
apiVersion: v1
kind: Node
metadata:
  name: k8s-node-1
  labels:
    ...
    hwameistor.io/eviction: completed
    ...
  name: k8s-node-1
spec:
  ...
status:
  ...
```

Check if there is any volume replica still located in the evicted node by:

```console
$ kubectl get localvolumereplica
NAME                                              CAPACITY     NODE         STATE   SYNCED   DEVICE                                                                  AGE
pvc-1427f36b-adc4-4aef-8d83-93c59064d113-957f7g   1073741824   k8s-node-3   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1427f36b-adc4-4aef-8d83-93c59064d113   20h
pvc-1427f36b-adc4-4aef-8d83-93c59064d113-qlpbmq   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-1427f36b-adc4-4aef-8d83-93c59064d113   30m
pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4-scrxjb   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD/pvc-6ca4c0d4-da10-4e2e-83b2-19cbf5c5e3e4      30m
pvc-f8f017f9-eb09-4fbe-9795-a6e2d6873148-5t782b   1073741824   k8s-node-2   Ready   true     /dev/LocalStorage_PoolHDD-HA/pvc-f8f017f9-eb09-4fbe-9795-a6e2d6873148   30m

```

## Pod Eviction

When a Kubernetes node is overloaded, it will evict some low-priority pods to recycle system's resources to keep other pods safe. HwameiStor will detect the evicted pod and migrate the associated volumes to another available node. So that, the pod can continue to run on it.

## Pod Migration

The migration can be pro-actively triggered on the pod and associated HwameiStor volume by using either one of following methods.

1) Method #1

```console
$ kubectl label pod mysql-pod hwameistor.io/eviction=start
$ kubectl delete pod mysql-pod
```

2) Method #2

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
