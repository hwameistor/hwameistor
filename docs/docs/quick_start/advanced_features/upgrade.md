---
sidebar_position: 4
sidebar_label: "Upgrade"
---

# Upgrade

Helm makes upgrading HwameiStor very easy:

```console
$ helm upgrade -n hwameistor hwameistor -f new.values.yaml
```

The upgrade will restart each HwameiStor pod in a rolling fashion.

:::caution
The volumes will continue to serve pods uninterrupted during a HwameiStor upgrade.
:::
