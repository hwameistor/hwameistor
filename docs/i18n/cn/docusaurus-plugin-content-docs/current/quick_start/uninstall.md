---
sidebar_position: 4
sidebar_label: "卸载"
---

# 卸载 (仅用于测试环境)

这部分介绍了两种卸载 HwameiStor 系统的方式。

## 卸载并保留已有数据卷

如果想要卸载 HwameiStor 的系统组件，但是保留已经创建的数据卷并服务于数据应用，采用下列方式：

```console
$ kubectl get cluster.hwameistor.io
NAME             AGE
cluster-sample   21m

$ kubectl delete cluster cluster-sample
```

最终，所有的 HwameiStor 系统组件（Pods）将被删除。用下列命令检查，结果为空

```console
$ kubectl -n hwameistor get pod
```

## 完全卸载

:::danger
在卸载之前，请确认所有数据都可以被删除
:::

如果想要卸载 HwameiStor 所有组件，并删除所有数据卷及数据，采用下列方式：

1. 清理有状态数据应用

   1. 删除应用

   2. 删除数据卷 PVCs

      相关的 PVs，LVs，LVRs，LVGs 都将被删除.

2. 清理 HwameiStor 系统组件

   1. 删除 HwameiStor 组件

      ```console
      $ kubectl delete cluster cluster-sample
      ```
      
   2. 删除 HwameiStor 系统空间

      ```console
      kubectl delete ns hwameistor
      ```

   3. 删除 CRD, Hook, 以及 RBAC

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

   4. 删除 StorageClass

      ```bash
      kubectl get sc -o name \
        | grep hwameistor-storage-lvm- \
        | xargs -t kubectl delete
      ```

最后，你仍然需要清理每个节点上的 LVM 配置，并采用额外的系统工具（例如：wipefs）清除磁盘上的所有数据
