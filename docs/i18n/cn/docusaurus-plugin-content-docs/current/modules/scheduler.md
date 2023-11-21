---
sidebar_position: 4
sidebar_label: "调度器"
---

# 调度器

调度器 (scheduler) 是 HwameiStor 的重要组件之一。它自动将 Pod 调度到配有 HwameiStor 存储卷的正确节点。使用调度器后，Pod 不必再使用 NodeAffinity 或 NodeSelector 字段来选择节点。调度器能处理 LVM 和 Disk 存储卷。

## 安装

调度器应在集群中以 HA 模式部署，这是生产环境中的最佳实践。
