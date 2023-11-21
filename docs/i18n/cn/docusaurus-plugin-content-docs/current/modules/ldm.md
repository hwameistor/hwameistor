---
sidebar_position: 1
sidebar_label: "本地磁盘管理器"
---

# 本地磁盘管理器

本地磁盘管理器 (Local Disk Manager, LDM) 是 HwameiStor 系统的一个重要功能模块。
`LDM` 旨在简化管理节点上的磁盘。它将磁盘抽象成一种可以被管理和监控的资源。
它本身是一种 DaemonSet 对象，集群中每一个节点都会运行该服务，通过该服务检测存在的磁盘并将其转换成相应的 LocalDisk 资源。

![LDM 架构图](../img/ldm.png)

目前 LDM 还处于 `alpha` 阶段。

## 基本概念

**LocalDisk (LD)**: 这是 LDM 抽象的磁盘资源，一个 `LD` 代表了节点上的一块物理磁盘。

**LocalDiskClaim (LDC)**: 这是系统使用磁盘的方式，通过创建 `LDC` 对象来向系统申请磁盘。用户可以添加一些对磁盘的描述来选择磁盘。

> 目前，LDC 支持以下对磁盘的描述选项：
>
> - NodeName
> - Capacity
> - DiskType（例如 HDD/SSD）

## 用法

1. 查看 LocalDisk 信息

    ```bash
    $ kubectl get localdisk
    NAME               NODEMATCH        PHASE
    10-6-118-11-sda    10-6-118-11      Available
    10-6-118-11-sdb    10-6-118-11      Available
    ```

    该命令用于获取集群中磁盘资源信息，获取的信息总共有三列，含义分别如下：

    - **NAME：** 代表磁盘在集群中的名称。
    - **NODEMATCH：** 表明磁盘所在的节点名称。
    - **PHASE：** 表明这个磁盘当前的状态。

    通过 `kubectl get localdisk <name> -o yaml` 查看更多关于某块磁盘的信息。

2. 申请可用磁盘

    1. 创建 LocalDiskClaim

        ```bash
        cat << EOF | kubectl apply -f -
        apiVersion: hwameistor.io/v1alpha1
        kind: LocalDiskClaim
        metadata:
          name: <localDiskClaimName>
        spec:
          description:
            # 比如：HDD,SSD,NVMe
            diskType: <diskType>
          # 磁盘所在节点
          nodeName: <nodeName>
          # 使用磁盘的系统名称 比如：local-storage,local-disk-manager
          owner: <ownerName>
        EOF
        ```

        该命令用于创建一个磁盘使用的申请请求。在这个 yaml 文件里面，您可以在 description 字段添加对申请磁盘的描述，比如磁盘类型、磁盘的容量等等。

    2. 查看 LocalDiskClaim 信息

        ```bash
        $ kubectl get localdiskclaim <name>
        ```

    3. LDC 被处理完成后，将立即被系统自动清理。如果 `owner` 是 local-storage，处理后的结果可以在对应的 `LocalStorageNode` 里查看。
