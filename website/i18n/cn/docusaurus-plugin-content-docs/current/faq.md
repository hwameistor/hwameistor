---
sidebar_position: 6
sidebar_label: "常见问题"
---

# 常见问题

### 问题 1：HwameiStor 本地存储调度器 scheduler 在 kubernetes 平台中是如何工作的？ 

HwameiStor 的调度器是以 pod 的形式部署在 HwameiStor 的命名空间。

![img](images/clip_image002.png)

当应用（Deployment 或 StatefulSet ）被创建后，应用的 Pod 会被自动部署到已配置好具备 HwameiStor 本地存储能力的 worker 节点上。

### 问题 2: HwameiStor 如何应对应用多副本工作负载的调度？与传统通用型共享存储有什么不同？

HwameiStor 建议使用有状态的 StatefulSet 用于多副本的工作负载。

有状态应用 StatefulSet 会将复制的副本部署到同一 worker 节点，但会为每一个 Pod 副本创建一个对应的 pv 数据卷。如果需要部署到不同节点分散 workload，需要通过 pod affinity 手动配置。

![img](images/clip_image004.png)

由于无状态应用 deployment 不能共享 block 数据卷，所以建议使用单副本。

对于传统通用型共享存储：

有状态应用 statefulSet 会将复制的副本优先部署到其他节点以分散 workload，但会为每一个 Pod 副本创建一个对应的 pv 数据卷。只有当副本数超过 worker 节点数的时候会出现多个副本在同一个节点。

无状态应用 deployment 会将复制的副本优先部署到其他节点以分散 workload，并且所有的 pod 共享一个 pv 数据卷（目前仅支持 NFS）。只有当副本数超过 worker 节点数的时候会出现多个副本在同一个节点。对于 block 存储，由于数据卷不能共享，所以建议使用单副本。