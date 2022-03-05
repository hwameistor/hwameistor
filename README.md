# Local Storage System (local-storage)

English | [Simplified_Chinese](https://github.com/hwameistor/local-storage/blob/main/README_zh.md)

## Introduction

Local Storage System is a cloud native storage system. It manages the free disks of each node and provision high performance persistent volume with local access to application.

Support local volume kind: `LVM`, `Disk`, `RAMDisk`.

Support disk type: `HDD`, `SSD`, `NVMe`, `RAMDisk`.

## Software Structure

![image](https://github.com/hwameistor/local-storage/blob/main/doc/design/HwameiStor-arch.png)

## Features and Roadmap

This [Roadmap](https://github.com/hwameistor/local-storage/blob/main/doc/roadmap.md) provides information about local-storage releases, as well as feature tracking 

## Use Cases

Currently, local-storage offers high performance local volume without HA. It's one of best data persistent solution for the following use cases:

* ***Database*** with HA capability, such as MySQL, OceanBase, MongoDB, etc..
* ***Messaging system*** with HA capability, such as Kafka, RabbitMQ, etc..
* ***Key-value store*** with HA capability, such as Redis, etc..
* ***Distributed storage system***, such as MinIO, Ozone, etc..
* Others with HA capability

## Usage With Helm Chart

Users can start using local-storage With [helm-charts](https://github.com/hwameistor/helm-charts/blob/main/README.md)

## Usage With Independent Installation

Users can start using local-storage With [independent-installation](https://github.com/hwameistor/local-storage/blob/main/doc/installation.md)， This is for developing or test, and will deploy local-storage from github repo.

## Glossary

* ***LocalDisk*** LDM abstracts disk resources into objects in k8s. A LocalDisk (LD) resource object represents the disk resources on the host..
* ***LocalDiskClaim*** The way to use disk, users can add a description of the disk to select the disk to be used..
* ***LocalVolume*** LocalVolume is a logical concept in the system, with control node management..
* ***LocalVolumeReplica*** The LocalVolumeReplica resource is created by a control node when creating or updating the Volume.The LocalVolumeReplica is assigned to the specified node that creates / manages the local storage (e.g. LV) based on its content, and maintains it in real-time..
* ***LocalStorageNode*** Each node should create its own Node CRD resource and proactively maintain and update relevant information..

## Feedbacks

Please submit any feedback and issue at: [Issues](https://github.com/hwameistor/local-storage/issues)
