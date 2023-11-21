---
sidebar_position: 1
sidebar_label: "准备工作"
---

# 准备工作

## Kubernetes 平台

1. Kubernetes `1.18+`
2. 部署 CoreDNS

### 不支持的平台

1. Openshift
2. Rancher

:::note
暂时不支持以上平台，但是计划未来支持。
:::

## 主机配置

### Linux 发行版

1. CentOS/RHEL `7.4+`
2. Rocky Linux `8.4+`
3. Ubuntu `18+`
4. Kylin 麒麟`V10`

### 处理器架构

1. x86_64
2. ARM64

### 软件依赖

1. 安装 `LVM2`
2. 高可用功能需要安装和当前运行的 kernel 版本一致的 `kernel-devel`
3. 数据卷扩容功能需要安装文件系统大小调整工具。使用 `xfs` 作为默认文件系统。因此节点上面需要安装 `xfs_growfs`

```console title="CentOS/RHEL, Rocky 和 Kylin"
$ yum install -y lvm2
$ yum install -y kernel-devel-$(uname -r)
$ yum install -y xfsprogs
```

```console title="Ubuntu"
$ apt-get install -y lvm2
$ apt-get install -y linux-headers-$(uname -r)
$ apt-get install -y xfsprogs
```

### Secure Boot

高可用功能暂时不支持 `Secure Boot`，确认 `Secure Boot` 是 `disabled` 状态：

```console
$ mokutil --sb-state
SecureBoot disabled

$ dmesg | grep secureboot
[    0.000000] secureboot: Secure boot disabled
```

### 数据盘

HwameiStor 支持物理硬盘 (HDD)、固态硬盘 (SSD) 和 NVMe 闪存盘.

测试环境里，每个主机必须要有至少一块空闲的 `10GiB` 数据盘。

生产环境里，建议每个主机至少要有一块空闲的 `200GiB` 数据盘，而且建议使用固态硬盘 (SSD)。

:::note
对于虚拟机环境，请确保每台虚拟机已经启用磁盘序列号的功能，这会帮助 HwameiStor 更好的识别管理主机上的磁盘。

为了避免磁盘识别冲突，请确保提供的虚拟磁盘序列号不能重复。
:::

### 网络

生产环境里，开启高可用模式后，建议使用有冗余保护的`万兆 TCP/IP` 网络。
