---
id: "intro"
sidebar_position: 1
sidebar_label: "What is HwameiStor"
---

# What is HwameiStor

HwameiStor is an HA local storage system for cloud-native stateful workloads.

HwameiStor creates a local storage resource pool for centrally managing all disks
such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed
services with local volumes, and provides data persistence capabilities for
stateful cloud-native workloads or components.

HwameiStor is an open source, lightweight, and cost-efficient local storage system
that can replace expensive traditional SAN storage. The system architecture of HwameiStor is as follows.

![System architecture](img/architecture.png)

By using the CAS pattern, users can achieve the benefits of higher performance,
better cost-efficiency, and easier management of their container storage.
It can be deployed by helm charts or directly use the independent installation.
You can easily enable high-performance local storage across the entire cluster
with one click and automatically identify disks.

HwameiStor is easy to deploy and ready to go.

## Features

1. Automated Maintenance

    Disks can be automatically discovered, identified, managed, and allocated.
    Smart scheduling of applications and data based on affinity. Automatically
    monitor disk status and give early warning.

2. High Availability

    Use cross-node replicas to synchronize data for high availability.
    When a problem occurs, the application will be automatically scheduled to
    the high-availability data node to ensure the continuity of the application.

3. Full-Range support of Storage Medium

   Aggregate HDD, SSD, and NVMe disks to provide low-latency, high-throughput data services.

4. Agile Linear Scalability

   Dynamically expand the cluster according to flexibly meet the data persistence requirements of the application.
