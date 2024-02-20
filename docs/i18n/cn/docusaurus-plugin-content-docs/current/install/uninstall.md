---
sidebar_position: 5
sidebar_label: "卸载"
---

# 卸载 (仅用于测试环境)

为保证数据安全，强烈建议不要卸载生产环境的 HwameiStor 系统。本节介绍下列两种测试环境的卸载场景。

## 卸载但保留已有数据卷

如果想要卸载 HwameiStor 的系统组件，但是保留已经创建的数据卷并服务于数据应用，采用下列方式：

```console
$ kubectl get cluster.hwameistor.io
NAME             AGE
cluster-sample   21m

$ kubectl delete clusters.hwameistor.io hwameistor-cluster
```

最终，所有的 HwameiStor 系统组件（Pods）将被删除。用下列命令检查，结果为空。

```bash
kubectl -n hwameistor get pod
```

## 卸载并删除已有数据卷

:::danger
在卸载之前，请确认所有数据都可以被删除。
:::

如果想要卸载 HwameiStor 所有组件，并删除所有数据卷及数据，采用下列方式：

1. 清理有状态数据应用。

   1. 删除应用。

   2. 删除数据卷 PVC。

      相关的 PV、LV、LVR、LVG 都将被删除.

2. 清理 HwameiStor 系统组件。

   1. 删除 HwameiStor 组件。

      ```bash
      kubectl delete clusters.hwameistor.io hwameistor-cluster
      ```
      
   2. 删除 HwameiStor 系统空间。

      ```bash
      kubectl delete ns hwameistor
      ```

   3. 删除 CRD、Hook 以及 RBAC。

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

   4. 删除 StorageClass。

      ```bash
      kubectl get sc -o name \
        | grep hwameistor-storage-lvm- \
        | xargs -t kubectl delete
      ```

   5. 删除 hwameistor-operator。

      ```bash
      helm uninstall hwameistor-operator -n hwameistor
      ```

3. 最后，您仍然需要清理每个节点上的 LVM 配置，并采用额外的系统工具
  （例如 [wipefs](https://man7.org/linux/man-pages/man8/wipefs.8.html)）清除磁盘上的所有数据。

   ```bash
   wipefs -a /dev/sdx
   blkid /dev/sdx
   ```
