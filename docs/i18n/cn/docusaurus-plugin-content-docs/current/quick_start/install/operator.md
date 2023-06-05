---
sidebar_position: 2
sidebar_label: "HwameiStor-Operator 安装"
---

# 通过 HwameiStor-Operator 安装

Hwameistor-Operator 负责自动化安装并管理 HwameiStor 系统。

- HwameiStor 组件的全生命周期管理 (LCM)：
    - LocalDiskManager
    - LocalStorage
    - Scheduler
    - AdmissionController
    - VolumeEvictor
    - Exporter
    - Apiserver
    - Graph UI

- 根据不同目的和用途配置节点磁盘
- 自动发现节点磁盘的类型，并以此自动创建 HwameiStor 存储池
- 根据 HwameiStor 系统的配置和功能自动创建相应的 StorageClass

## 安装步骤

1. 添加 hwameistor-operator Helm Repo

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   ```

2. 部署 hwameistor-operator

   :::note
   如果没有可用的干净磁盘，Operator 就不会自动创建 StorageClass。
   Operator 会在安装过程中自动纳管磁盘，可用的磁盘会被添加到 LocalStorage 的 pool 里。
   如果可用磁盘是在安装后提供的，则需要手动下发 LocalDiskClaim 将磁盘纳管到 LocalStorageNode 里。
   一旦 LocalStorageNode 的 pool 里有磁盘，Operator 就会自动创建 StorageClass。
   也就是说，如果没有容量，就不会自动创建 StorageClass。
   :::
  
   ```console
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator
   ```
  
可选参数:

- 开启验证:

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator \
  --set apiserver.authentication.enable=true \
  --set apiserver.authentication.accessId={用户名} \
  --set apiserver.authentication.secretKey={密码}
  ```

  您也可以在安装后通过修改 deployment/apiserver 来开启验证。

- 使用国内源:

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator \
  --set global.hwameistorImageRegistry=ghcr.m.daocloud.io \
  --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
  ```
