---
sidebar_position: 4
sidebar_label: "卸载"
---

# 卸载

本页讲述了卸载 HwameiStor 的 2 个方案。

## 若要保留数据卷

如果要在卸载 HwameiStor 时保留数据卷，请执行以下操作：

1. 清理 helm 实例

   1. 删除 HwameiStor helm 实例

      ```bash
      helm delete -n hwameistor hwameistor
      ```

   1. 删除 DRBD helm 实例

      ```bash
      helm delete -n hwameistor drbd-adapter
      ```

1. 清理 HwameiStor 组件

   1. 删除 hwameistor 命名空间

      ```bash
      kubectl delete ns hwameistor
      ```

   1. 删除 LocalVolumeGroup 实例

      ```bash
      kubectl delete localvolumegroups.hwameistor.io --all
      ```

   1. 删除 CRD、Hook 和 RBAC

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

## 若要删除数据卷

:::danger
执行卸载操作之前，请务必确认您确实要删除数据。
:::

如果你确认要在卸载 HwameiStor 时删除数据卷，请执行以下步骤：

1. 清理有状态应用

   1. 删除有状态应用

   1. 删除 PVC

      相关的 PV、LV、LVR、LVG 也会被删除。

1. 清理 helm 实例

   1. 删除 HwameiStor helm 实例

      ```bash
      helm delete -n hwameistor hwameistor
      ```

   1. 删除 DRBD helm 实例

      ```bash
      helm delete -n hwameistor drbd-adapter
      ```

1. 清理 HwameiStor 组件

   1. 删除 hwameistor 命名空间

      ```bash
      kubectl delete ns hwameistor
      ```

   1. 删除 CRD、Hook 和 RBAC

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

   1. 删除 StorageClass

      ```bash
      kubectl get sc -o name \
        | grep hwameistor-storage-lvm- \
        | xargs -t kubectl delete
      ```
