---
sidebar_position: 3
sidebar_label: "PVC 自动扩容"
---

# PVC 自动扩容

组件 hwameistor-pvc-autoresizer 提供了 PVC 自动扩容的能力。扩容行为是通过 `ResizePolicy` 这个 CRD 来控制的。

## ResizePolicy

下面是一个示例 CR:

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: ResizePolicy
metadata:
  name: resizepolicy1
spec:
  warningThreshold: 60
  resizeThreshold: 80
  nodePoolUsageLimit: 90
```

`warningThreshold`、`resizeThreshold`、`resizeThreshold` 三个 int 类型的字段都表示一个百分比。

- `warningThreshold` 目前暂时还没有关联任何告警动作，它是作为一个目标比例，即扩容完成后卷的使用率会在这个比例以下。
- `resizeThreshold` 指示了一个使用率，当卷的使用率达到这个比例时，扩容动作就会被触发。
- `nodePoolUsageLimit` 表示节点存储池使用率的上限，如果某个池的使用率达到了这个比例，
  那么落在这个池的卷将不会自动扩容。

## 匹配规则

这是一个带有 label-selector 的示例 CR。

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: ResizePolicy
metadata:
  name: example-policy
spec:
  warningThreshold: 60
  resizeThreshold: 80
  nodePoolUsageLimit: 90
  storageClassSelector:
    matchLabels:
      pvc-resize: auto
  namespaceSelector:
    matchLabels:
      pvc-resize: auto
  pvcSelector:
    matchLabels:
      pvc-resize: auto
```

`ResizePolicy` 有三个 label-selector：

- `pvcSelector` 表示被这个 selector 选中的 PVC 会依照选中它的 policy 自动扩容。
- `namespaceSelector` 表示被这个 selector 选中的 namespace 下的 PVC 会依照这个 policy 自动扩容。
- `storageClassSelector` 表示从被这个 selector 选中的 storageclass 创建出来的 PVC 会依照这个 policy 自动扩容。

这三个 selector 之间是“且”的关系，如果你在一个 `ResizePolicy` 里指明了多个 selector，
那么要符合全部的 selector 的 PVC 才会匹配这个 policy。如果 `ResizePolicy` 中没有指明任何 selector，
它就是一个集群 `ResizePolicy`，也就像是整个集群中所有 PVC 的默认 policy。
