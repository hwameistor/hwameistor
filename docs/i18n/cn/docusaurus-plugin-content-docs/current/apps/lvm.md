---
sidebar_position: 1
sidebar_label: "使用 LVM 数据卷"
---

# 使用 LVM 数据卷

HwameiStor 提供了基于 LVM 的数据卷。
这种类型的数据卷提供了接近原生磁盘的读写性能，并且在此之上提供了数据卷扩容、迁移、HA 等等高级特性。

使用以下命令创建一个 `nginx` 应用并使用数据卷 `hwameistor-lvm-volume`：

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
