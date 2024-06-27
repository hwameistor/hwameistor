---
sidebar_position: 8
sidebar_label:  "Local Cache Volumes "
---

# Local Cache Volumes

It is very simple to run AI training applications using HwameiStor.

As an example, we will deploy an Nginx application by creating a local cache volume.

Before use, please ensure that Dragonfly has been installed in the cluster and relevant configurations have been completed.

## Install Dragonfly
1. Configure /etc/hosts according to the cluster.
   ```console
   $ vi /etc/hosts
   host1-IP hostName1
   host2-IP hostName2
   host3-IP hostName3
   ```
2. To install Dragonfly components, ensure a default storage class is configured, as it is required to create storage volumes.
   ```console
    kubectl patch storageclass hwameistor-storage-lvm-hdd -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
   ```

3. Install dragonfly using helm.
   ```console
   $ helm repo add dragonfly https://dragonflyoss.github.io/helm-charts/
   $ helm install --create-namespace --namespace dragonfly-system dragonfly dragonfly/dragonfly --version 1.1.63
   ```

4. dragonfly-dfdaemon configuration.
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

5. Install the dfget client command line tool.
   Each node executes:
   ```console
   $ wget https://github.com/dragonflyoss/Dragonfly2/releases/download/v2.1.44/dfget-2.1.44-linux-amd64.rpm
   $ rpm -ivh dfget-2.1.44-linux-amd64.rpm
   ```

6. To avoid issues, cancel the previously configured default storage class.
   ```console
    kubectl patch storageclass hwameistor-storage-lvm-hdd -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
   ```

## Verify dragonfly
```console
$  kubectl -n dragonfly-system get pod -owide
NAME                                 READY   STATUS    RESTARTS      AGE   IP                NODE                NOMINATED NODE   READINESS GATES
dragonfly-dfdaemon-d2fzp             1/1     Running   0             19h   200.200.169.158   hwameistor-test-1   <none>           <none>
dragonfly-dfdaemon-p7smf             1/1     Running   0             19h   200.200.29.171    hwameistor-test-3   <none>           <none>
dragonfly-dfdaemon-tcwkr             1/1     Running   0             19h   200.200.39.71     hwameistor-test-2   <none>           <none>
dragonfly-manager-5479bf9bc9-tp4g5   1/1     Running   1 (19h ago)   19h   200.200.29.174    hwameistor-test-3   <none>           <none>
dragonfly-manager-5479bf9bc9-wpbr6   1/1     Running   0             19h   200.200.39.92     hwameistor-test-2   <none>           <none>
dragonfly-manager-5479bf9bc9-zvrdj   1/1     Running   0             19h   200.200.169.142   hwameistor-test-1   <none>           <none>
dragonfly-mysql-0                    1/1     Running   0             19h   200.200.29.178    hwameistor-test-3   <none>           <none>
dragonfly-redis-master-0             1/1     Running   0             19h   200.200.169.137   hwameistor-test-1   <none>           <none>
dragonfly-redis-replicas-0           1/1     Running   0             19h   200.200.39.72     hwameistor-test-2   <none>           <none>
dragonfly-redis-replicas-1           1/1     Running   0             19h   200.200.29.130    hwameistor-test-3   <none>           <none>
dragonfly-redis-replicas-2           1/1     Running   0             19h   200.200.169.134   hwameistor-test-1   <none>           <none>
dragonfly-scheduler-0                1/1     Running   0             19h   200.200.169.190   hwameistor-test-1   <none>           <none>
dragonfly-scheduler-1                1/1     Running   0             19h   200.200.39.76     hwameistor-test-2   <none>           <none>
dragonfly-scheduler-2                1/1     Running   0             19h   200.200.29.163    hwameistor-test-3   <none>           <none>
dragonfly-seed-peer-0                1/1     Running   1 (19h ago)   19h   200.200.169.138   hwameistor-test-1   <none>           <none>
dragonfly-seed-peer-1                1/1     Running   0             19h   200.200.39.80     hwameistor-test-2   <none>           <none>
dragonfly-seed-peer-2                1/1     Running   0             19h   200.200.29.151    hwameistor-test-3   <none>           <none>
```

   
## Verify `DataSet`

Take minio as an example

```yaml
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

## Create a `PVC` and bind it to dataset PV

```yaml
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

```yaml
   apiVersion: apps/v1
   kind: StatefulSet
   metadata:
      name: nginx-dataload
      namespace: default
   spec:
      serviceName: nginx-dataload
      replicas: 1
      selector:
         matchLabels:
            app: nginx-dataload
      template:
         metadata:
            labels:
               app: nginx-dataload
         spec:
            hostNetwork: true
            hostPID: true
            hostIPC: true
            containers:
               - name: nginx
                 image: docker.io/library/nginx:latest
                 imagePullPolicy: IfNotPresent
                 securityContext:
                    privileged: true
                 env:
                    - name: DATASET_NAME
                      value: dataset-test
                 volumeMounts:
                    - name: data
                      mountPath: /data
                 ports:
                    - containerPort: 80
            volumes:
               - name: data
                 persistentVolumeClaim:
                    claimName: hwameistor-dataset
```
:::info
`claimName` uses the name of the pvc bound to the dataset. env: DATASET_NAME=datasetName
:::

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