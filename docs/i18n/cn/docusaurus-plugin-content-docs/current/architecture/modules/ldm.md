---
sidebar_position: 1
sidebar_label: "本地磁盘管理器"
---

# 本地磁盘管理器

本地磁盘管理器 (Local Disk Manager, LDM) 是 HwameiStor 系统的一个重要功能模块。
`LDM` 旨在简化管理节点上的磁盘。它将磁盘抽象成一种可以被管理和监控的资源。
它本身是一种 DaemonSet 对象，集群中每一个节点都会运行该服务，通过该服务检测存在的磁盘并将其转换成相应的 LocalDisk 资源。

![LDM 架构图](../../img/ldm.png)

目前 LDM 还处于 `alpha` 阶段。

## 基本概念

**LocalDisk (LD)**: 这是 LDM 抽象的磁盘资源，一个 `LD` 代表了节点上的一块物理磁盘。

**LocalDiskClaim (LDC)**: 这是系统使用磁盘的方式，通过创建 `LDC` 对象来向系统申请磁盘。用户可以添加一些对磁盘的描述来选择磁盘。

> 目前，LDC 支持以下对磁盘的描述选项：
>
> - NodeName
> - Capacity
> - DiskType(e.g. HDD/SSD)

## 用法

如果想完整地部署 HwameiStor，请参考[使用 Helm Chart 安装部署](../../quick_start/install/deploy.md)。

如果只想单独部署 LDM，可以参考下面的步骤进行安装。

## 安装本地磁盘管理器

1. 克隆  repo 到本机

    ```bash
    $ git clone https://github.com/hwameistor/local-disk-manager.git
    ```

2. 进入 repo 对应的目录

    ```bash
    $ cd deploy
    ```

3. 安装 CRDs 和 运行 LocalDiskManager

    1. 安装 LocalDisk 和 LocalDiskClaim 的 CRD

        ```bash
        $ kubectl apply -f deploy/crds/
        ```

    2. 安装权限认证的 CR 以及 LDM 的 Operator

        ```bash
        $ kubectl apply -f deploy/
        ```

4. 查看 LocalDisk 信息

    ```bash
    $ kubectl get localdisk
    10-6-118-11-sda    10-6-118-11                             Unclaimed
    10-6-118-11-sdb    10-6-118-11                             Unclaimed
    ```

    该命令用于获取集群中磁盘资源信息，获取的信息总共有四列，含义分别如下：

    - **NAME：** 代表磁盘在集群中的名称。
    - **NODEMATCH：** 表明磁盘所在的节点名称。
    - **CLAIM：** 表明这个磁盘是被哪个 `Claim` 所引用。
    - **PHASE：** 表明这个磁盘当前的状态。

    通过`kuebctl get localdisk <name> -o yaml` 查看更多关于某块磁盘的信息。

5. 申请可用磁盘

    1. 创建 LocalDiskClaim

        ```bash
        $ kubectl apply -f deploy/samples/hwameistor.io_v1alpha1_localdiskclaim_cr.yaml
        ```

        该命令用于创建一个磁盘使用的申请请求。在这个 yaml 文件里面，您可以在 description 字段添加对申请磁盘的描述，比如磁盘类型、磁盘的容量等等。

    2. 查看 LocalDiskClaim 信息

        ```bash
        $ kubectl get localdiskclaim <name>
        ```

查看 `Claim` 的 Status 字段信息。如果存在可用的磁盘，您将会看到该字段的值为 `Bound`。

## 路线规划图

| 功能                   | 状态   | 版本 | 描述                                     |
| :--------------------- | ------ | ---- | ---------------------------------------- |
| CSI for disk volume    | Planned |      | `Disk` 模式下创建数据卷的 `CSI` 接口     |
| Disk management        | Planned |      | 磁盘管理、磁盘分配、磁盘事件感知处理     |
| Disk health management | Planned |      | 磁盘健康管理，包括故障预测和状态上报等等 |
| HA disk Volume         | Planned |      | Disk 数据卷的高可用                      |

## 反馈

如果您有任何的疑问和建议，请反馈至 [Issues](https://github.com/hwameistor/local-disk-manager/issues)