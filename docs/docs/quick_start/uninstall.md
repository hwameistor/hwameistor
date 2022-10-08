---
sidebar_position: 4
sidebar_label: "Uninstall"
---

# Uninstall

:::danger
Before uninstalling HwameiStor, please make sure you have backed up all the data.
:::

## Delete helm instance

```console
$ helm delete -n hwameistor hwameistor
```

## Cleanup

### 1. Remove namespace

```console
$ kubectl delete ns hwameistor
```

### 2. Remove `LocalVolumeGroup` instances
   
:::note
   The `LocalVolumeGroup` object has a special finalizer, so its instances must be deleted before its definition is deleted.
:::

```console
$ kubectl delete localvolumegroups.hwameistor.io --all
```

### 3. Remove CRD, Hook, and RBAC

```console
$ kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
      | grep hwameistor \
      | xargs -t kubectl delete
```

### 4. Remove StorageClass

```console
$ kubectl get sc -o name \
      | grep hwameistor-storage-lvm- \
      | xargs -t kubectl delete
```
