---
sidebar_position: 8
sidebar_label: "PVC 自动扩容"
---

# PVC 自动扩容

组件hwameistor-pvc-autoresizer提供了PVC自动扩容的能力。扩容行为是通过resizepolicy这个crd来控制的。

## ResizePolicy
下面是一个示例cr:
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

warningThreshold,resizeThreshold,resizeThreshold三个int类型的字段都表示一个百分比。resizeThreshold指示了一个使用率，当卷的使用率达到这个比例时，扩容动作就会被触发。warningThreshold目前暂时还没有关联任何告警动作，它是作为一个目标比例，即扩容完成后卷的使用率会在这个比例以下。nodePoolUsageLimit表示节点存储池使用率的上限，如果某个池的使用率达到了这个比例，那么落在这个池的卷将不会自动扩容。

## 匹配规则
这是一个带有label-selector的示例cr。

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

ResizePolicy有三个label-selector。pvcSelector表示被这个selector选中的pvc会依照选中它的policy自动扩容。namespaceSelector表示被这个selector选中的namespace下的pvc会依照这个policy自动扩容。storageClassSelector表示从被这个selector选中的storageclass创建出来的pvc会依照这个policy自动扩容。这三个selector之间是“且”的关系，如果你在一个resizepolicy里指明了多个selector，那么要符合全部的selector的pvc才会匹配这个policy。如果resizepolicy中没有指明任何selector，它就是一个集群resizepolicy，也就像是整个集群中所有pvc的默认policy。