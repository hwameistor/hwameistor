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
- Setup the StorageClasses automatically according to the HwameiStor's configurations and capabilities;

## Steps

1. Install hwameistor-operator

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator
   ```
