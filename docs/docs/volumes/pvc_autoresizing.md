---
sidebar_position: 3
sidebar_label: "PVC Autoresizing"
---

# PVC Autoresizing

The component "hwameistor-pvc-autoresizer" provides the ability to automatically resize Persistent Volume Claims (PVCs).
The resizing behavior is controlled by the `ResizePolicy` custom resource definition (CRD).

## ResizePolicy

An example of CR is as below:

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

The three fields `warningThreshold`, `resizeThreshold`, and `nodePoolUsageLimit` are all of type integer and represent percentages.

- `warningThreshold` currently does not have any associated alert actions. It serves as a
  target ratio, indicating that the usage rate of the volume will be below this percentage
  after resizing is completed.
- `resizeThreshold` indicates a usage rate at which the resizing action will be triggered
  when the volume's usage rate reaches this percentage.
- `nodePoolUsageLimit` represents the upper limit of storage pool usage on a node. If the
  usage rate of a pool reaches this percentage, volumes assigned to that pool will not automatically resize.

## Match Rules

This is an examle of CR with label selectors.

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

The `ResizePolicy` has three label selectors:

- `pvcSelector` indicates that PVCs selected by this selector will automatically resize according
  to the policy that selected them.
- `namespaceSelector` indicates that PVCs under namespaces selected by this selector will
  automatically resize according to this policy.
- `storageClassSelector` indicates that PVCs created from storage classes selected by this
  selector will automatically resize according to this policy.

These three selectors have an "AND" relationship. If you specify multiple selectors in a `ResizePolicy`,
the PVCs must match all of the selectors in order to be associated with that policy. If no selectors are
specified in the `ResizePolicy`, it becomes a cluster-wide `ResizePolicy`, acting as the default policy
for all PVCs in the entire cluster.
