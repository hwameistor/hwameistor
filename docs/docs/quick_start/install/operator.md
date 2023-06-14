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
- Configure the disks for different purpose
- Set up the storage pools automatically by discovering the underlying disks' type (e.g. HDD, SSD)
- Set up the StorageClasses automatically according to the Hwameistor's configurations and capabilities

## Steps

1. Add hwameistor-operator Helm Repo

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   ```

2. Install hwameistor-operator

   :::note
   If no available clean disk provided, the operator will not create StorageClass automatically.
   Operator will claim disk automatically while installing, the available disks will be added into
   pool of LocalStorageNode. If available clean disk provided after installing, it's needed to apply
   a LocalDiskClaim manually to added the disk into pool of LocalStorageNode. Once LocalStorageNode has
   any disk available in its pool, the operator will create StorageClass automatically.
   That is to say, no capacity, no StorageClass.
   :::

   ```console
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace
   ```

Optional installation parameters:

- Enable authentication

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator  -n hwameistor --create-namespace\
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
