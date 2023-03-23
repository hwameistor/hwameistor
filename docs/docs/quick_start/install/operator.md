---
sidebar_position: 3
sidebar_label: "Deploy with hwameistor-operator"
---

# Deploy with hwameistor-operator

You can use hwameistor-operator to deploy and manage HwameiStor components.

- Perform Life Cycle Management (LCM) for HwameiStor components:
  - Apiserver
  - LocalStorage
  - LocalDiskManager
  - Scheduler
  - AdmissionController
  - VolumeEvictor
  - Exporter
- Automate local disk claim for ensuring HwameiStor ready
- Manage admission control configuration for verifying HwameiStor volumes

## Steps

1. Install hwameistor-operator

   ```bash
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   ```

2. Deploy HwameiStor with hwameistor-operator

   ```bash
   helm repo update hwameistor-operator
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator
   kubectl apply -f https://raw.githubusercontent.com/hwameistor/hwameistor-operator/main/config/samples/hwameistor.io_hmcluster.yaml
   ```
