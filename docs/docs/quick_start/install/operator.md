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
- Configure the disks for different purpose;
- Setup the storage pools automatically by discovering the underlying disks' type (e.g. HDD, SSD);
- Setup the StorageClasses automatically according to the Hwameistor's configurations and capabilities;

## Steps

1. Add hwameistor-operator Helm Repo

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   ```

2. Install hwameistor-operator

   Note:If no available clean disk provided, operator will not create storageclass automaticlly.
   Operator will claim disk automaticlly while installing, the available disks will be added into
   pool of localstoragenode. If available clean disk provided after installing, it's needed to apply 
   a localdiskclaim manually to added the disk into pool of localstoragenode. Once localstoragenode has
   pool holding disks, the operator will create storageclass automaticlly, that means, no capacity, no storageclass.

   ```console
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator
   ```

Optional installation parameters:
- Enable authentication
  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator \
  --set apiserver.authentication.enable=true \
  --set apiserver.authentication.accessId={YourName} \
  --set apiserver.authentication.secretKey={YourPassword}
  ```
  You can enable authentication by edit deployment/apiserver also.

- Install operator by with DaoCloud image registry:
  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator \
  --set global.hwameistorImageRegistry=ghcr.m.daocloud.io \
  --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
  ```