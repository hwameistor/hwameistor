---
sidebar_position: 2
sidebar_label: "安装 HwameiStor"
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

2. 通过 hwameistor-operator 部署 HwameiStor

   :::note
   如果没有可用的干净磁盘，Operator 就不会自动创建 StorageClass。
   Operator 会在安装过程中自动纳管磁盘，可用的磁盘会被添加到 LocalStorage 的 pool 里。
   如果可用磁盘是在安装后提供的，则需要手动下发 LocalDiskClaim 将磁盘纳管到 LocalStorageNode 里。
   一旦 LocalStorageNode 的 pool 里有磁盘，Operator 就会自动创建 StorageClass。
   也就是说，如果没有容量，就不会自动创建 StorageClass。
   :::

   ```console
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace
   ```

可选参数：

- 磁盘预留

  可用的干净磁盘默认会被纳管并且添加到 LocalStorageNode 的 pool 里。
  如果您想在安装前预留一部分磁盘留作他用，您可以通过 helm 的 values 来设置磁盘预留配置。

  方式 1：

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace \
  --set diskReserve\[0\].nodeName=node1 \
  --set diskReserve\[0\].devices={/dev/sdc\,/dev/sdd} \
  --set diskReserve\[1\].nodeName=node2 \
  --set diskReserve\[1\].devices={/dev/sdc\,/dev/sde}
  ```

  这个例子展示了在 helm install 时通过--set 选项来设置磁盘预留配置，可能比较棘手。我们更建议把磁盘预留的配置写到一个文件里。

  方式 2：

  ```console
  diskReserve:
  - nodeName: node1
    devices:
    - /dev/sdc
    - /dev/sdd
  - nodeName: node2
    devices:
    - /dev/sdc
    - /dev/sde
  ```

  比如，您可以把如上的 helm values 写到一个叫 diskReserve.yaml 的文件里，并在 helm install 时 apply。

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace -f diskReserve.yaml
  ```

- 开启验证：

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace \
  --set apiserver.authentication.enable=true \
  --set apiserver.authentication.accessId={用户名} \
  --set apiserver.authentication.secretKey={密码}
  ```

  您也可以在安装后通过修改 deployment/apiserver 来开启验证。

- 使用国内源：

  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace \
  --set global.hwameistorImageRegistry=ghcr.m.daocloud.io \
  --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
  ```
