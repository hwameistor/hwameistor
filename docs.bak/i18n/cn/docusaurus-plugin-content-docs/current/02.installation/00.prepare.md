---
sidebar_position: 1
sidebar_label: "准备工作"
---

# 准备工作

HwameiStor 本地磁盘是云原生本地存储系统，应部署在 Kubernetes 集群或以 Kubernetes 为内核的容器平台中，整个集群应满足下列条件：

- LocalDisk Version: `4.0+`
- Kubernetes Version: `1.18+`
- 节点
  - 空闲磁盘
  - LVM 逻辑存储卷（可选）
