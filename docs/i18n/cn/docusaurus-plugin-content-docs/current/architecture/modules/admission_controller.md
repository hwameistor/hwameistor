---
sidebar_position: 5
sidebar_label: "准入控制器"
---

# 准入控制器

准入控制器是一种 Webhook，可以自动验证 HwameiStor 数据卷，协助将 `schedulerName` 修改为 hwameistor-scheduler。具体信息，请参见 [K8S 动态准入控制](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/)。

## 识别 HwameiStor 数据卷

准入控制器可以获取 Pod 使用的所有 PVC，并检查每个 PVC [存储制备器](https://kubernetes.io/zh-cn/docs/concepts/storage/storage-classes/#provisioner)。如果制备器的名称后缀是 `*.hwameistor.io`，表示 Pod 正在使用 HwameiStor 提供的数据卷。

## 验证资源

准入控制器只验证 `POD` 资源，并在创建资源时就进行验证。

:::info
为确保 HwameiStor 的 Pod 可以顺利启动，不会校验 HwameiStor 所在的命名空间下的 Pod。
:::
