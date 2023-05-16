---
sidebar_position: 2
sidebar_label: "Disk Volume"
---

# Disk Volume

Another type of data volume provided by HwameiStor is the raw disk data volume. 
This type of data volume is based on the raw disk on the node and is directly mounted for container use. 
Therefore, this type of data volume provides more efficient data read and write performance, 
fully releasing the performance of the disk.

## Steps

### 1. Prepare Disk Storage Node

It is necessary to ensure that the storage node has available capacity. If there is not enough capacity,
please refer to [Expanding LVM Storage Nodes](../node_expansion/disk_nodes.md).

Check for available capacity using the following command:

```shell
$ kubectl get loaldisknodes
NAME           FREECAPACITY   TOTALCAPACITY   TOTALDISK   STATUS   AGE
k8s-worker-2   1073741824     1073741824      1           Ready    19d
```

### 2. Prepare StorageClass

Create a StorageClass named `hwameistor-storage-disk-ssd` using the following command:

```console
$ cat << EOF | kubectl apply -f - 
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:  
  name: hwameistor-storage-disk-ssd
parameters:
  diskType: SSD
provisioner: disk.hwameistor.io
allowVolumeExpansion: false
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
EOF 
```

### 3. Create Volume

Create a PVC named `hwameistor-disk-volume` using the following command:

```console
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hwameistor-disk-volume
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: hwameistor-storage-disk-ssd
EOF
```

### 4. Use Volume

Create an nginx application and use `hwameistor-disk-volume` volume using the following command:

```console
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
containers:
- name: nginx
  image: docker.io/library/nginx:latest
  imagePullPolicy: IfNotPresent
  volumeMounts:
  - name: data
    mountPath: /data
  ports:
  - containerPort: 80
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: hwameistor-disk-volume
```


