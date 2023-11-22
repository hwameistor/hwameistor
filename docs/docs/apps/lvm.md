---
sidebar_position: 1
sidebar_label: "Use LVM Volume"
---

# Use LVM Volume

HwameiStor provides LVM-based data volumes,
which offer read and write performance comparable to that of native disks.
These data volumes also provide advanced features such as data volume expansion, migration, high availability, and more.

Create a nginx application and use `hwameistor-lvm-volume` volume using the following command:

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
      claimName: hwameistor-lvm-volume
EOF     
```
