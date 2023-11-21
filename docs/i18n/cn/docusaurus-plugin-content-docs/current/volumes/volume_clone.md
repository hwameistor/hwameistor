---
sidebar_position: 5
sidebar_label: "数据卷克隆"
---

# 数据卷克隆

在 HwameiStor 中，用户可以为数据卷创建克隆卷，克隆的数据卷具有和源卷在克隆发生时刻相同的数据。

:::note
目前仅支持对非高可用 LVM 类型数据卷创建克隆卷，并且只支持原地克隆。

数据卷克隆使用了快照技术实现，为了避免数据不一致，请先暂停或者停止 I/O 然后再克隆。
:::

请按照以下步骤来创建和使用克隆卷。

## 1. 创建克隆卷

可以创建 pvc，对数据卷进行克隆操作。具体如下：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hwameistor-lvm-volume-clone
spec:
  storageClassName: hwameistor-storage-lvm-ssd
  dataSource:
    # 必须提供已经 Bound 的数据卷
    name: data-sts-mysql-local-0
    kind: PersistentVolumeClaim
    apiGroup: ""
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

## 2. 使用克隆卷

使用以下命令创建一个 `nginx` 应用并使用数据卷 `hwameistor-lvm-volume-clone`：

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
