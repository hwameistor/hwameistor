---
sidebar_position: 5
sidebar_label: "卷"
---

# 卷

容器中的文件在磁盘上是临时存放的，这给容器中运行的较重要的应用程序带来一些问题。

- 问题一：当容器崩溃时文件会丢失。kubelet 会重新启动容器，但容器会以干净的状态重启。
- 问题二：在同一个 `Pod` 中运行多个容器并共享文件时出现此问题。

Kubernetes 卷 (Volume) 这一抽象概念能够解决这两个问题。

Kubernetes 支持很多类型的卷。Pod 可以同时使用任意数目的卷类型。临时卷类型的生命周期与 Pod 相同，但持久卷可以比 Pod 的存活期长。
当 Pod 不再存在时，Kubernetes 也会销毁临时卷；不过 Kubernetes 不会销毁持久卷。对于给定 Pod 中任何类型的卷，在容器重启期间数据都不会丢失。

卷的核心是一个目录，其中可能保存了数据，Pod 中的容器可以访问该目录中的数据。所采用的特定卷类型将决定该目录如何形成、使用何种介质保存数据以及目录中存放的内容。

使用卷时，在 `.spec.volumes` 字段中设置为 Pod 提供的卷，并在 `.spec.containers[*].volumeMounts` 字段中声明卷在容器中的挂载位置。

参考 Kubernetes 官方文档：

- [卷](https://kubernetes.io/zh-cn/docs/concepts/storage/volumes/)
- [持久卷](https://kubernetes.io/zh-cn/docs/concepts/storage/persistent-volumes/)
- [临时卷](https://kubernetes.io/zh-cn/docs/concepts/storage/ephemeral-volumes/)