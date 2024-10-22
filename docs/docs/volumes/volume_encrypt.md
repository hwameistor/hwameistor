---
sidebar_position: 10
sidebar_label:  "Volumes Encrypt"
---


#  Volume encryption

In HwameiStor, users can use LUKS to encrypt local data volumes.

:::note
Currently,  non-HA encrypted volumes do not support lvmigrate.
:::


Please follow the steps below to create and use encrypted volumes.

## 1. Create a Secret for encrypting the data volume

The details are as follows:

```yaml
apiVersion: v1
data:
  key: H4sIAAAAAAAA/+xVP2/bPhDd9SluyGAHUJTgtzFTgPyKLkmDZiwK4kxebNYUSRwp10Xb715IkWzLf5II6eAC3iTe3eO9R/Le1PoJWpEB0AJthcl4J41LxAu0Av67jBlAVIyBdpZpmYgdWmlxQjbWIACBfUlpRlUUEDyn756XRVjm6/WtNMkUrFEo4Gz08OlW3t/c/T/OuLIkn4ylKLIcCkqqWJcUdTRuLOS9DfI63NTml8X5xQ8sbdZSUN49mWmD+c1Pp
  MOSBETihVF0551Jnot1193HZQYw885zxxSe0EbKAObVhNhRoiijXqMD5MDekgByOnjjUs2aqSnvp0VfsaKehE1vzVdCnlJ6DgqQMpVBbijXUWiAUNVnJ2BOFJrivchSlpRQbvb9zP45r4N7Q2pgiuTSuoJpSksBo0628XXi6n29derJGnNnp7DMMZjDKr6Ah1ozxShbgefG6cFFq3YiBWRMngVcb/Z37zVdjy7Ox+1isKioJJcEnP28+r3nhJ3XdLx8HrweRid4P
  YRN3UAMqGifMhuxNwN293XFrI/ZhocgBq8PoQ0kWyMp7xIaR3wIc5XQe0Wa/aBXVG8VZhj7z/QDGkv612OlFJEmPf6Lynbza98lybdyu+u4W/D6++5usDw4LWcYZ02w9LqytSldNb+dlnW8fPnkejCtemejx483n2/HPax2upWU2Ci5Pe3hy9dhrnN1cp0jdZ35Qk+Od0yfbOdkOyfbeZftvPbA/zXf+RMAAP//IK/8i+YNAAA=
kind: Secret
metadata:
  name: hwameistor-encrypt-secret
  namespace: hwameistor
type: Opaque  
```

## 2. Create a StorageClass

Use the following command to create a StorageClass and specify the Secret created above:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage-hdd-encrypt
provisioner: lvm.hwameistor.io
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
reclaimPolicy: Delete
parameters:
  secretName: hwameistor/hwameistor-encrypt-secret
  encryptType: LUKS
  replicaNumber: "1"
  poolClass: "HDD"
  poolType: "REGULAR"
  volumeKind: "LVM"
  striped: "true"
  csi.storage.k8s.io/fstype: "xfs"
```
## 3. Create a pvc and an Deployment 

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: local-storage-pvc-encrypt
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-storage-hdd-encrypt
  resources:
    requests:
      storage: 1Gi
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-local-storage-lvm
  labels:
    app: nginx-local-storage-lvm
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-local-storage-lvm
  template:
    metadata:
      labels:
        app: nginx-local-storage-lvm
      name: nginx-local-storage-lvm
    spec:
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      containers:
        - image: nginx:latest
          imagePullPolicy: IfNotPresent
          name: nginx
          ports:
            - containerPort: 80
          command:
            - sh
            - -xc
            - |
              VOL="$( df | grep /usr/share/nginx/html | awk '{print $1,$NF}' )"
              echo "<center><h1>Demo volume: ${VOL}</h1></center>" > /usr/share/nginx/html/index.html
              nginx -g "daemon off;"
          volumeMounts:
            - name: html-root
              mountPath: /usr/share/nginx/html
          resources:
            limits:
              cpu: '100m'
              memory: '100Mi'
      volumes:
        - name: html-root
          persistentVolumeClaim:
            claimName: local-storage-pvc-encrypt
```

## 4.Check pod status

```console
# kubectl get pod -o wide                                                                                                                                                                                                                             
NAME                                      READY   STATUS    RESTARTS   AGE   IP               NODE        NOMINATED NODE   READINESS GATES                                                                                                                                  
nginx-local-storage-lvm-79886d9dd-44fsg   1/1     Running   0          20m   100.111.156.91   k8s-node1                     
```

## 5.Check whether the volume is an encrypted volume

You can use the "lsblk" to check whether the TYPE of the volume is crypt.
```console
# lsblk                                                                                                                                                                                                                                                   
NAME                                                                 MAJ:MIN RM   SIZE RO TYPE  MOUNTPOINT                                                                                                                                                                  
sda                                                                    8:0    0   160G  0 disk                                                                                                                                                                              
├─sda1                                                                 8:1    0     1G  0 part  /boot                                                                                                                                                                       
└─sda2                                                                 8:2    0   159G  0 part                                                                                                                                                                              
  ├─centos-root                                                      253:0    0    50G  0 lvm   /                                                                                                                                                                           
  ├─centos-swap                                                      253:1    0   7.9G  0 lvm                                                                                                                                                                               
  └─centos-home                                                      253:2    0 101.1G  0 lvm   /home                                                                                                                                                                       
sdb                                                                    8:16   0   200G  0 disk                                                                                                                                                                              
└─LocalStorage_PoolHDD-pvc--2c097032--690d--4510--99ad--54119b6b650c 253:3    0     1G  0 lvm                                                                                                                                                                               
  └─pvc-2c097032-690d-4510-99ad-54119b6b650c-encrypt                 253:4    0  1008M  0 crypt /var/lib/kubelet/pods/4c2b76f3-a84f-4e62-88c8-a71abeb68efd/volumes/kubernetes.io~csi/pvc-2c097032-690d-4510-99ad-54119b6b650c/mount                                         
sr0                                                                   11:0    1  1024M  0 rom                      
```

You can use  the "blkid" command to check whether the TYPE of the volume is crypto_LUKS
```console
# blkid /dev/LocalStorage_PoolHDD/pvc-2c097032-690d-4510-99ad-54119b6b650c                                                                                                                                                                                
/dev/LocalStorage_PoolHDD/pvc-2c097032-690d-4510-99ad-54119b6b650c: UUID="a1910adf-f1dc-45a4-aeb3-6a8cf045bb9d" TYPE="crypto_LUKS"                
```