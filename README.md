# Local Storage Module

[![codecov](https://codecov.io/gh/hwameistor/local-storage/branch/main/graph/badge.svg?token=AWRUI46FEX)](https://codecov.io/gh/hwameistor/local-storage)

English | [Simplified_Chinese](https://github.com/hwameistor/local-storage/blob/main/README_zh.md)

## Introduction

Local Storage is one of modules of HwameiStor which is a cloud native local storage system. it aims to provision high performance persistent LVM volume with local access to applications.

Support local volume kinds: `LVM`.

Support disk types: `HDD`, `SSD`, `NVMe`.

## Architecture of HwameiStor

![image](https://github.com/hwameistor/local-storage/blob/main/doc/design/HwameiStor-arch.png)

## Features and Roadmap

This [Roadmap](https://github.com/hwameistor/local-storage/blob/main/doc/roadmap.md) provides information about local-storage releases, as well as feature tracking.

## Use Cases

HwameiStor provisions two kind of local volumes, LVM and Disk. As a component of HwameiStor, Local Storage provisions two types of local LVM volumes, HA and non-HA.

For the non-HA local LVM volume, it's the best data persistent solution for the following use cases:

* ***Database*** with HA capability, such as MongoDB, etc..
* ***Messaging system*** with HA capability, such as Kafka, RabbitMQ, etc..
* ***Key-value store*** with HA capability, such as Redis, etc..
* Others with HA capability

For the HA local LVM volume, it's the best data persistent solution for the following use cases:

* ***Database***, such as MySQL, PostgreSQL etc..
* Other applications which requires the data HA feature.

## Usage With Helm Chart

Local Storage must work with Local Disk Manager module. It's suggested to install by [helm-charts](https://github.com/hwameistor/helm-charts/blob/main/README.md)

## Usage With Independent Installation

Developer can start using local-storage With [independent-installation](https://github.com/hwameistor/local-storage/blob/main/doc/installation.md)， This is for developing or test, and will deploy local-storage from github repo. Please install the Local Disk Manager firstly.

## Glossary

* ***LocalDisk*** LDM abstracts disk resources into objects in k8s. A LocalDisk (LD) resource object represents the disk resources on the host..
* ***LocalDiskClaim*** The way to use disk, users can add a description of the disk to select the disk to be used..
* ***LocalVolume*** LocalVolume is a logical concept in the system, with control node management..
* ***LocalVolumeReplica*** The LocalVolumeReplica resource is created by a control node when creating or updating the Volume.The LocalVolumeReplica is assigned to the specified node that creates / manages the local storage (e.g. LV) based on its content, and maintains it in real-time..
* ***LocalStorageNode*** Each node should create its own Node CRD resource and proactively maintain and update relevant information..

## Feedbacks

Please submit any feedback and issue at: [Issues](https://github.com/hwameistor/local-storage/issues)
