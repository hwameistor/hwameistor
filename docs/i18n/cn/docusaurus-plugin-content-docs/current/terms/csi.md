---
sidebar_position: 3
sidebar_label: "CSI 接口"
---

# CSI 接口

CSI 即 Container Storage Interfaces，容器存储接口。目前，Kubernetes 中的存储子系统仍存在一些问题。
存储驱动程序代码在 Kubernetes 核心存储库中进行维护，这很难测试。Kubernetes 还需要授予存储供应商许可，便于将代码嵌入 Kubernetes 核心存储库。

CSI 旨在定义行业标准，该标准将使支持 CSI 的存储提供商能够在支持 CSI 的容器编排系统中使用。

下图描述了一种与 CSI 集成的高级 Kubernetes 原型。

![CSI 接口](../img/csi.png)

- 引入了三个新的外部组件以解耦 Kubernetes 和存储提供程序逻辑
- 蓝色箭头表示针对 API 服务器进行调用的常规方法
- 红色箭头显示 gRPC 以针对 Volume Driver 进行调用

## 扩展 CSI 和 Kubernetes

为了实现在 Kubernetes 上扩展卷的功能，应该扩展几个组件，包括 CSI 规范、“in-tree” 卷插件、external-provisioner 和 external-attacher。

## 扩展 CSI 规范

最新的 CSI 0.2.0 仍未定义扩展卷的功能。应引入新的 3 个 RPC：`RequiresFSResize`、`ControllerResizeVolume` 和 `NodeResizeVolume`。

```jade
service Controller {
  rpc CreateVolume (CreateVolumeRequest)
    returns (CreateVolumeResponse) {}
……
  rpc RequiresFSResize (RequiresFSResizeRequest)
    returns (RequiresFSResizeResponse) {}
  rpc ControllerResizeVolume (ControllerResizeVolumeRequest)
    returns (ControllerResizeVolumeResponse) {}
}
service Node {
  rpc NodeStageVolume (NodeStageVolumeRequest)
    returns (NodeStageVolumeResponse) {}
……
  rpc NodeResizeVolume (NodeResizeVolumeRequest)
    returns (NodeResizeVolumeResponse) {}
}
```

## 扩展 “In-Tree” 卷插件

除了扩展的 CSI 规范之外，Kubernetes 中的 `csiPlugin` 接口还应实现 `expandablePlugin`。
`csiPlugin` 接口将扩展代表 `ExpanderController` 的 `PersistentVolumeClaim`。

```jade
type ExpandableVolumePlugin interface {
VolumePlugin
ExpandVolumeDevice(spec Spec, newSize resource.Quantity, oldSize resource.Quantity) (resource.Quantity, error)
RequiresFSResize() bool
}
```

### 实现卷驱动程序

最后，为了抽象化实现的复杂性，应将单独的存储提供程序管理逻辑硬编码为以下功能，这些功能在 CSI 规范中已明确定义：

- CreateVolume
- DeleteVolume
- ControllerPublishVolume
- ControllerUnpublishVolume
- ValidateVolumeCapabilities
- ListVolumes
- GetCapacity
- ControllerGetCapabilities
- RequiresFSResize
- ControllerResizeVolume

## 展示

以具体的用户案例来演示此功能。

- 为 CSI 存储供应商创建存储类

  ```yaml
  allowVolumeExpansion: true
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: csi-qcfs
  parameters:
    csiProvisionerSecretName: orain-test
    csiProvisionerSecretNamespace: default
  provisioner: csi-qcfsplugin
  reclaimPolicy: Delete
  volumeBindingMode: Immediate
  ```

- 在 Kubernetes 集群上部署包括存储供应商 `csi-qcfsplugin` 在内的 CSI 卷驱动
- 创建 PVC `qcfs-pvc`，它将由存储类 `csi-qcfs` 动态配置

  ```yaml
  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
    name: qcfs-pvc
    namespace: default
  ....
  spec:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 300Gi
    storageClassName: csi-qcfs
  ```

- 创建 MySQL 5.7 实例以使用 PVC `qcfs-pvc`
- 为了反映完全相同的生产级别方案，实际上有两种不同类型的工作负载，包括：
  - 批量插入使 MySQL 消耗更多的文件系统容量
  - 浪涌查询请求
- 通过编辑 pvc `qcfs-pvc` 配置动态扩展卷容量
