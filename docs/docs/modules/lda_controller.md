---
sidebar_position: 9
sidebar_label: "LDA Controller"
---

# LDA Controller

The LDA controller provides a separate CRD - `LocalDiskAction`, which is used to match
the localdisk and execute the specified action. Its yaml definition is as follows:

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

The above yaml indicates:

1. Localdisks smaller than 1024 bytes and whose devicePath meets the `/dev/rbd*` matching condition will be reserved
2. Localdisks larger than 1048576 bytes and whose devicePath meets the `/dev/sd*` matching condition will be reserved

Note that the currently supported actions are only **reserve**.
