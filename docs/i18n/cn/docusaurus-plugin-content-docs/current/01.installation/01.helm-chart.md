---
sidebar_position: 2
sidebar_label: "通过 Helm Chart 部署"
---

# 通过 Helm Chart 部署

本地磁盘是 HwameiStor 的一部分，必须与本地磁盘管理器一起工作，建议用户通过 helm-charts 进行部署。

目前此功能为 alpha 版本，如有变更，恕不另行通知。所有代码按原样提供，不作任何保证。Alpha 功能不受正式发行版本功能的 SLA 要求约束。


## 第 1 步：安装 HwameiStor

必须先安装 [Helm](https://helm.sh/) 才能使用 chart，请参阅 Helm [官方文档](https://helm.sh/docs/)。

```bash
$ git clone https://github.com/hwameistor/helm-charts.git 
$ cd helm-charts/charts 
$ helm install hwameistor -n hwameistor --create-namespace --generate-name
```

或

```bash
$ helm repo add hwameistor http://hwameistor.io/helm-charts 
$ helm install hwameistor/hwameistor -n hwameistor --create-namespace --generate-name
```

然后运行 `helm search repo hwameistor` 来查看 chart。

## 第 2 步：在节点上启用 HwameiStor

Helm chart 安装完成后，在对应节点运行以下命令来启用 HwameiStor：

```bash
$ kubectl label node <node-name> "lvm.hwameistor.io/enable=true"
```

## 第 3 步：在节点上按类型 Claim 磁盘

通过申请 LocalDiskClaim 自定义资源，为本地磁盘申请磁盘：

```bash
cat > ./local-disk-claim.yaml <<- EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: <anyname>
  namespace: hwameistor
spec:
  nodeName: <node-name>
  description:
    diskType: <HDD or SSD or NVMe>
EOF
```

恭喜！现在 HwameiStor 已部署到您的集群上。
