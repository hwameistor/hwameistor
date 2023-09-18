---
sidebar_position: 8
sidebar_label: "LDA控制器"
---

# LDA控制器

LDA控制器提供了一个单独的CRD---localdiskactions，用于匹配localdisk，并执行指定的action。如下所示：

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

以上的yaml表示：
1. 比1024字节更小且devicePath满足/dev/rbd\*匹配条件的Localdisk将被预留
2. 比1048576字节更大且devicePath满足/dev/sd\*匹配条件的Localdisk将被预留

请注意，当前**支持的action仅reserve**
