# Local Storage

[![codecov](https://codecov.io/gh/hwameistor/local-storage/branch/main/graph/badge.svg?token=AWRUI46FEX)](https://codecov.io/gh/hwameistor/local-storage)

简体中文 | [英文](https://github.com/hwameistor/local-storage/blob/main/README.md)

## 介绍

Local Storage是云原生本地存储系统HwameiStor的一个模块。它旨在为应用提供高性能的本地持久化LVM存储卷。

目前支持的本地持久化数据卷类型: `LVM`.

目前支持的本地磁盘类型: `HDD`, `SSD`, `NVMe`.

## HwameiStor系统架构图

![image](https://github.com/hwameistor/local-storage/blob/main/doc/design/HwameiStor-arch.png)

## 功能与路线图

该[功能路线图](https://github.com/hwameistor/local-storage/blob/main/doc/roadmap_zh.md) 提供了Local Storage版本发布及特性追踪功能

## 适用场景

HwameiStor提供两种本地数据卷：LVM、Disk。Local Storage作为HwameiStor的一部分，负责提供LVM本地数据卷，其中包括：高可用LVM数据卷、非高可用LVM数据卷。

非高可用的LVM本地数据卷，适用下列场景和应用：

* 具备高可用特性的 ***数据库***。例如： MongoDB，等等
* 具备高可用特性的 ***消息中间件***。例如： Kafka，RabbitMQ，等等
* 具备高可用特性的 ***键值存储系统***。例如： Redis，等等
* 其他具备高可用功能的应用

高可用的LVM本地数据卷，适用下列场景和应用：

* ***数据库***。例如： MySQL、PostgreSQL等
* 其他需要数据高可用特性的应用

## 使用Helm Chart安装部署

Local Storage是HwameiStor的一部分，必须和Local Disk Manager一起工作。建议用户通过[helm-charts](https://github.com/hwameistor/helm-charts/blob/main/README.md) 安装部署。

## 独立安装部署使用方式

开发者可以通过 [独立安装](https://github.com/hwameistor/local-storage/blob/main/doc/installation_zh.md)部署local-storage，这里介绍从源代码进行安装、使用。主要用于开发、测试。需要事先安装Local Disk Manager。

## 名词解释

* ***LocalDisk*** LDM 抽象的磁盘资源（缩写：LD），一个 LD 代表了节点上面的一块物理磁盘。
* ***LocalDiskClaim*** 系统使用磁盘的方式，通过创建 Claim 对象来向系统申请磁盘。
* ***LocalVolume*** LocalVolume在系统中是一个逻辑概念，有控制节点管理。控制节点接受外部请求（e.g. Kubernetes的PVC），根据当时的集群全局信息，创建LocalVolume以及LocalVolumeReplicas，并将LocalVolumeReplica分配给相应的节点
* ***LocalVolumeReplica*** LocalVolumeReplica资源是有控制节点在创建或更新Volume时创建的。LocalVolumeReplica是分配给指定节点的，并由该节点根据其内容进行创建/管理本地存储（e.g. LV），并实时进行维护。
* ***LocalStorageNode*** 每个节点都应该创建一个自己的Node CRD 资源，并主动维护、更新相关信息。

## 反馈

如果有任何问题、意见、建议，请反馈至：[Issues](https://github.com/hwameistor/local-storage/issues)
