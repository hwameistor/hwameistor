---
sidebar_position: 2
sidebar_label: "Deploy with hwameistor-operator"
---

# Deploy with hwameistor-operator

You can use hwameistor-operator to deploy and manage HwameiStor system.

- Perform Life Cycle Management (LCM) for HwameiStor components:
  - LocalDiskManager
  - LocalStorage
  - Scheduler
  - AdmissionController
  - VolumeEvictor
  - Exporter
  - HA module
  - Apiserver
  - Graph UI
- Configure the disks for different purposes
- Set up the storage pools automatically by discovering the underlying disks' type (e.g. HDD, SSD)
- Set up the StorageClasses automatically according to the Hwameistor's configurations and capabilities

## Steps

1. Add hwameistor-operator Helm Repo

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   ```

2. Install HwameiStor with hwameistor-operator

   :::note
   If no available clean disk provided, the operator will not create StorageClass automatically.
   Operator will claim disk automatically while installing, the available disks will be added into
   pool of LocalStorageNode. If available clean disk provided after installing, it's needed to apply
   a LocalDiskClaim manually to add the disk into pool of LocalStorageNode. Once LocalStorageNode has
   any disk available in its pool, the operator will create StorageClass automatically.
   That is to say, no capacity, no StorageClass.
   :::

   ```console
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace
   ```

Optional installation parameters:

- Disk Reserve

  Available clean disk will be claimed and added into pool of LocalStorageNode by default. If you want to
  reserve some disks for other use before installing, you can set diskReserveConfigurations by helm values.

  Method 1:

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace \
  --set diskReserve\[0\].nodeName=node1 \
  --set diskReserve\[0\].devices={/dev/sdc\,/dev/sdd} \
  --set diskReserve\[1\].nodeName=node2 \
  --set diskReserve\[1\].devices={/dev/sdc\,/dev/sde}
  ```

  This is a example to set diskReserveConfigurations by `helm install --set`, it may be hard to
  write `--set` options like that. If it's possible, we suggest write the diskReserveConfigurations
  values into a file.

  Method 2:

  ```console
  diskReserve:
  - nodeName: node1
    devices:
    - /dev/sdc
    - /dev/sdd
  - nodeName: node2
    devices:
    - /dev/sdc
    - /dev/sde
  ```

  For example, you write values like this into a file call diskReserve.yaml,
  you can apply the file when running `helm install`.

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace -f diskReserve.yaml
  ```

- Enable authentication

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator  -n hwameistor --create-namespace \
  --set apiserver.authentication.enable=true \
  --set apiserver.authentication.accessId={YourName} \
  --set apiserver.authentication.secretKey={YourPassword}
  ```

  You can also enable authentication by editing deployment/apiserver.

- Install operator by using DaoCloud image registry:

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator  -n hwameistor --create-namespace \
  --set global.hwameistorImageRegistry=ghcr.m.daocloud.io \
  --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
  ```
