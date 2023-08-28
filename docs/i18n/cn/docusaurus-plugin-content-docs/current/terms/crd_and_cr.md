---
sidebar_position: 4
sidebar_label: "CRD 和 CR"
---

# CRD 和 CR

## CRD

`CRD` 是 `Custom Resource Definition` 的缩写，是 `Kubernetes` 内置原生的一个资源类型。它是自定义资源 (CR) 的定义，用来描述什么是自定义资源。

CRD 可以向 `Kubernetes` 集群注册一种新资源，用于拓展 `Kubernetes` 集群的能力。有了 `CRD`，就可以自定义底层基础设施的抽象，根据业务需求来定制资源类型，利用 `Kubernetes` 已有的资源和能力，通过乐高积木的模式定义出更高层次的抽象。

## CR

`CR` 是 `Custom Resource` 的缩写，它实际是 `CRD` 的一个实例，是符合 `CRD` 中字段格式定义的一个资源描述。

## CRDs + Controllers

我们都知道 `Kubernetes` 的扩展能力很强大，但仅有 `CRD` 并没有什么用，还需要有控制器 (`Custom Controller`) 的支撑，才能体现出 `CRD` 的价值，`Custom Controller` 可以监听 `CR` 的 `CRUD` 事件来实现自定义业务逻辑。

在 `Kubernetes` 中，可以说是 `CRDs + Controllers = Everything`。

另请参考 Kubernetes 官方文档：

- [CustomResource](https://kubernetes.io/zh-cn/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [CustomResourceDefinition](https://kubernetes.io/zh-cn/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)