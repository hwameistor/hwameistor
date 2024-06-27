---
sidebar_position: 8
sidebar_label:  "Local Cache Volumes "
---

# Local Cache Volumes

It is very simple to run AI training applications using HwameiStor

As an example, we will deploy an Nginx application by creating a local cache volume.

Before use, please ensure that Dragonfly has been installed in the cluster and relevant configurations have been completed.

## Install Dragonfly
1. Configure /etc/hosts according to the cluster
2. Configure the default sc according to the selection
   kubectl patch storageclass hwameistor-storage-lvm-hdd -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
3. Install dragonfly using helm
   helm repo add dragonfly https://dragonflyoss.github.io/helm-charts/
   helm install --create-namespace --namespace dragonfly-system dragonfly dragonfly/dragonfly --version 1.1.63
4. dragonfly-dfdaemon configuration
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
5. Install the dfget client command line tool
   Each node executes:
   wget https://github.com/dragonflyoss/Dragonfly2/releases/download/v2.1.44/dfget-2.1.44-linux-amd64.rpm
   rpm -ivh dfget-2.1.44-linux-amd64.rpm
6. Cancel the cluster default configuration sc

   
## Verify `DataSet`

Take minio as an example

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
    bucket: BucketName/Dir  #Defined according to the directory level where your dataset is located
    secretKey: minioadmin
    accessKey: minioadmin
    region: ap-southeast-2  
```

## Create `DataSet`


```Console
$ kubectl apply -f dataset.yaml
```

Confirm that the cache volume has been created successfully

```Console
$ k get dataset
NAME           TYPE    LASTREFRESHTIME   CONNECTED   AGE     ERROR
dataset-test   minio                                 4m38s

$ k get lv
NAME                                       POOL                   REPLICAS   CAPACITY     USED        STATE   PUBLISHED           AGE
dataset-test                               LocalStorage_PoolHDD              211812352                Ready                       4m27s

$ k get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                                                    STORAGECLASS                 REASON   AGE
dataset-test                               202Mi      ROX            Retain           Available                                                                                                  35s

```

The size of pv is determined by the size of your data set

## Create a PVC and bind it to dataset PV

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
      storage: 202Mi  #dataset size
  volumeMode: Filesystem
  volumeName: dataset-test  #dataset name
```

Confirm that the pvc has been created successfully

```Console

## Verify  PVC

$ k get pvc
k get pvc
NAME                 STATUS   VOLUME         CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hwameistor-dataset   Bound    dataset-test   202Mi      ROX                           4s
```

## Create `StatefulSet`

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
According to the log, loading data took 1m20.24310857s

## [Optional] Scale Nginx out into a 3-node Cluster

HwameiStor cache volumes support horizontal expansion of `StatefulSet`. Each `pod` of `StatefulSet` will attach and mount a HwameiStor cache volume bound to the same dataset.

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

According to the log, the second and third loading of data only took 3.24310857s and 2.598923144s respectively. Compared with the first loading, the speed has been greatly improved.