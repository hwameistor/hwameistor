---
sidebar_position: 4
sidebar_label: "升级"
---

# 升级

Helm 让 HwameiStor 的升级变得非常简单。只需运行以下命令：

```console
$ helm upgrade -n hwameistor hwameistor -f new.values.yaml
```

升级过程中将以滚动的方式重启每个 HwameiStor Pod。

:::caution
在升级 HwameiStor 期间，这些卷将继续不间断地为 Pod 服务。
:::
