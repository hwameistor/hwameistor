---
sidebar_position: 7
sidebar_label: "PV 和 PVC"
---

# PV 和 PVC

PV（PersistentVolume，持久卷）是对存储资源的抽象，将存储定义为一种容器应用可以使用的资源。
PV 由管理员创建和配置，与存储提供商的具体实现直接相关，例如文件存储、块存储、对象存储或 DRBD 等，
通过插件式的机制进行管理，供应用访问和使用。除 EmptyDir 类型的存储卷，PV 的生命周期独立于使用它的 Pod。

PVC（PersistentVolumeClaim，持久卷声明）是用户对存储资源的一个申请。就像 Pod 消耗 Node 的资源一样，PVC 会消耗 PV 的资源。
PVC 可以申请存储空间的大小 (Size) 和访问模式（例如 ReadWriteOnce、ReadOnlyMany 或 ReadWriteMany）。

使用 PVC 申请的存储空间仍然不满足应用对存储设备的各种需求。在很多情况下，应用程序对存储设备的特性和性能都有不同的要求，
包括读写速度、并发性能、数据冗余等要求。这就需要使用资源对象 StorageClass，用于标记存储资源和性能，根据 PVC 的需求动态供给合适的 PV 资源。
StorageClass 和存储资源动态供应的机制经完善后，实现了存储卷的按需创建，在共享存储的自动化管理进程中实现了重要的一步。

另请参考 Kubernetes 官方文档：

- [持久卷](https://kubernetes.io/zh-cn/docs/concepts/storage/persistent-volumes/)
- [存储类](https://kubernetes.io/zh-cn/docs/concepts/storage/storage-classes/)
- [动态卷供应](https://kubernetes.io/zh-cn/docs/concepts/storage/dynamic-provisioning/)