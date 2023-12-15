---
sidebar_position: 9
sidebar_label: "LDA 控制器"
---

# LDA 控制器

LDA 控制器提供了一个单独的 CRD - `localdiskactions`，用于匹配 localdisk，并执行指定的 action。
其代码示例如下：

```yaml
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskAction
metadata:
  name: forbidden-1
spec:
  action: reserve
  rule:
    minCapacity: 1024
    devicePath: /dev/rbd*

---
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskAction
metadata:
  name: forbidden-2
spec:
  action: reserve
  rule:
    maxCapacity: 1048576
    devicePath: /dev/sd*
```

以上的 yaml 表示：

1. 比 1024 字节更小且 devicePath 满足 `/dev/rbd*` 匹配条件的 Localdisk 将被预留
2. 比 1048576 字节更大且 devicePath 满足 `/dev/sd*` 匹配条件的 Localdisk 将被预留

请注意，当前支持的 action **仅 reserve**。
