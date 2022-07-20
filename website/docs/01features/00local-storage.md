---
sidebar_position: 1
sidebar_label: "Local Storage"
---

# Local Storage

Local Storage is one of modules of HwameiStor which is a cloud native local storage system. It aims to provision high performance and persistent LVM volume with local access to applications.

Supported kinds for local volume: `LVM`.

Supported types for disk: `HDD`, `SSD`, `NVMe`.

## Applicable scenarios

HwameiStor provisions two kind of local volumes: LVM and Disk. As a component of HwameiStor, Local Storage provisions two types of local LVM volumes, such as HA and non-HA.

For the non-HA local LVM volume, it's the best solution for data persistency in the following use cases:

- **Database** with HA capability, such as MongoDB, etc.
- **Messaging system** with HA capability, such as Kafka, RabbitMQ, etc.
- **Key-value store** with HA capability, such as Redis, etc.
- Others with HA capability

For the HA local LVM volume, it's the best solution for data persistency in the following use cases:

- **Database**, such as MySQL, PostgreSQL, etc.
- Other applications which require the data with HA features.

## Usage with Helm Chart

Local Storage is a component of HwameiStor and must work with [Local Disk Manager](01local-disk-manager.md) module. It's suggested to [install by helm-charts](../02installation/01helm-chart.md).

## Usage with Independent Installation

Developer can start using local-storage with [independent-installation](../02installation/02install.md). This is for developing or test, and will deploy local-storage from github repo. In this case, you should install the Local Disk Manager firstly.

## Roadmap

[Roadmap](https://github.com/hwameistor/local-storage/blob/main/doc/roadmap.md) provides a release plan about local storage and its features.