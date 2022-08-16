---
sidebar_position: 4
sidebar_label: "卸载"
---

# 卸载

:::danger
请务必先备份好所有数据，再卸载 HwameiStor。
:::

## 删除 Helm 实例

```console
$ helm delete -n hwameistor hwameistor
```

## 清理工作

### 1. 移除命名空间

```console
$ kubectl delete ns hwameistor
```

### 2. 删除 `LocalVolumeGroup` 实例
   
:::note
   `LocalVolumeGroup` 对象有特殊的终结器（finalizer），所以必须先删除它的实例再删除它的定义。
:::

```console
$ kubectl delete localvolumegroups.hwameistor.io --all
```

### 3. 移除 CRD、Hook 和 RBAC

```console
$ kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
      | grep hwameistor \
      | xargs -t kubectl delete
```

### 4. 移除 StorageClass

```console
$ kubectl get sc -o name \
   | grep hwameistor-storage-lvm- \
   | xargs -t kubectl delete
```
