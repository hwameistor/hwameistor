---
sidebar_position: 2
sidebar_label: "通过 Helm Chart 安装"
---

# 通过 Helm Chart 安装

推荐使用这种安装方式，HwameiStor 的任何组件都可以通过 Helm Charts 轻松安装。

## 步骤

### 1. 准备 Helm 工具

安装 [Helm](https://helm.sh/) 命令行工具，请参阅 [Helm 文档](https://helm.sh/docs/)。

### 2. 下载 `hwameistor` Repo

下载并解压 Repo 文件到本地

```console
$ helm repo add hwameistor http://hwameistor.io/hwameistor

$ helm repo update hwameistor

$ helm pull hwameistor/hwameistor --untar
```

### 3. 安装 HwameiStor

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace
```

**安装完成!**

要验证安装效果，请参见下一章[安装后检查](./post_check.md)。

## 使用镜像仓库镜像

:::tip

默认的镜像仓库是 `quay.io` 和 `ghcr.io`。
如果无法访问，可尝试使用 DaoCloud 提供的镜像源：`quay.m.daocloud.io` 和 `ghcr.m.daocloud.io`。

:::

要切换镜像仓库的镜像，请使用 `--set` 更改这两个参数值：`global.k8sImageRegistry` 和 `global.hwameistorImageRegistry`。

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    --set global.k8sImageRegistry=k8s-gcr.m.daocloud.io \
    --set global.hwameistorImageRegistry=ghcr.m.daocloud.io
```

## 自定义 kubelet 根目录

:::caution

默认的 `kubelet` 目录为 `/var/lib/kubelet`。
如果您的 Kubernetes 发行版使用不同的 `kubelet` 目录，必须设置参数 `kubeletRootDir`。

:::

例如，在将 `/var/snap/microk8s/common/var/lib/kubelet/` 用作 `kubelet` 目录的 [Canonical 的 MicroK8s](https://microk8s.io/) 上，HwameiStor 需要按以下方式安装：
 
```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    --set kubeletRootDir=/var/snap/microk8s/common/var/lib/kubelet/
```

## 生产环境安装

生产环境需要：

- 指定资源配置
- 避免部署到 Master 节点
- 实现控制器的快速故障切换
  
`values.extra.prod.yaml` 文件中提供了一些推荐值，具体用法为：

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    -f ./hwameistor/values.yaml \
    -f ./hwameistor/values.extra.prod.yaml
```

:::caution

在资源紧张的测试环境中，设置上述数值会造成 Pod 无法启动！

:::

## [可选] 安装 DRBD

如果要启用高可用卷, 必须安装 DRBD:

:::caution
请注意 [**准备工作**](./prereq.md) 里的环境要求
:::

```console
$ helm pull hwameistor/drbd-adapter --untar

$ helm install drbd-adapter ./drbd-adapter \
    -n hwameistor --create-namespace
```

中国用户可以使用镜像仓库 `daocloud.io/daocloud` 加速

```console
$ helm install drbd-adapter ./drbd-adapter \
    -n hwameistor --create-namespace \
    --set registry=daocloud.io/daocloud
```
