---
sidebar_position: 2
sidebar_label: "Use Disk Volume"
---

# Use Disk Volume

HwameiStor provides another type of data volume known as raw disk data volume.
This volume is based on the raw disk present on the node and can be directly mounted for container use.
As a result, this type of data volume offers more efficient data read and write performance,
thereby fully unleashing the performance of the disk.

Create a nginx application and use `hwameistor-disk-volume` volume using the following command:

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
EOF
```


