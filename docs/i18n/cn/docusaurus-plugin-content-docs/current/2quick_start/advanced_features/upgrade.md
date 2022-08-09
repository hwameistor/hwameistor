---
sidebar_position: 3
sidebar_label: "升级"
---

# 升级

Helm 让 HwameiStor 的升级变得非常简单。只需运行以下命令：

```bash
helm upgrade \
  --namespace hwameistor \
  hwameistor \
  -f new.values.yaml
```

升级过程中将以滚动的方式重启每个 HwameiStor Pod。

:::caution
在升级 HwameiStor 期间，这些卷将继续不间断地为 Pod 服务。
:::

### 移除 CRD

运行以下命令移除 CRD。

```bash
kubectl get crd -o name \
  | grep hwameistor \
  | xargs -t kubectl delete
```

### 移除 clusterroles 和 rolebindings

运行以下命令执行移除操作。

```bash
kubectl get clusterrolebinding,clusterrole -o name \
  | grep hwameistor \
  | xargs -t kubectl delete
```
