---
sidebar_position: 4
sidebar_label: "卸载"
---

# 卸载

:::danger
Before uninstalling HwameiStor, make sure you have backed up all the data.
:::

## Step 1: Delete helm instance

```bash
$ helm delete \
    --namespace hwameistor \
    hwameistor
```

## Step 2: Cleanup

### Remove namespace

```bash
$ kubectl delete ns hwameistor
```

### Remove CRDs

```bash
$ kubectl get crd -o name \
    | grep hwameistor \
    | xargs -t kubectl delete
```

### Remove clusterRoles and roleBindings

```bash
$ kubectl get clusterrolebinding,clusterrole -o name \
    | grep hwameistor \
    | xargs -t kubectl delete
```

### Remove storageClass

```bash
$ kubectl get sc -o name \
    | grep hwameistor-storage-lvm- \
    | xargs -t kubectl delete
```