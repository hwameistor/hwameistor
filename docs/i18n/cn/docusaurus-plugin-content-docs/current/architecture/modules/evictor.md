---
sidebar_position: 6
sidebar_label: "驱逐器"
---

# 驱逐器

驱逐器是 HwameiStor 系统运维的重要组件，能够保障 HwameiStor 在生产环境中持续正常运行。
当系统节点或者应用 Pod 由于各种原因被驱逐时，驱逐器会自动发现节点或者 Pod 所关联的 HwameiStor 数据卷，并自动将其迁移到其他节点，从而保证被驱逐的 Pod 可以调度到其他节点上并正常运行。

在生产环境中，应该采用高可用模式部署驱逐器.

## 如何使用

请参考 [Eviction](../../quick_start/advanced_features/volume_eviction.md)

## 安装（Helm Chart)

驱逐器依赖 HwameiStor 的本地存储和本地磁盘管理器，建议通过 [Helm Chart 进行安装](../../quick_start/install/deploy.md) 进行安装.
