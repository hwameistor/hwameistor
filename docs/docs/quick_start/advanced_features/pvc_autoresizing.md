---
sidebar_position: 8
sidebar_label: "PVC Autoresizing"
---

# PVC Autoresizing

The component hwameistor-pvc-autoresizer provide the ability to autoresize pvc.The resize behavior is controlled by the crd resizepolicy.

## ResizePolicy
A example cr is as below:
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

The warningThreshold,resizeThreshold,resizeThreshold three fields of type int all present a percentage.The resizeThreshold specifies the usage percentage of volume beyond which the resizing action will be triggered. The field warningThreshold is not relative with any warning action temporarily, it's just for the resize action that volume usage percentage is expected to under the warningThreshold after resize action done. The field nodePoolUsageLimit means a limit of pool usage percentage of localstoragenode, when a pool usage percentage reached this limit, the autoresizer will not do autoresizing for a volume located on the localstoragenode. 

## Match Rules
This is a examle cr with label selectors.

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

ResizePolicy have three label selectors. pvcSelector means pvc selected by the selector of the resizepolicy will autoresize according to the policy that select it. namespaceSelector means pvc in the namespace selected by the policy will autoresize according to the policy. storageClassSelector means pvc created from storageclass selected by the policy will autoresize according to the policy. It's a "&&" relation between the three selectors, if you specify more than one selector in a resizepolicy, the pvc match all the selector you specified will match this policy. If no selector specified in the resizepolicy, it's a cluster resizepolicy, it's like default policy for all pvc in the cluster.