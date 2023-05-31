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
   
2. 部署hwameistor-operator

   注意：如果没有可用的干净磁盘，operator就不会自动创建storageclass。operator会在安装过程中自动纳管磁盘，可用的磁盘会被添加到localstorage的pool里。如果可用盘是在安装后提供的，则需要手动下发localdiskclaim将磁盘纳管到localstoragenode里。一旦localstoragenode的pool里有磁盘，operator就会自动创建storageclass，也就是说，如果没有容量，就不会自动创建storageclass。
  
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
  您也可以在安装后通过修改deployment/apiserver来开启验证。


- 使用国内源:
  ```console
  helm install hwameistor-operator hwameistor-operator/hwameistor-operator \
  --set global.hwameistorImageRegistry=ghcr.m.daocloud.io \
  --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
  ```
