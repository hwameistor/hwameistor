# HwameiStor

[English](./README.md) | 简体中文

HwameiStor 是一款 Kubernetes 原生的容器附加存储 (CAS) 解决方案，将 HDD、SSD 和 NVMe
磁盘形成本地存储资源池进行统一管理，使用 CSI 架构提供分布式的本地数据卷服务，为有状态的云原生应用或组件提供数据持久化能力。

![System architecture](./docs/i18n/cn/docusaurus-plugin-content-docs/current/img/architecture.png)

## 目前状态

<img src="https://github.com/cncf/artwork/blob/master/other/illustrations/ashley-mcnamara/transparent/cncf-cloud-gophers-transparent.png" style="width:600px;" />

**HwameiStor 是一个[云原生计算基金会 (CNCF)](https://cncf.io/) 沙箱孵化项目。**

HwameiStor 的最新版本为 [![hwameistor-releases](https://img.shields.io/github/v/release/hwameistor/hwameistor.svg?include_prereleases)](https://github.com/hwameistor/hwameistor/releases)

## 构建状态

![period-check](https://github.com/hwameistor/hwameistor/actions/workflows/period-check.yml/badge.svg) [![codecov](https://codecov.io/gh/hwameistor/hwameistor/branch/main/graph/badge.svg?token=AWRUI46FEX)](https://codecov.io/gh/hwameistor/hwameistor) [![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/5685/badge)](https://bestpractices.coreinfrastructure.org/projects/5685)

## 发版状态

参阅[当前发行版](https://github.com/hwameistor/hwameistor/releases)。

## 运行环境

### Kubernetes 兼容性

| Kubernetes     | v0.4.3 | >=v0.5.0 |
| -------------- | ------ | -------- |
| >=1.18&&<=1.20 | 是     | 否       |
| 1.21           | 是     | 是       |
| 1.22           | 是     | 是       |
| 1.23           | 是     | 是       |
| 1.24           | 是     | 是       |
| 1.25           | No     | 是       |

## 模块和代码

HwameiStor 包含若干模块：

- [local-disk-manager](#local-disk-manager)
- [local-storage](#local-storage)
- [scheduler](#scheduler)
- [admission-controller](#admission-controller)
- [Evictor](#evictor)
- [Exporter](#exporter)
  [HA module installer](#高可用模块安装器)

### local-disk-manager

local-disk-manager（LDM）旨在管理节点上的磁盘。
像 local-storage 等其他模块可以利用 LDM 提供的磁盘管理功能。
[了解更多](docs/docs/architecture/modules/ldm.md)

### local-storage

local-storage（LS）提供了一个云原生的本地存储系统。
它旨在为应用程序提供具有本地访问权限的高性能持久 LVM 卷。
[了解更多](docs/docs/architecture/modules/ls.md)

### Scheduler

Scheduler 可以自动将 Pod 调度到具有相关 HwameiStor 卷的正确节点。
[了解更多](docs/docs/architecture/modules/scheduler.md)

### admission-controller

admission-controller 是一种 Webhook，可以自动确定哪个 Pod 使用 HwameiStor 卷，
并帮助修改 schedulerName 为 hwameistor-scheduler。
[了解更多](docs/docs/architecture/modules/admission_controller.md)

### Evictor

驱逐器（Evictor）用于在节点或 Pod 驱逐的情况下自动迁移 HwameiStor 卷。
当按计划或未按计划驱逐节点或 Pod 时，将自动检测并从节点迁移具有副本的关联 HwameiStor 卷。
[了解更多](docs/docs/architecture/modules/evictor.md)

## 高可用模块安装器

DRBD（Distributed Replicated Block Device）是 HwameiStor 将利用的第三方高可用模块之一，用于提供高可用卷。
它由 Linux 内核模块和相关脚本组成，用于构建高可用集群。通过在网络上镜像整个设备来实现，可以看作是一种网络 RAID。
这个安装器可以直接将 DRBD 安装到容器集群中。目前，此模块仅用于测试目的。
[了解更多](docs/docs/architecture/modules/drbd.md)

## Exporter

Exporter 将收集系统指标，包括节点、存储池、卷、磁盘等。支持 Prometheus。

## 文档

有关完整文档，请参阅文档站 [hwameistor.io](https://hwameistor.io/docs/intro)。

有关已在生产环境或用户验收测试环境中部署了 HwameiStor 的详细用户清单，请查阅 [HwameiStor 用户列表](./adopters.md)。

## Roadmap

| 特性                | 状态     | 版本   | 说明                                 |
| ------------------- | -------- | ------ | ------------------------------------ |
| LVM 卷 CSI          | 已完成   | v0.3.2 | 用 lvm 制备卷                        |
| 磁盘卷 CSI          | 已完成   | v0.3.2 | 用磁盘制备卷                         |
| HA LVM 卷           | 已完成   | v0.3.2 | 高可用卷                             |
| LVM 卷扩容          | 已完成   | v0.3.2 | 在线扩展 LVM 容量                    |
| LVM 卷转换          | 已完成   | v0.3.2 | 将 LVM 卷转换为高可用卷              |
| LVM 卷迁移          | 已完成   | v0.4.0 | 将 LVM 卷副本迁移到不同的节点        |
| 卷组 (VG)           | 已完成   | v0.3.2 | 支持卷组分配                         |
| 磁盘健康检查        | 已完成   | v0.7.0 | 磁盘故障预测、状态报告               |
| LVM 高可用卷恢复    | 计划中   |        | 恢复有故障的 LVM HA 卷               |
| HwameiStor Operator | 已完成   | v0.9.0 | HwameiStor Operator 用于安装和维护等 |
| 可观测性            | 已完成   | v0.9.2 | 支持指标、日志等可观测性             |
| 故障转移            | 计划中   |        | HwameiStor 卷对 Pod 进行故障转移     |
| IO 节流             | 已完成   | v0.11.0 | 限制访问 HwameiStor 卷的 IO 带宽    |
| 换盘                | 计划中   |        | 更换故障或即将故障的磁盘             |
| LVM 卷自动扩容      | 正在开发 |        | 自动扩展 LVM 卷                      |
| LVM 卷快照          | 正在开发 |        | LVM 卷快照                           |
| LVM 卷克隆          | 计划中   |        | 克隆 LVM 卷                          |
| LVM 卷薄制备        | 还未计划 |        | LVM 卷薄制备                         |
| LVM 卷条带模式      | 还未计划 |        | LVM 卷条带读写                       |
| 数据加密            | 还未计划 |        | 数据加密                             |
| 系统一致性          | 计划中   |        | 一致性检查和灾难恢复                 |
| 卷备份              | 计划中   |        | 将卷数据备份到远程服务器并恢复       |

## 社区

我们欢迎任何形式的贡献。如果您有任何有关贡献方面的疑问，请参阅[贡献指南](./CONTRIBUTING.md)。

### 博客

请关注我们的[每周博客](https://hwameistor.io/blog)。

### Slack

如果你想加入我们在 CNCF Slack 的 hwameistor 频道，**请先[接受 CNCF Slack 邀请](https://slack.cncf.io/)**，然后加入 [#hwameistor](https://cloud-native.slack.com/messages/hwameistor)。

### 微信

HwameiStor 技术沟通群：

![扫描二维码入群](./docs/docs/img/wechat.png)

## 讨论

欢迎在[此处](https://github.com/hwameistor/hwameistor/discussions)查阅 Roadmap 相关的讨论。

## PR 和 Issue

欢迎与开发团队聊天沟通，也非常欢迎一切形式的 PR 和 Issue。

我们将尽力回应在社区报告的每个问题，但我们会首先解决在[此 repo 中报告的](https://github.com/hwameistor/hwameistor/discussions)问题。

## 许可证

版权所有 (c) 2014-2023 HwameiStor 开发团队

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
<http://www.apache.org/licenses/LICENSE-2.0>

<p align="center">
<img src="https://landscape.cncf.io/images/left-logo.svg" width="300"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="350"/>
<br/><br/>
HwameiStor 位列 <a href="https://landscape.cncf.io/?selected=hwamei-stor">CNCF 云原生全景图。</a>
</p>
