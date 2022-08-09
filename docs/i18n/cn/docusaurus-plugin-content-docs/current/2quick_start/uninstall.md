---
sidebar_position: 4
sidebar_label: "卸载"
---

# 卸载

:::danger
务必先备份好所有数据，再卸载 HwameiStor。
:::

## 删除 Helm 实例

```bash
helm delete \
  --namespace hwameistor \
  hwameistor
```

## 清理工作

1. 移除命名空间。

   ```bash
   kubectl delete ns hwameistor
   ```

2. 移除 CRD。

   ```bash
   kubectl get crd -o name \
     | grep hwameistor \
     | xargs -t kubectl delete
   ```

3. 移除 clusterroles 和 rolebindings。

   ```bash
   kubectl get clusterrolebinding,clusterrole -o name \
     | grep hwameistor \
     | xargs -t kubectl delete
   ```

4. 移除 storageClass

   ```bash
   kubectl get sc -o name \
     | grep hwameistor-storage-lvm- \
     | xargs -t kubectl delete
   ```
