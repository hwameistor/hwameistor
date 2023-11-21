---
sidebar_position: 5
sidebar_label: "Volume Clone"
---

# Volume Clone

In HwameiStor, users can create clone volume for data volumes that have the same data as the source volume at the moment the clone occurs.

:::note
Creating clones of non-HA volume is currently only supported, and only in-place cloning is supported.

Volume clone is implemented using snapshot technology. To avoid data inconsistency, please pause or stop I/O before cloning.
:::

Follow the steps below to create and use clone volume.

## 1. Create Clone Volume

You can create a pvc to perform a cloning operation on a data volume. The details are as follows:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
    name: hwameistor-lvm-volume-clone
spec:
  storageClassName: hwameistor-storage-lvm-ssd
  dataSource:
    # Bound data volumes must be provided
    name: data-sts-mysql-local-0
    kind: PersistentVolumeClaim
    apiGroup: ""
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

## 2. Use Clone Volume

Use the following command to create an `nginx` application and use the data volume `hwameistor-lvm-volume-clone`:

```console
cat << EOF | kubectl apply -f -
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
      claimName: hwameistor-lvm-volume-clone
EOF
```
