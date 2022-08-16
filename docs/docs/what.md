---
id: "intro"
sidebar_position: 1
sidebar_label: "What is HwameiStor"
---

# What is HwameiStor

Hwameistor is an HA local storage system for cloud-native stateful workloads.

HwameiStor creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes, and provides data persistence capabilities for stateful cloud-native workloads or components.

HwameiStor is an open source, lightweight, and cost-efficient local storage system that can replace expensive traditional SAN storage. The system architecture of HwameStor is as follows.

![System architecture](img/architecture.png)

 By using the CAS pattern, users can achieve the benefits of higher performance, better cost-efficiency, and easier management for their container storage. It can be deployed by helm charts or directly use the independent installation. You can easily enable the high-performance local storage across entire cluster with one click and automatically identify disks.

HwameiStor is easy to deploy and ready to go.
