---
sidebar_position: 8
sidebar_label:  "本地缓存卷 "
---

# 本地缓存卷

使用 HwameiStor 运行 AI 训练应用程序非常简单

作为示例，我们将通过创建本地缓存卷来部署 Nginx 应用程序。

在生产实践中替换成对应的训练Pod即可，这里只是为了简化演示。专注于如何使用加载数据集

使用前请确保集群已安装Dragonfly,并完成相关配置

## 安装 Dragonfly

1. 根据集群配置/etc/hosts
2. 根据选择进行配置默认sc
   kubectl patch storageclass hwameistor-storage-lvm-hdd -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

3. 使用helm安装dragonfly
   helm repo add dragonfly https://dragonflyoss.github.io/helm-charts/
   helm install --create-namespace --namespace dragonfly-system dragonfly dragonfly/dragonfly --version 1.1.63
4. dragonfly-dfdaemon 配置
```console
$ kubectl -n dragonfly-system get ds
$ kubectl -n dragonfly-system edit ds dragonfly-dfdaemon

...
spec:
      spec:
        containers:
        - image: docker.io/dragonflyoss/dfdaemon:v2.1.45
       ...
          securityContext:
            capabilities:
              add:
              - SYS_ADMIN
            privileged: true
          volumeMounts:
          ...
            
          - mountPath: /var/run
            name: host-run
          - mountPath: /mnt
            mountPropagation: Bidirectional
            name: host-mnt
          ...
      volumes:
      ...
      - hostPath:
          path: /var/run
          type: DirectoryOrCreate
        name: host-run
      - hostPath:
          path: /mnt
          type: DirectoryOrCreate
        name: host-mnt
      ... 

```

5. 安装dfget 客户端命令行工具
   每个节点执行：
   wget https://github.com/dragonflyoss/Dragonfly2/releases/download/v2.1.44/dfget-2.1.44-linux-amd64.rpm
   rpm -ivh dfget-2.1.44-linux-amd64.rpm
6. 取消集群默认配置sc


## 查看 `DataSet`

以 minio 为例

```console
apiVersion: datastore.io/v1alpha1
kind: DataSet
metadata:
  name: dataset-test
spec:
  refresh: true
  type: minio
  minio:
    endpoint: Your service ip address:9000
    bucket: BucketName/Dir  #根据你的数据集所在的目录级别定义
    secretKey: minioadmin
    accessKey: minioadmin
    region: ap-southeast-2  
```

## 步骤一 创建 `DataSet`


```Console
$ kubectl apply -f dataset.yaml
```

确认缓存卷已成功创建

```Console

$ kubectl get lv dataset-test
NAME           POOL                   REPLICAS   CAPACITY     USED        STATE   PUBLISHED   AGE
dataset-test   LocalStorage_PoolHDD   3          1073741824   906514432   Ready               20d

$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                    STORAGECLASS                    REASON   AGE
dataset-test                               1Gi        ROX            Retain           Bound    default/hwameistor-dataset                                                                        20d
```

PV的大小是根据你数据集的大小而决定的，您也可以手动配置

## 步骤二 创建 PVC 绑定 PV

```Console
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hwameistor-dataset
  namespace: default
spec:
  accessModes:
  - ReadOnlyMany
  resources:
    requests:
      storage: 1Gi  # 数据集大小
  volumeMode: Filesystem
  volumeName: dataset-test
```

确认pvc已经创建成功

```Console

## Verify  PVC

$ kubectl get pvc
NAME                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                    AGE
hwameistor-dataset   Bound    dataset-test                               1Gi        ROX                                            20d
```

## 步骤三 创建 `StatefulSet`

```Console
$ kubectl apply -f sts-nginx-AI.yaml
```

Please note the `claimName` uses the name of the pvc bound to the dataset

```yaml
    spec:
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: hwameistor-dataset
```
## Verify Nginx Pod
```Console
$ kubectl get pod
NAME               READY   STATUS            RESTARTS   AGE
nginx-dataload-0   1/1     Running           0          3m58s
$ kubectl  logs nginx-dataload-0 hwameistor-dataloader
Created custom resource
Custom resource deleted, exiting
DataLoad execution time: 1m20.24310857s
```
根据日志，加载数据耗时1m20.24310857s

## [Optional] Scale Nginx out into a 3-node Cluster

HwameiStor 缓存卷支持 `StatefulSet` 横向扩展。`StatefulSet` 的每个 `pod` 都会附加并挂载一个绑定同一份数据集的 HwameiStor 缓存卷。

```console
$ kubectl scale sts/sts-nginx-AI --replicas=3

$ kubectl get pod -o wide
NAME               READY   STATUS    RESTARTS   AGE
nginx-dataload-0   1/1     Running   0          41m
nginx-dataload-1   1/1     Running   0          37m
nginx-dataload-2   1/1     Running   0          35m


$ kubectl logs nginx-dataload-1 hwameistor-dataloader
Created custom resource
Custom resource deleted, exiting
DataLoad execution time: 3.24310857s

$ kubectl logs nginx-dataload-2 hwameistor-dataloader
Created custom resource
Custom resource deleted, exiting
DataLoad execution time: 2.598923144s

```

根据日志，第二次和第三次加载数据只耗时3.24310857s、2.598923144s 。对比首次加载速度得到了很大的提升