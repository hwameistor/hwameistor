# HwameiStor

English | [简体中文](./README_zh.md)

HwameiStor is an HA local storage system for cloud-native stateful workloads.
It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe.
It uses the CSI architecture to provide distributed services with local volumes and provides data
persistence capabilities for stateful cloud-native workloads or components.

![System architecture](docs/docs/img/architecture.png)

## Current Status

<img src="https://github.com/cncf/artwork/blob/master/other/illustrations/ashley-mcnamara/transparent/cncf-cloud-gophers-transparent.png" style="width:600px;" />

**HwameiStor is a [Cloud Native Computing Foundation](https://cncf.io/) sandbox project.**

The latest release of HwameiStor is [![hwameistor-releases](https://img.shields.io/github/v/release/hwameistor/hwameistor.svg?include_prereleases)](https://github.com/hwameistor/hwameistor/releases)

## Build Status

![period-check](https://github.com/hwameistor/hwameistor/actions/workflows/period-check.yml/badge.svg) [![codecov](https://codecov.io/gh/hwameistor/hwameistor/branch/main/graph/badge.svg?token=AWRUI46FEX)](https://codecov.io/gh/hwameistor/hwameistor) [![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/5685/badge)](https://bestpractices.coreinfrastructure.org/projects/5685)

## Release Status

See [current releases](https://github.com/hwameistor/hwameistor/releases).

## Running Environments

### Kubernetes compatiblity

| kubernetes     | v0.4.3 | >=v0.5.0 |
| -------------- | ------ | -------- |
| >=1.18&&<=1.20 | Yes    | No       |
| 1.21           | Yes    | Yes      |
| 1.22           | Yes    | Yes      |
| 1.23           | Yes    | Yes      |
| 1.24           | Yes    | Yes      |
| 1.25           | No     | Yes      |

## Modules and Code

HwameiStor contains several modules:

 * [local-disk-manager](#local-disk-manager)
 * [local-storage](#local-storage)
 * [scheduler](#scheduler)
 * [admission-controller](#admission-controller)
 * [Evictor](#evictor)
 * [Exporter](#exporter)
 * [HA module installer](#ha-module-installer)
 * [Volume Snapshot](#volume-snapshot)
 * [Volume Auto Resize](#volume-auto-resize)
 * [Volume IO Throtting](#volume-io-throtting)
 * [App Failover](#app-failover)
 * [Audit](#audit)
 * [UI](#ui)

### local-disk-manager

local-disk-manager (LDM) is designed to hold the management of disks on nodes.
Other modules such as local-storage can take advantage of the disk management feature provided by LDM.
[Learn more](docs/docs/architecture/modules/ldm.md)

### local-storage

local-storage (LS) provides a cloud-native local storage system.
It aims to provision high-performance persistent LVM volume with local access to applications.
[Learn more](docs/docs/architecture/modules/ls.md)

### Scheduler

Scheduler is to automatically schedule a pod to a correct node which has the associated HwameiStor volumes.
[Learn more](docs/docs/architecture/modules/scheduler.md)

### admission-controller

admission-controller is a webhook that can automatically determine which pod uses the HwameiStor volume and,
help to modify the schedulerName to hwameistor-scheduler.
[Learn more](docs/docs/architecture/modules/admission_controller.md)

### Evictor

Evictor is used to automatically migrate HwameiStor volumes in case of node or pod eviction.
When a node or pod is evicted as either Planned or Unplanned, the associated HwameiStor volumes,
which have a replica on the node, will be detected and migrated out this node automatically.
[Learn more](docs/docs/architecture/modules/evictor.md)

### HA module installer

DRBD (Distributed Replicated Block Device) is one of third-party HA modules which the HwameiStor will leverage to provide HA volume.
It composed of Linux kernel modules and related scripts
to build high available clusters. It is implemented by mirroring the entire device over the network,
which can be thought of as a kind of network RAID. This installer can directly install DRBD to a
container cluster. Currently this module is only for testing purposes.
[Learn more](docs/docs/architecture/modules/drbd.md)

### Exporter

Exporter will collect the system metrics including nodes, storage pools, volumes, disks. It supports Prometheus.

### Volume Snapshot

HwameiStor provides the feature of snapshot and the restore on the LVM volumes.
Currently, the snapshot/restore feature works for LVM non-HA volume.

### Volume Auto Resize

HwameiStor can automatically expand the LVM volume according the pre-defined resize policy.
User can define the preferred policy and describe how and when to expand the volume, and HwameiStor will take the policy into effect.

### Volume IO Throtting

HwameiStor can set a maxmium rate (e.g. bandwidth, IOPS) to access a volume.
This feature is very important to prevent the Pod from crashing, especially in the low-resource condition. 

### App Failover

The feature of failover is to actively help the application to fail over to another health node with the volume replica, and continue the working.

### Audit

HwameiStor provides the information about the resource history, including cluster, node, storage pool, volume, etc..

### UI

HwameiStor provides a friendly UI to the user to operate the cluster.

## Documentation

For full documentation, please see our website [hwameistor.io](https://hwameistor.io/docs/intro).

For detailed adopters that have HwameiStor deployed in a production environment or a user acceptance testing environment,
please check the [adopters list](./adopters.md).

## Roadmap

| Features                  | Status    | Release | Description                                       |
| ------------------------- |-----------|---------| ------------------------------------------------- |
| CSI for LVM volume        | Completed | v0.3.2  | Provision volume with lvm                         |
| CSI for disk volume       | Completed | v0.3.2  | Provision volume with disk                        |
| HA LVM Volume             | Completed | v0.3.2  | Volume with HA                                    |
| LVM Volume expansion      | Completed | v0.3.2  | Expand LVM volume capacity online                 |
| LVM Volume conversion     | Completed | v0.3.2  | Convert a non-HA LVM volume to the HA             |
| LVM Volume migration      | Completed | v0.4.0  | Migrate a LVM volume replica to a different node  |
| Volume Group              | Completed | v0.3.2  | Support volume group allocation                   |
| Disk health check         | Completed | v0.7.0  | Disk fault prediction, status reporting           |  
| LVM HA Volume Recovery    | Planed    |         | Recover the LVM HA volume in problem              |
| HwameiStor Operator       | Completed | v0.9.0  | Operator for HwameiStor install, maintain, etc.   |
| Observability             | Completed | v0.9.2  | Observability, such as metrics, logs, etc.        |
| Failover                  | Completed | v0.12.0 | Fail over the pod with HwameiStor volume          |
| IO throttling             | Completed | v0.11.0 | Limit IO bandwith to access the HwameiStor volume |
| Disk replacement          | Planed    |         | Replace disk which fails or will fail soon        |
| LVM volume auto-expansion | Completed | v0.12.0 | Expand LVM volume automatically                   |
| LVM volume snapshot       | Completed | v0.12.0 | Snapshot of LVM volume                            |
| LVM volume clone          | Planed    |         | Clone LVM volume                                  |
| LVM volume thin provision | Unplaned  |         | LVM volume thin provision                         |
| LVM volume stripe mode    | Unplaned  |         | LVM volume stripe read/write                      |
| Data encryption           | Unplaned  |         | Data encryption                                   |
| System Consistency        | Planed    |         | Consistent check and recovery from a disaster     |
| Volume backup             | Planed    |         | Backup the volume data to remote server and restore  |
| HwameiStor CLI command    | In Progress  |      | CLI command is to manage the HwameiStor cluster  |
| HwameiStor GUI            | Completed | v0.11.0 | Manage the HwameiStor cluster  |


## Community

We welcome contributions of any kind.
If you have any questions about contributing, please consult the [contributing documentation](./docs/docs/contribute/CONTRIBUTING.md).

### Blog

Please follow our [weekly blogs](https://hwameistor.io/blog).

### Slack

If you want to join the hwameistor channel on CNCF slack, please **[get invite to CNCF slack](https://slack.cncf.io/)** 
and then join the [#hwameistor](https://cloud-native.slack.com/messages/hwameistor) channel.

### WeChat

HwameiStor tech-talk group:

![QR code for Wechat](./docs/docs/img/wechat.png)

## Discussion

Welcome to follow our roadmap discussions [here](https://github.com/hwameistor/hwameistor/discussions)

## Pull Requests and Issues

Please feel free to raise requests on chats or by a PR.

We will try our best to respond to every issue reported on community channels,
but the issues reported [here](https://github.com/hwameistor/hwameistor/discussions)
on this repo will be addressed first.

## License

Copyright (c) 2014-2023 The HwameiStor Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
<http://www.apache.org/licenses/LICENSE-2.0>

<p align="center">
<img src="https://landscape.cncf.io/images/left-logo.svg" width="300"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="350"/>
<br/><br/>
HwameiStor enriches the <a href="https://landscape.cncf.io/?selected=hwamei-stor">CNCF CLOUD NATIVE Landscape.</a>
</p>
