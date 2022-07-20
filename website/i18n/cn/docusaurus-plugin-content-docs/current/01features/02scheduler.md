---
sidebar_position: 3
sidebar_label: "调度器"
---

# 调度器

调度器是 HwameiStor 的重要组件之一。它自动将 Pod 调度到配有 HwameiStor 存储卷的正确节点。使用调度器后，Pod 不必再使用 NodeAffinity 或 NodeSelector 字段来选择节点。调度器能处理 LVM 和 Disk 存储卷。

## 安装

调度器应在集群中以 HA 模式部署，这是生产环境中的最佳实践。

## 通过 Helm Chart 部署

调度器必须与本地磁盘和本地磁盘管理器配合使用。建议通过 [Helm Chart 进行安装](../02installation/01helm-chart.md)。

## 通过 YAML 部署（针对开发）

```bash
$ kubectl apply -f deploy/scheduler.yaml
```
