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
1. Rancher

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
1. ARM64

### 软件依赖

1. 安装 `LVM2`
2. 高可用功能需要安装和当前运行的 kernel 版本一致的 `kernel-devel`

```console title="CentOS/RHEL, Rocky 和 Kylin"
$ yum install -y lvm2
$ yum install -y kernel-devel-$(uname -r)
```

```console title="Ubuntu"
$ apt-get install -y lvm2
$ apt-get install -y linux-headers-$(uname -r)
```

### 数据盘

HwameiStor 支持物理硬盘(HDD)、固态硬盘(SSD) 和 NVMe 闪存盘.

测试环境里，每个主机必须要有至少一块空闲的 `10GiB` 数据盘。

生产环境里，建议每个主机至少要有一块空闲的 `200GiB` 数据盘，而且建议使用固态硬盘 (SSD)。

### 网络

生产环境里，开启高可用模式后，建议使用有冗余保护的`万兆 TCP/IP` 网络。
