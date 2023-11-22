---
sidebar_position: 2
sidebar_label: "使用裸磁盘数据卷"
---

# 使用裸磁盘数据卷

HwameiStor 提供的另一种类型数据卷是裸磁盘数据卷。
这种数据卷是基于节点上面的裸磁盘并将其直接挂载给容器使用。
因此这种类型的数据卷提供了更高效的数据读写性能，将磁盘的性能完全释放。

使用以下命令创建一个 `nginx` 应用并使用数据卷 `hwameistor-disk-volume`：

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
